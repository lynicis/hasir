package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"

	"apps/api/internal"
	"apps/api/internal/user"
	"apps/api/pkg/config"
	_ "apps/api/pkg/log"
)

func main() {
	cfgReader := config.NewConfigReader()
	cfg := cfgReader.Read()

	traceProvider := initTracer(cfg)

	userPgRepository := user.NewPgRepository(cfg, traceProvider)

	userService := user.NewService(cfg, userPgRepository)

	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(traceProvider),
	)
	if err != nil {
		zap.L().Fatal("failed to create connect opentelementry interceptor", zap.Error(err))
	}

	userHandler := user.NewHandler(validateInterceptor, otelInterceptor, userService)
	handlers := []internal.GlobalHandler{userHandler}

	mux := http.NewServeMux()
	for _, handler := range handlers {
		path, h := handler.RegisterRoutes()
		mux.Handle(path, h)
	}

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	server := &http.Server{
		Addr:      cfg.Server.GetServerAddress(),
		Handler:   mux,
		Protocols: protocols,
	}

	go func() {
		if err = server.ListenAndServe(); err != nil {
			zap.L().Fatal("HTTP server error: %v", zap.Error(err))
		}
	}()

	gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		zap.L().Fatal("HTTP shutdown error: %v", zap.Error(err))
	}
}

func initTracer(cfg *config.Config) *sdktrace.TracerProvider {
	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("hasir-api"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		zap.L().Fatal("failed to create resource", zap.Error(err))
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OtelTraceEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		zap.L().Fatal("failed to create trace exporter", zap.Error(err))
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tracerProvider
}
