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
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"apps/api/pkg/config"
)

type Repository interface {
	CreateOrganization(ctx context.Context, org *OrganizationDTO) error
	GetOrganizations(ctx context.Context) (*[]OrganizationDTO, error)
	GetOrganizationByName(ctx context.Context, name string) (*OrganizationDTO, error)
	GetOrganizationById(ctx context.Context, id string) (*OrganizationDTO, error)
	CreateInvite(ctx context.Context, invite *OrganizationInviteDTO) error
	GetInviteByToken(ctx context.Context, token string) (*OrganizationInviteDTO, error)
	UpdateInviteStatus(ctx context.Context, id string, status InviteStatus, acceptedAt *time.Time) error
	AddMember(ctx context.Context, member *OrganizationMemberDTO) error
}

var (
	ErrFailedAcquireConnection   = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	ErrOrganizationAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("organization already exists"))
	ErrOrganizationNotFound      = connect.NewError(connect.CodeNotFound, errors.New("organization not found"))
	ErrInviteNotFound            = connect.NewError(connect.CodeNotFound, errors.New("invite not found"))
	ErrMemberAlreadyExists       = connect.NewError(connect.CodeAlreadyExists, errors.New("member already exists"))
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
		tracer = traceProvider.Tracer("OrganizationPostgreSQLRepository")
	} else {
		tracer = noop.NewTracerProvider().Tracer("OrganizationPostgreSQLRepository")
	}

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

func (r *PgRepository) GetOrganizationById(ctx context.Context, id string) (*OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizationById", trace.WithAttributes(
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

	sql := "SELECT * FROM organizations WHERE id = $1 AND deleted_at IS NULL"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, id)
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

func (r *PgRepository) CreateInvite(ctx context.Context, invite *OrganizationInviteDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateInvite", trace.WithAttributes(attribute.KeyValue{
		Key:   "invite",
		Value: attribute.StringValue(fmt.Sprintf("%+v", invite)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `INSERT INTO organization_invites (id, organization_id, email, token, invited_by, status, created_at, expires_at) 
			VALUES (@Id, @OrganizationId, @Email, @Token, @InvitedBy, @Status, @CreatedAt, @ExpiresAt)`
	sqlArgs := pgx.NamedArgs{
		"Id":             invite.Id,
		"OrganizationId": invite.OrganizationId,
		"Email":          invite.Email,
		"Token":          invite.Token,
		"InvitedBy":      invite.InvitedBy,
		"Status":         invite.Status,
		"CreatedAt":      invite.CreatedAt,
		"ExpiresAt":      invite.ExpiresAt,
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return connect.NewError(connect.CodeAlreadyExists, errors.New("invite already exists"))
			}
			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to execute insert invite query"))
	}

	return nil
}

func (r *PgRepository) GetInviteByToken(ctx context.Context, token string) (*OrganizationInviteDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetInviteByToken", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "token",
			Value: attribute.StringValue(token),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM organization_invites WHERE token = $1"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, token)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query invite"))
	}
	defer rows.Close()

	var invite OrganizationInviteDTO
	invite, err = pgx.CollectOneRow[OrganizationInviteDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInviteNotFound
		}

		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &invite, nil
}

func (r *PgRepository) UpdateInviteStatus(ctx context.Context, id string, status InviteStatus, acceptedAt *time.Time) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateInviteStatus", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "status",
			Value: attribute.StringValue(string(status)),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `UPDATE organization_invites SET status = @Status, accepted_at = @AcceptedAt WHERE id = @Id`
	sqlArgs := pgx.NamedArgs{
		"Id":         id,
		"Status":     status,
		"AcceptedAt": acceptedAt,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to update invite status"))
	}

	if result.RowsAffected() == 0 {
		return ErrInviteNotFound
	}

	return nil
}

func (r *PgRepository) AddMember(ctx context.Context, member *OrganizationMemberDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "AddMember", trace.WithAttributes(attribute.KeyValue{
		Key:   "member",
		Value: attribute.StringValue(fmt.Sprintf("%+v", member)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `INSERT INTO organization_members (id, organization_id, user_id, role, joined_at) 
			VALUES (@Id, @OrganizationId, @UserId, @Role, @JoinedAt)`
	sqlArgs := pgx.NamedArgs{
		"Id":             member.Id,
		"OrganizationId": member.OrganizationId,
		"UserId":         member.UserId,
		"Role":           member.Role,
		"JoinedAt":       member.JoinedAt,
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrMemberAlreadyExists
			}
			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to add member"))
	}

	return nil
}
