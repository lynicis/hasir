package organization

import (
	"context"
	"errors"
	"fmt"
	"sync"
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

	"hasir-api/pkg/config"
	"hasir-api/pkg/email"
)

type Repository interface {
	CreateOrganization(ctx context.Context, org *OrganizationDTO) error
	GetOrganizations(ctx context.Context, page, pageSize int) (*[]OrganizationDTO, error)
	GetOrganizationsCount(ctx context.Context) (int, error)
	GetUserOrganizations(ctx context.Context, userId string, page, pageSize int) (*[]OrganizationDTO, error)
	GetUserOrganizationsCount(ctx context.Context, userId string) (int, error)
	GetOrganizationByName(ctx context.Context, name string) (*OrganizationDTO, error)
	GetOrganizationById(ctx context.Context, id string) (*OrganizationDTO, error)
	UpdateOrganization(ctx context.Context, org *OrganizationDTO) error
	DeleteOrganization(ctx context.Context, id string) error
	CreateInvitesAndEnqueueEmailJobs(ctx context.Context, invites []*OrganizationInviteDTO, jobs []*EmailJobDTO) error
	GetInviteByToken(ctx context.Context, token string) (*OrganizationInviteDTO, error)
	UpdateInviteStatus(ctx context.Context, id string, status InviteStatus, acceptedAt *time.Time) error
	AddMember(ctx context.Context, member *OrganizationMemberDTO) error
	GetMembers(ctx context.Context, organizationId string) ([]*OrganizationMemberDTO, []string, []string, error)
	GetMemberRole(ctx context.Context, organizationId, userId string) (MemberRole, error)
	GetMemberRoleString(ctx context.Context, organizationId, userId string) (string, error)
	UpdateMemberRole(ctx context.Context, organizationId, userId string, role MemberRole) error
	DeleteMember(ctx context.Context, organizationId, userId string) error
	GetPendingEmailJobs(ctx context.Context, limit int) ([]*EmailJobDTO, error)
	UpdateEmailJobStatus(ctx context.Context, jobId string, status EmailJobStatus, errorMsg *string) error
	StartEmailJobProcessor(ctx context.Context, emailService email.Service, batchSize int, pollInterval time.Duration)
	StopEmailJobProcessor()
}

var (
	ErrFailedAcquireConnection   = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	ErrOrganizationAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("organization already exists"))
	ErrOrganizationNotFound      = connect.NewError(connect.CodeNotFound, errors.New("organization not found"))
	ErrInviteNotFound            = connect.NewError(connect.CodeNotFound, errors.New("invite not found"))
	ErrMemberAlreadyExists       = connect.NewError(connect.CodeAlreadyExists, errors.New("member already exists"))
	ErrMemberNotFound            = connect.NewError(connect.CodeNotFound, errors.New("member not found"))
	ErrUniqueViolationCode       = "23505"
)

type PgRepository struct {
	connectionPool *pgxpool.Pool
	tracer         trace.Tracer
	stopChan       chan struct{}
	processorWg    sync.WaitGroup
	stopOnce       sync.Once
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
		stopChan:       make(chan struct{}),
	}
}

func querySingleRow[T any](ctx context.Context, conn *pgxpool.Conn, span trace.Span, sql string, args []any, notFoundErr error) (*T, error) {
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to execute query"))
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow[T](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notFoundErr
		}
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &result, nil
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

func (r *PgRepository) GetOrganizations(ctx context.Context, page, pageSize int) (*[]OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizations", trace.WithAttributes(
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
	sql := "SELECT * FROM organizations WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, pageSize, offset)
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

func (r *PgRepository) GetUserOrganizations(ctx context.Context, userId string, page, pageSize int) (*[]OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetUserOrganizations", trace.WithAttributes(
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
		SELECT o.*
		FROM organizations o
		INNER JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND o.deleted_at IS NULL
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := connection.Query(ctx, sql, userId, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query user organizations"))
	}
	defer rows.Close()

	orgs, err := pgx.CollectRows[OrganizationDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect user organization rows"))
	}

	return &orgs, nil
}

func (r *PgRepository) GetOrganizationsCount(ctx context.Context) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizationsCount")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT COUNT(*) FROM organizations WHERE deleted_at IS NULL"

	var count int
	err = connection.QueryRow(ctx, sql).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count organizations"))
	}

	return count, nil
}

