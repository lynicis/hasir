package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
	"github.com/rs/cors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"
	gossh "golang.org/x/crypto/ssh"

	"hasir-api/internal"
	internalOrganization "hasir-api/internal/organization"
	"hasir-api/internal/registry"
	"hasir-api/internal/user"
	"hasir-api/pkg/authentication"
	"hasir-api/pkg/authorization"
	"hasir-api/pkg/config"
	"hasir-api/pkg/email"
	_ "hasir-api/pkg/log"
	postgresOrganization "hasir-api/pkg/postgres/organization"
	postgresRegistry "hasir-api/pkg/postgres/registry"
	postgresUser "hasir-api/pkg/postgres/user"
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

	userPgRepository := postgresUser.NewPgRepository(cfg, traceProvider)
	repositoryPgRepository := postgresRegistry.NewPgRepository(cfg, traceProvider)
	organizationPgRepository := postgresOrganization.NewOrganizationRepository(cfg, traceProvider)

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
	organizationHandler := internalOrganization.NewHandler(organizationService, organizationPgRepository, repositoryPgRepository, interceptors...)
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
		sshServer = startSshServer(cfg, userPgRepository, gitSshHandler, sdkSshHandler)
	}

	gracefulShutdown(server, sshServer, traceProvider, emailJobQueue, sdkGenerationQueue)
}

func gracefulShutdown(server *http.Server, sshServer *ssh.Server, traceProvider *sdktrace.TracerProvider, emailJobQueue internalOrganization.Queue, sdkGenerationQueue registry.SdkGenerationQueue) {
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

func startSshServer(cfg *config.Config, userRepo user.Repository, gitSshHandler *registry.GitSshHandler, sdkSshHandler *registry.SdkSshHandler) *ssh.Server {
	hostKey, err := loadOrGenerateHostKey(cfg.Ssh.HostKeyPath)
	if err != nil {
		zap.L().Fatal("failed to load SSH host key", zap.Error(err))
	}

	sshServer := &ssh.Server{
		Addr: ":" + cfg.Ssh.Port,
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			publicKeyStr := strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key)))
			userDTO, err := userRepo.GetUserBySshPublicKey(context.Background(), publicKeyStr)
			if err != nil {
				zap.L().Debug("SSH auth failed", zap.Error(err))
				return false
			}
			ctx.SetValue("userId", userDTO.Id)
			zap.L().Info("SSH auth success", zap.String("userId", userDTO.Id))
			return true
		},
		Handler: func(session ssh.Session) {
			userId, ok := session.Context().Value("userId").(string)
			if !ok || userId == "" {
				_, _ = fmt.Fprintln(session.Stderr(), "Authentication required")
				_ = session.Exit(1)
				return
			}

			cmd := session.RawCommand()
			var handlerErr error
			if cmd != "" {
				parts := strings.SplitN(cmd, " ", 2)
				if len(parts) == 2 {
					repoPath := strings.Trim(parts[1], "'\"")
					repoPath = strings.TrimPrefix(repoPath, "/")
					if strings.HasPrefix(repoPath, "sdk/") {
						handlerErr = sdkSshHandler.HandleSession(session, userId)
					} else {
						handlerErr = gitSshHandler.HandleSession(session, userId)
					}
				} else {
					handlerErr = gitSshHandler.HandleSession(session, userId)
				}
			} else {
				handlerErr = gitSshHandler.HandleSession(session, userId)
			}

			if handlerErr != nil {
				_, _ = fmt.Fprintln(session.Stderr(), handlerErr.Error())
				_ = session.Exit(1)
				return
			}
			_ = session.Exit(0)
		},
	}
	sshServer.AddHostKey(hostKey)

	go func() {
		zap.L().Info("SSH server starting", zap.String("port", cfg.Ssh.Port))
		if err := sshServer.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			zap.L().Fatal("SSH server error", zap.Error(err))
		}
	}()

	return sshServer
}

func loadOrGenerateHostKey(path string) (gossh.Signer, error) {
	// #nosec G304 -- path is from application config, not user input
	keyBytes, err := os.ReadFile(path)
	if err == nil {
		signer, err := gossh.ParsePrivateKey(keyBytes)
		if err == nil {
			zap.L().Info("Loaded SSH host key", zap.String("path", path))
			return signer, nil
		}
	}

	zap.L().Info("Generating SSH host key", zap.String("path", path))
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err := os.WriteFile(path, privateKeyPEM, 0600); err != nil {
		return nil, err
	}

	return gossh.NewSignerFromKey(privateKey)
}
