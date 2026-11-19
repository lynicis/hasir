package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/gliderlabs/ssh"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"

	"hasir/api/internal"
	internalOrganization "hasir/api/internal/organization"
	"hasir/api/internal/registry"
	"hasir/api/internal/user"
	"hasir/api/pkg/authentication"
	"hasir/api/pkg/authorization"
	"hasir/api/pkg/config"
	"hasir/api/pkg/email"
	"hasir/api/pkg/idempotency"
	_ "hasir/api/pkg/log"
	"hasir/api/pkg/csrf"
	"hasir/api/pkg/postgres"
	postgresIdempotency "hasir/api/pkg/postgres/idempotency"
	postgresOrganization "hasir/api/pkg/postgres/organization"
	postgresRegistry "hasir/api/pkg/postgres/registry"
	postgresUser "hasir/api/pkg/postgres/user"
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

	dbPool := postgres.MustNewPool(cfg, traceProvider)

	idempotencyPgRepository := postgresIdempotency.NewIdempotencyRepositoryWithPool(dbPool, traceProvider)
	userPgRepository := postgresUser.NewPgRepositoryWithPool(dbPool, traceProvider)
	repositoryPgRepository := postgresRegistry.NewPgRepositoryWithPool(dbPool, traceProvider)
	organizationPgRepository := postgresOrganization.NewOrganizationRepositoryWithPool(dbPool, traceProvider)

	emailService := email.NewService(cfg)

	ctx := context.Background()
	emailJobQueue := postgresOrganization.NewEmailJobQueue(
		organizationPgRepository.GetConnectionPool(),
		organizationPgRepository.GetTracer(),
	)
	emailJobQueue.Start(ctx, emailService, 10, 5*time.Second)

	orgRepoAdapter := authorization.NewOrgRepositoryAdapter(organizationPgRepository)

	sdkGenerationQueue := postgresRegistry.NewSdkGenerationJobQueue(
		repositoryPgRepository.GetConnectionPool(),
		repositoryPgRepository.GetTracer(),
	)

	registryService := registry.NewService(repositoryPgRepository, orgRepoAdapter, sdkGenerationQueue, cfg)

	pollInterval, err := time.ParseDuration(cfg.SdkGeneration.PollInterval)
	if err != nil {
		zap.L().Fatal("invalid SDK generation poll interval", zap.Error(err))
	}

	sdkGenerationQueue.Start(ctx, registryService, registryService, cfg.SdkGeneration.WorkerCount, pollInterval)
	zap.L().Info(
		"SDK generation queue listening",
		zap.Int("workerCount", cfg.SdkGeneration.WorkerCount),
		zap.Duration("pollInterval", pollInterval),
	)

	userService := user.NewService(cfg, userPgRepository, emailService)
	organizationService := internalOrganization.NewService(
		organizationPgRepository,
		emailJobQueue,
		registryService,
		emailService,
		userPgRepository,
	)

	authInterceptor := authentication.NewAuthInterceptor(cfg.JwtSecret)
	idempotencyInterceptor := idempotency.NewIdempotencyInterceptor(idempotencyPgRepository)
	csrfInterceptor := csrf.NewCsrfInterceptor(cfg.DashboardUrl, cfg.Server.PublicUrl)

	interceptors := []connect.Interceptor{
		csrfInterceptor.Interceptor(),
		validate.NewInterceptor(),
		idempotencyInterceptor.Interceptor(),
		authInterceptor.Interceptor(),
	}
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
	organizationHandler := internalOrganization.NewHandler(organizationService, organizationPgRepository, repositoryPgRepository, interceptors...)
	handlers := []internal.GlobalHandler{
		userHandler,
		registryHandler,
		organizationHandler,
	}

	mux := http.NewServeMux()
	
	corsOptions := cors.Options{
		AllowedOrigins: []string{
			strings.TrimRight(cfg.DashboardUrl, "/"),
			strings.TrimRight(cfg.Server.PublicUrl, "/"),
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}
	handler := cors.New(corsOptions).Handler(mux)
	for _, handler := range handlers {
		path, h := handler.RegisterRoutes()
		mux.Handle(path, h)
	}

	gitHttpHandler := registry.NewGitHttpHandler(registryService, userPgRepository, registry.DefaultReposPath)
	mux.Handle("/git/", gitHttpHandler)

	sdkHttpHandler := registry.NewSdkHttpHandler(cfg.SdkGeneration.OutputPath)
	mux.Handle("/sdk/", sdkHttpHandler)

	sdkPath := cfg.SdkGeneration.GetOutputPath()
	docHttpHandler := registry.NewDocumentationHttpHandler(
		registryService,
		repositoryPgRepository,
		cfg.JwtSecret,
		sdkPath,
	)
	mux.Handle("/docs/", docHttpHandler)

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	server := &http.Server{
		Addr:              cfg.Server.GetServerAddress(),
		Handler:           handler,
		Protocols:         protocols,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("HTTP server error", zap.Error(err))
		}
	}()
	zap.L().Info("Server started on port", zap.String("port", cfg.Server.Port))

	var sshServer *ssh.Server
	if cfg.Ssh.Enabled {
		gitSshHandler := registry.NewGitSshHandler(registryService, registry.DefaultReposPath)
		sdkSshHandler := registry.NewSdkSshHandler(cfg.SdkGeneration.OutputPath)
		sshServer = registry.StartSshServer(cfg, userPgRepository, gitSshHandler, sdkSshHandler)
	}

	gracefulShutdown(server, sshServer, dbPool, traceProvider, emailJobQueue, sdkGenerationQueue)
}

func gracefulShutdown(server *http.Server, sshServer *ssh.Server, dbPool *pgxpool.Pool, traceProvider *sdktrace.TracerProvider, emailJobQueue internalOrganization.Queue, sdkGenerationQueue registry.SdkGenerationQueue) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	zap.L().Info("Shutting down server...")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	emailJobQueue.Stop()
	sdkGenerationQueue.Stop()

	if err := server.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("HTTP shutdown error", zap.Error(err))
	}

	if sshServer != nil {
		if err := sshServer.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("SSH shutdown error", zap.Error(err))
		}
	}

	if dbPool != nil {
		dbPool.Close()
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