func (r *PgRepository) GetUserOrganizationsCount(ctx context.Context, userId string) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetUserOrganizationsCount", trace.WithAttributes(
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
		FROM organizations o
		INNER JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND o.deleted_at IS NULL`

	var count int
	err = connection.QueryRow(ctx, sql, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count user organizations"))
	}

	return count, nil
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
	return querySingleRow[OrganizationDTO](ctx, connection, span, sql, []any{name}, ErrOrganizationNotFound)
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
	return querySingleRow[OrganizationDTO](ctx, connection, span, sql, []any{id}, ErrOrganizationNotFound)
}

func (r *PgRepository) UpdateOrganization(ctx context.Context, org *OrganizationDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateOrganization", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(org.Id),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `UPDATE organizations
			SET name = @Name,
				visibility = @Visibility
			WHERE id = @Id AND deleted_at IS NULL`
	sqlArgs := pgx.NamedArgs{
		"Id":         org.Id,
		"Name":       org.Name,
		"Visibility": org.Visibility,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrOrganizationAlreadyExists
			}
			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to update organization"))
	}

	if result.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

func (r *PgRepository) DeleteOrganization(ctx context.Context, id string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteOrganization", trace.WithAttributes(
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
	sql := `UPDATE organizations SET deleted_at = @DeletedAt WHERE id = @Id AND deleted_at IS NULL`
	sqlArgs := pgx.NamedArgs{
		"Id":        id,
		"DeletedAt": now,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to delete organization"))
	}

	if result.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

func (r *PgRepository) CreateInvitesAndEnqueueEmailJobs(ctx context.Context, invites []*OrganizationInviteDTO, jobs []*EmailJobDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateInvitesAndEnqueueEmailJobs", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "inviteCount",
			Value: attribute.IntValue(len(invites)),
		},
		attribute.KeyValue{
			Key:   "jobCount",
			Value: attribute.IntValue(len(jobs)),
		},
	))
	defer span.End()

	if len(invites) == 0 {
		return nil
	}

	if len(invites) != len(jobs) {
		return connect.NewError(connect.CodeInternal, errors.New("invites and jobs must have the same length"))
	}

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	orgID := invites[0].OrganizationId

	emails := make([]string, 0, len(invites))
	for _, invite := range invites {
		emails = append(emails, invite.Email)
	}

	existingEmails := make(map[string]struct{}, len(emails))
	rows, err := connection.Query(
		ctx,
		`SELECT email FROM organization_invites WHERE organization_id = $1 AND status = 'pending' AND email = ANY($2)`,
		orgID,
		emails,
	)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to check existing pending invites"))
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			span.RecordError(err)
			return connect.NewError(connect.CodeInternal, errors.New("failed to scan existing pending invites"))
		}
		existingEmails[email] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to iterate existing pending invites"))
	}

	filteredInvites := make([]*OrganizationInviteDTO, 0, len(invites))
	filteredJobs := make([]*EmailJobDTO, 0, len(jobs))
	for i, invite := range invites {
		if _, found := existingEmails[invite.Email]; found {
			continue
		}

		filteredInvites = append(filteredInvites, invite)
		filteredJobs = append(filteredJobs, jobs[i])
	}

	invites = filteredInvites
	jobs = filteredJobs

	if len(invites) == 0 {
		return nil
	}

	tx, err := connection.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to begin transaction"))
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	inviteSQL := `INSERT INTO organization_invites (id, organization_id, email, token, invited_by, role, status, created_at, expires_at)
			VALUES (@Id, @OrganizationId, @Email, @Token, @InvitedBy, @Role, @Status, @CreatedAt, @ExpiresAt)`

	inviteBatch := &pgx.Batch{}
	for _, invite := range invites {
		sqlArgs := pgx.NamedArgs{
			"Id":             invite.Id,
			"OrganizationId": invite.OrganizationId,
			"Email":          invite.Email,
			"Token":          invite.Token,
			"InvitedBy":      invite.InvitedBy,
			"Role":           invite.Role,
			"Status":         invite.Status,
			"CreatedAt":      invite.CreatedAt,
			"ExpiresAt":      invite.ExpiresAt,
		}
		inviteBatch.Queue(inviteSQL, sqlArgs)
	}

	inviteResults := tx.SendBatch(ctx, inviteBatch)

	for i := 0; i < len(invites); i++ {
		_, err := inviteResults.Exec()
		if err != nil {
			span.RecordError(err)
			_ = inviteResults.Close()
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == ErrUniqueViolationCode {
					return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("invite already exists for email: %s", invites[i].Email))
				}
			}
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create invite %d: %w", i, err))
		}
	}

	if err := inviteResults.Close(); err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to close invite batch results: %w", err))
	}

	jobSQL := `INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at)
			VALUES (@Id, @InviteId, @OrganizationId, @Email, @OrganizationName, @InviteToken, @Status, @Attempts, @MaxAttempts, @CreatedAt)`

	jobBatch := &pgx.Batch{}
	for _, job := range jobs {
		sqlArgs := pgx.NamedArgs{
			"Id":               job.Id,
			"InviteId":         job.InviteId,
			"OrganizationId":   job.OrganizationId,
			"Email":            job.Email,
			"OrganizationName": job.OrganizationName,
			"InviteToken":      job.InviteToken,
			"Status":           job.Status,
			"Attempts":         job.Attempts,
			"MaxAttempts":      job.MaxAttempts,
			"CreatedAt":        job.CreatedAt,
		}
		jobBatch.Queue(jobSQL, sqlArgs)
	}

	jobResults := tx.SendBatch(ctx, jobBatch)

	for i := 0; i < len(jobs); i++ {
		_, err := jobResults.Exec()
		if err != nil {
			span.RecordError(err)
			_ = jobResults.Close()
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to enqueue email job %d: %w", i, err))
		}
	}

	if err := jobResults.Close(); err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to close job batch results: %w", err))
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to commit transaction"))
	}
	committed = true

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
	return querySingleRow[OrganizationInviteDTO](ctx, connection, span, sql, []any{token}, ErrInviteNotFound)
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

func (r *PgRepository) GetMembers(ctx context.Context, organizationId string) ([]*OrganizationMemberDTO, []string, []string, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetMembers", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, nil, nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `SELECT id, organization_id, user_id, role, joined_at, username, email
			FROM organization_members_view
			WHERE organization_id = $1
			ORDER BY joined_at ASC`

	rows, err := connection.Query(ctx, sql, organizationId)
	if err != nil {
		span.RecordError(err)
		return nil, nil, nil, connect.NewError(connect.CodeInternal, errors.New("failed to query members"))
	}
	defer rows.Close()

	memberRows, err := pgx.CollectRows[memberRow](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, nil, nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect member rows"))
	}

	members := make([]*OrganizationMemberDTO, len(memberRows))
	usernames := make([]string, len(memberRows))
	emails := make([]string, len(memberRows))

	for i, row := range memberRows {
		members[i] = &row.OrganizationMemberDTO
		usernames[i] = row.Username
		emails[i] = row.Email
	}

	return members, usernames, emails, nil
}

func (r *PgRepository) GetMemberRole(ctx context.Context, organizationId, userId string) (MemberRole, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetMemberRole", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return "", ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2`

	var role MemberRole
	err = connection.QueryRow(ctx, sql, organizationId, userId).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrMemberNotFound
		}
		span.RecordError(err)
		return "", connect.NewError(connect.CodeInternal, errors.New("failed to query member role"))
	}

	return role, nil
}

