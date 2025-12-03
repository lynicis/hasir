package registry

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
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"hasir-api/internal/registry"
	"hasir-api/pkg/config"
)

var (
	ErrFailedAcquireConnection = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
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

	if traceProvider != nil {
		pgConfig.ConnConfig.Tracer = otelpgx.NewTracer(
			otelpgx.WithTracerProvider(traceProvider),
			otelpgx.WithDisableConnectionDetailsInAttributes(),
		)
	}

	var pgConnectionPool *pgxpool.Pool
	pgConnectionPool, err = pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		zap.L().Fatal("failed to connect database", zap.Error(err))
	}

	if traceProvider != nil {
		if err := otelpgx.RecordStats(pgConnectionPool); err != nil {
			zap.L().Fatal("unable to record database stats", zap.Error(err))
		}
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

	var tracer trace.Tracer
	if traceProvider != nil {
		tracer = traceProvider.Tracer("RepositoryPostgreSQLRepository")
	} else {
		tracer = noop.NewTracerProvider().Tracer("RepositoryPostgreSQLRepository")
	}

	return &PgRepository{
		connectionPool: pgConnectionPool,
		tracer:         tracer,
	}
}

func (r *PgRepository) CreateRepository(ctx context.Context, repo *registry.RepositoryDTO) error {
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

	now := time.Now().UTC()
	sql := `INSERT INTO repositories (id, name, created_by, organization_id, path, visibility, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err = connection.Exec(ctx, sql,
		repo.Id,
		repo.Name,
		repo.CreatedBy,
		repo.OrganizationId,
		repo.Path,
		repo.Visibility,
		now,
		&now,
	); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == ErrUniqueViolationCode {
			return registry.ErrRepositoryAlreadyExists
		}

		return connect.NewError(
			connect.CodeInternal,
			errors.New("failed to execute insert repository query"),
		)
	}

	return nil
}

func (r *PgRepository) GetRepositoryByName(ctx context.Context, name string) (*registry.RepositoryDTO, error) {
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
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repository by name"))
	}
	defer rows.Close()

	var repo registry.RepositoryDTO
	repo, err = pgx.CollectOneRow[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, registry.ErrRepositoryNotFound
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &repo, nil
}

func (r *PgRepository) GetRepositories(ctx context.Context, page, pageSize int) (*[]registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositories", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "page",
			Value: attribute.IntValue(page),
		},
		attribute.KeyValue{
			Key:   "pageSize",
			Value: attribute.IntValue(pageSize),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	offset := (page - 1) * pageSize
	sql := "SELECT * FROM repositories WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repositories"))
	}
	defer rows.Close()

	repos, err := pgx.CollectRows[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect rows"))
	}

	return &repos, nil
}

func (r *PgRepository) GetRepositoriesByUser(ctx context.Context, userId string, page, pageSize int) (*[]registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesByUser", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
		attribute.KeyValue{
			Key:   "page",
			Value: attribute.IntValue(page),
		},
		attribute.KeyValue{
			Key:   "pageSize",
			Value: attribute.IntValue(pageSize),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	offset := (page - 1) * pageSize
	sql := `
		SELECT r.*
		FROM repositories r
		INNER JOIN organization_members m ON m.organization_id = r.organization_id
		WHERE m.user_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := connection.Query(ctx, sql, userId, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repositories by user"))
	}
	defer rows.Close()

	repos, err := pgx.CollectRows[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect repositories by user rows"))
	}

	return &repos, nil
}

func (r *PgRepository) GetRepositoriesByOrganizationId(ctx context.Context, organizationId string) (*[]registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesByOrganizationId", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organization_id",
			Value: attribute.StringValue(organizationId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM repositories WHERE organization_id = $1 AND deleted_at IS NULL"

	rows, err := connection.Query(ctx, sql, organizationId)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repositories by organization id"))
	}
	defer rows.Close()

	repos, err := pgx.CollectRows[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect rows"))
	}

	return &repos, nil
}

func (r *PgRepository) GetRepositoriesCount(ctx context.Context) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesCount")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT COUNT(*) FROM repositories WHERE deleted_at IS NULL"

	var count int
	err = connection.QueryRow(ctx, sql).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count repositories"))
	}

	return count, nil
}

func (r *PgRepository) GetRepositoriesByUserCount(ctx context.Context, userId string) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesByUserCount", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		SELECT COUNT(*)
		FROM repositories r
		INNER JOIN organization_members m ON m.organization_id = r.organization_id
		WHERE m.user_id = $1 AND r.deleted_at IS NULL`

	var count int
	err = connection.QueryRow(ctx, sql, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count repositories by user"))
	}

	return count, nil
}

func (r *PgRepository) GetRepositoryById(ctx context.Context, id string) (*registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoryById", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM repositories WHERE id = $1 AND deleted_at IS NULL"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, id)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repository by id"))
	}
	defer rows.Close()

	var repo registry.RepositoryDTO
	repo, err = pgx.CollectOneRow[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, registry.ErrRepositoryNotFound
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &repo, nil
}

func (r *PgRepository) UpdateRepository(ctx context.Context, repo *registry.RepositoryDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateRepository", trace.WithAttributes(attribute.KeyValue{
		Key:   "repository",
		Value: attribute.StringValue(fmt.Sprintf("%+v", repo)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	now := time.Now().UTC()
	sql := `UPDATE repositories
			SET name = $1, updated_at = $2
			WHERE id = $3 AND deleted_at IS NULL`

	result, err := connection.Exec(ctx, sql, repo.Name, &now, repo.Id)
	if err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return registry.ErrRepositoryAlreadyExists
			}
			return connect.NewError(connect.CodeInternal, err)
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to execute update repository query"))
	}

	if result.RowsAffected() == 0 {
		return registry.ErrRepositoryNotFound
	}

	return nil
}

func (r *PgRepository) DeleteRepository(ctx context.Context, id string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteRepository", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	now := time.Now().UTC()
	sql := `UPDATE repositories
			SET deleted_at = $1
			WHERE id = $2 AND deleted_at IS NULL`

	result, err := connection.Exec(ctx, sql, &now, id)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(
			connect.CodeInternal,
			errors.New("failed to execute delete repository query"),
		)
	}

	if result.RowsAffected() == 0 {
		return registry.ErrRepositoryNotFound
	}

	return nil
}

