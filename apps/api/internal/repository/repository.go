package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"apps/api/pkg/config"
)

type Repository interface {
	CreateRepository(ctx context.Context, repo *RepositoryDTO) error
	GetRepositoryByName(ctx context.Context, name string) (*RepositoryDTO, error)
	GetRepositories(ctx context.Context) (*[]RepositoryDTO, error)
}

var (
	ErrFailedAcquireConnection = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	ErrRepositoryAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("repository already exists"))
	ErrRepositoryNotFound      = connect.NewError(connect.CodeNotFound, errors.New("repository not found"))
	ErrUniqueViolationCode     = "23505"
)

type PgRepository struct {
	connectionPool *pgxpool.Pool
	tracer         trace.Tracer
}

func NewPgRepository(
	cfg *config.Config,
	traceProvider *sdktrace.TracerProvider,
) *PgRepository {
	credential := cfg.PostgresConfig.GetPostgresDsn()
	pgConfig, err := pgxpool.ParseConfig(credential)
	if err != nil {
		zap.L().Fatal("failed to parse database config", zap.Error(err))
	}

	pgConfig.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithTracerProvider(traceProvider),
		otelpgx.WithDisableConnectionDetailsInAttributes(),
	)

	var pgConnectionPool *pgxpool.Pool
	pgConnectionPool, err = pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		zap.L().Fatal("failed to connect database", zap.Error(err))
	}

	if err := otelpgx.RecordStats(pgConnectionPool); err != nil {
		zap.L().Fatal("unable to record database stats", zap.Error(err))
	}

	var connection *pgxpool.Conn
	connection, err = pgConnectionPool.Acquire(context.Background())
	if err != nil {
		zap.L().Fatal("failed to acquire connection", zap.Error(err))
	}
	defer connection.Release()

	err = connection.Ping(context.Background())
	if err != nil {
		zap.L().Fatal("failed to ping database", zap.Error(err))
	}

	tracer := traceProvider.Tracer("RepositoryPostgreSQLRepository")

	return &PgRepository{
		connectionPool: pgConnectionPool,
		tracer:         tracer,
	}
}

func (r *PgRepository) CreateRepository(ctx context.Context, repo *RepositoryDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateRepository", trace.WithAttributes(attribute.KeyValue{
		Key:   "newRepository",
		Value: attribute.StringValue(fmt.Sprintf("%+v", repo)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `INSERT INTO repositories (id, name, owner_id, path, created_at, updated_at) 
			VALUES (@Id, @Name, @OwnerId, @Path, @CreatedAt, @UpdatedAt)`
	sqlArgs := pgx.NamedArgs{
		"Id":        repo.Id,
		"Name":      repo.Name,
		"OwnerId":   repo.OwnerId,
		"Path":      repo.Path,
		"CreatedAt": time.Now().UTC(),
		"UpdatedAt": time.Now().UTC(),
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrRepositoryAlreadyExists
			}
			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to execute insert repository query"))
	}

	return nil
}

func (r *PgRepository) GetRepositoryByName(ctx context.Context, name string) (*RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoryByName", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "name",
			Value: attribute.StringValue(name),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM repositories WHERE name = $1 AND deleted_at IS NULL"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("something went wrong"))
	}
	defer rows.Close()

	var repo RepositoryDTO
	repo, err = pgx.CollectOneRow[RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositoryNotFound
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &repo, nil
}

func (r *PgRepository) GetRepositories(ctx context.Context) (*[]RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositories")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM repositories WHERE deleted_at IS NULL ORDER BY created_at DESC"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repositories"))
	}
	defer rows.Close()

	repos, err := pgx.CollectRows[RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect rows"))
	}

	return &repos, nil
}