func (r *PgRepository) GetMemberRoleString(ctx context.Context, organizationId, userId string) (string, error) {
	role, err := r.GetMemberRole(ctx, organizationId, userId)
	return string(role), err
}

func (r *PgRepository) UpdateMemberRole(ctx context.Context, organizationId, userId string, role MemberRole) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateMemberRole", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
		attribute.KeyValue{
			Key:   "role",
			Value: attribute.StringValue(string(role)),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `UPDATE organization_members SET role = @Role WHERE organization_id = @OrganizationId AND user_id = @UserId`
	sqlArgs := pgx.NamedArgs{
		"OrganizationId": organizationId,
		"UserId":         userId,
		"Role":           role,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to update member role"))
	}

	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (r *PgRepository) DeleteMember(ctx context.Context, organizationId, userId string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteMember", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `DELETE FROM organization_members WHERE organization_id = @OrganizationId AND user_id = @UserId`
	sqlArgs := pgx.NamedArgs{
		"OrganizationId": organizationId,
		"UserId":         userId,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to delete member"))
	}

	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (r *PgRepository) GetPendingEmailJobs(ctx context.Context, limit int) ([]*EmailJobDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetPendingEmailJobs", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "limit",
			Value: attribute.IntValue(limit),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	tx, err := connection.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to begin transaction"))
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	sql := `UPDATE email_jobs
			SET status = 'processing', processed_at = NOW(), attempts = attempts + 1
			WHERE id IN (
				SELECT id FROM email_jobs
				WHERE status = 'pending'
				ORDER BY created_at ASC
				LIMIT $1
				FOR UPDATE SKIP LOCKED
			)
			RETURNING id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at, processed_at, completed_at, error_message`

	rows, err := tx.Query(ctx, sql, limit)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query and update pending email jobs"))
	}
	defer rows.Close()

	jobs, err := pgx.CollectRows(rows, pgx.RowToStructByName[EmailJobDTO])
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect email job rows"))
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to commit transaction"))
	}
	committed = true

	result := make([]*EmailJobDTO, len(jobs))
	for i := range jobs {
		result[i] = &jobs[i]
	}

	return result, nil
}

