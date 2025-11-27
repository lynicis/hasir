package organization

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
	CreateOrganization(ctx context.Context, org *OrganizationDTO) error
	GetOrganizations(ctx context.Context) (*[]OrganizationDTO, error)
	GetOrganizationByName(ctx context.Context, name string) (*OrganizationDTO, error)
}

var (
	ErrFailedAcquireConnection   = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	ErrOrganizationAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("organization already exists"))
	ErrOrganizationNotFound      = connect.NewError(connect.CodeNotFound, errors.New("organization not found"))
	ErrUniqueViolationCode       = "23505"
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

	tracer := traceProvider.Tracer("OrganizationPostgreSQLRepository")

	return &PgRepository{
		connectionPool: pgConnectionPool,
		tracer:         tracer,
	}
}

func (r *PgRepository) CreateOrganization(ctx context.Context, org *OrganizationDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateOrganization", trace.WithAttributes(attribute.KeyValue{
		Key:   "newOrganization",
		Value: attribute.StringValue(fmt.Sprintf("%+v", org)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `INSERT INTO organizations (id, name, visibility, created_by, created_at) 
			VALUES (@Id, @Name, @Visibility, @CreatedBy, @CreatedAt)`
	sqlArgs := pgx.NamedArgs{
		"Id":         org.Id,
		"Name":       org.Name,
		"Visibility": org.Visibility,
		"CreatedBy":  org.CreatedBy,
		"CreatedAt":  time.Now().UTC(),
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrOrganizationAlreadyExists
			}
			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to execute insert organization query"))
	}

	return nil
}

func (r *PgRepository) GetOrganizations(ctx context.Context) (*[]OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizations")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM organizations WHERE deleted_at IS NULL ORDER BY created_at DESC"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query organizations"))
	}
	defer rows.Close()

	orgs, err := pgx.CollectRows[OrganizationDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect rows"))
	}

	return &orgs, nil
}

func (r *PgRepository) GetOrganizationByName(ctx context.Context, name string) (*OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizationByName", trace.WithAttributes(
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

	sql := "SELECT * FROM organizations WHERE name = $1 AND deleted_at IS NULL"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("something went wrong"))
	}
	defer rows.Close()

	var org OrganizationDTO
	org, err = pgx.CollectOneRow[OrganizationDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrganizationNotFound
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &org, nil
}
