package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/cors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"

	"hasir-api/internal"
	internalOrganization "hasir-api/internal/organization"
	"hasir-api/internal/registry"
	"hasir-api/internal/user"
	"hasir-api/pkg/auth"
	"hasir-api/pkg/config"
	"hasir-api/pkg/email"
	_ "hasir-api/pkg/log"
	"hasir-api/pkg/organization"
)

func main() {
	cfgReader := config.NewConfigReader()
	cfg := cfgReader.Read()

	zap.L().Info("Server starting...")

	m, err := migrate.New(
		"file://migrations",
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.PostgresConfig.Username,
			cfg.PostgresConfig.Password,
			cfg.PostgresConfig.Host,
			cfg.PostgresConfig.Port,
			cfg.PostgresConfig.Database,
		),
	)
	if err != nil {
		zap.L().Fatal("failed to create migration client", zap.Error(err))
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		zap.L().Fatal("failed to apply migrations", zap.Error(err))
	}

	var traceProvider *sdktrace.TracerProvider
	if cfg.Otel.Enabled {
		traceProvider = initTracer(cfg)
	}

	userPgRepository := user.NewPgRepository(cfg, traceProvider)
	repositoryPgRepository := registry.NewPgRepository(cfg, traceProvider)
	organizationPgRepository := internalOrganization.NewPgRepository(cfg, traceProvider)

	emailService := email.NewService(cfg)

	ctx := context.Background()
	organizationPgRepository.StartEmailJobProcessor(ctx, emailService, 10, 5*time.Second)

	orgRepoAdapter := organization.NewOrgRepositoryAdapter(organizationPgRepository)

	userService := user.NewService(cfg, userPgRepository)
	registryService := registry.NewService(repositoryPgRepository, orgRepoAdapter)
	organizationService := internalOrganization.NewService(
		organizationPgRepository,
		registryService,
		emailService,
		userPgRepository,
	)

	authInterceptor := auth.NewAuthInterceptor(cfg.JwtSecret)

	interceptors := []connect.Interceptor{validate.NewInterceptor(), authInterceptor}
	if cfg.Otel.Enabled {
		otelInterceptor, err := otelconnect.NewInterceptor(
			otelconnect.WithTracerProvider(traceProvider),
		)
		if err != nil {
			zap.L().Fatal("failed to create connect opentelemetry interceptor", zap.Error(err))
		}
		interceptors = append(interceptors, otelInterceptor)
	}

	userHandler := user.NewHandler(userService, userPgRepository, interceptors...)
	registryHandler := registry.NewHandler(registryService, repositoryPgRepository, interceptors...)
	organizationHandler := internalOrganization.NewHandler(organizationService, organizationPgRepository, interceptors...)
	handlers := []internal.GlobalHandler{
		userHandler,
		registryHandler,
		organizationHandler,
	}

	mux := http.NewServeMux()
	handler := cors.AllowAll().Handler(mux)
	for _, handler := range handlers {
		path, h := handler.RegisterRoutes()
		mux.Handle(path, h)
	}

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	server := &http.Server{
		Addr:      cfg.Server.GetServerAddress(),
		Handler:   handler,
		Protocols: protocols,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("HTTP server error", zap.Error(err))
		}
	}()
	zap.L().Info("Server started on port", zap.String("port", cfg.Server.Port))

	gracefulShutdown(server, traceProvider, organizationPgRepository)
}

func gracefulShutdown(server *http.Server, traceProvider *sdktrace.TracerProvider, organizationRepo *internalOrganization.PgRepository) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	zap.L().Info("Shutting down server...")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if organizationRepo != nil {
		organizationRepo.StopEmailJobProcessor()
	}

	if err := server.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("HTTP shutdown error", zap.Error(err))
	}

	if traceProvider != nil {
		if err := traceProvider.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("TracerProvider shutdown error", zap.Error(err))
		}
	}

	zap.L().Info("Server gracefully stopped")
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
		otlptracegrpc.WithEndpoint(cfg.Otel.TraceEndpoint),
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

	zap.L().Info("OpenTelemetry tracing enabled", zap.String("endpoint", cfg.Otel.TraceEndpoint))
	return tracerProvider
}