func (r *PgRepository) UpdateEmailJobStatus(ctx context.Context, jobId string, status EmailJobStatus, errorMsg *string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateEmailJobStatus", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "jobId",
			Value: attribute.StringValue(jobId),
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

	now := time.Now().UTC()
	var sql string
	var sqlArgs pgx.NamedArgs

	switch status {
	case EmailJobStatusProcessing:

		sql = `UPDATE email_jobs SET status = @Status, processed_at = @ProcessedAt, attempts = attempts + 1 WHERE id = @Id AND status = 'pending'`
		sqlArgs = pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"ProcessedAt": now,
		}
	case EmailJobStatusCompleted:
		sql = `UPDATE email_jobs SET status = @Status, completed_at = @CompletedAt WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"CompletedAt": now,
		}
	case EmailJobStatusFailed:
		sql = `UPDATE email_jobs SET status = @Status, error_message = @ErrorMessage WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":           jobId,
			"Status":       status,
			"ErrorMessage": errorMsg,
		}
	case EmailJobStatusPending:

		sql = `UPDATE email_jobs SET status = @Status WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":     jobId,
			"Status": status,
		}
	default:
		sql = `UPDATE email_jobs SET status = @Status WHERE id = @Id`
		sqlArgs = pgx.NamedArgs{
			"Id":     jobId,
			"Status": status,
		}
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to update email job status"))
	}

	if result.RowsAffected() == 0 {
		return connect.NewError(connect.CodeNotFound, errors.New("email job not found or status mismatch"))
	}

	return nil
}

func (r *PgRepository) StartEmailJobProcessor(ctx context.Context, emailService email.Service, batchSize int, pollInterval time.Duration) {
	r.processorWg.Add(1)
	go func() {
		defer r.processorWg.Done()
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		zap.L().Info("email job processor started",
			zap.Int("batchSize", batchSize),
			zap.Duration("pollInterval", pollInterval))

		for {
			select {
			case <-r.stopChan:
				zap.L().Info("email job processor stopping")
				return
			case <-ticker.C:
				r.processEmailJobs(ctx, emailService, batchSize)
			}
		}
	}()
}

func (r *PgRepository) StopEmailJobProcessor() {
	r.stopOnce.Do(func() {
		close(r.stopChan)
		r.processorWg.Wait()
		zap.L().Info("email job processor stopped")
	})
}

func (r *PgRepository) processEmailJobs(ctx context.Context, emailService email.Service, batchSize int) {
	jobs, err := r.GetPendingEmailJobs(ctx, batchSize)
	if err != nil {
		zap.L().Error("failed to get pending email jobs", zap.Error(err))
		return
	}

	if len(jobs) == 0 {
		return
	}

	zap.L().Info("processing email jobs", zap.Int("count", len(jobs)))

	for _, job := range jobs {
		err := emailService.SendInvite(job.Email, job.OrganizationName, job.InviteToken)
		if err != nil {
			zap.L().Error("failed to send invite email",
				zap.Error(err),
				zap.String("jobId", job.Id),
				zap.String("email", job.Email))

			if job.Attempts < job.MaxAttempts {
				if updateErr := r.UpdateEmailJobStatus(ctx, job.Id, EmailJobStatusPending, nil); updateErr != nil {
					zap.L().Error("failed to reset job to pending", zap.Error(updateErr))
				}
			} else {
				errorMsg := err.Error()
				if updateErr := r.UpdateEmailJobStatus(ctx, job.Id, EmailJobStatusFailed, &errorMsg); updateErr != nil {
					zap.L().Error("failed to mark job as failed", zap.Error(updateErr))
				}
			}
			continue
		}

		if err := r.UpdateEmailJobStatus(ctx, job.Id, EmailJobStatusCompleted, nil); err != nil {
			zap.L().Error("failed to update job status to completed",
				zap.Error(err),
				zap.String("jobId", job.Id))
			continue
		}

		zap.L().Info("email job completed successfully",
			zap.String("jobId", job.Id),
			zap.String("email", job.Email))
	}
}
