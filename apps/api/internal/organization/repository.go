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
	GetOrganizationByName(ctx context.Context, name string) (*OrganizationDTO, error)
	GetOrganizationById(ctx context.Context, id string) (*OrganizationDTO, error)
	DeleteOrganization(ctx context.Context, id string) error
	CreateInvite(ctx context.Context, invite *OrganizationInviteDTO) error
	GetInviteByToken(ctx context.Context, token string) (*OrganizationInviteDTO, error)
	UpdateInviteStatus(ctx context.Context, id string, status InviteStatus, acceptedAt *time.Time) error
	AddMember(ctx context.Context, member *OrganizationMemberDTO) error
	EnqueueEmailJobs(ctx context.Context, jobs []*EmailJobDTO) error
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

func (r *PgRepository) EnqueueEmailJobs(ctx context.Context, jobs []*EmailJobDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "EnqueueEmailJobs", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "jobCount",
			Value: attribute.IntValue(len(jobs)),
		},
	))
	defer span.End()

	if len(jobs) == 0 {
		return nil
	}

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	batch := &pgx.Batch{}
	for _, job := range jobs {
		sql := `INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at) 
				VALUES (@Id, @InviteId, @OrganizationId, @Email, @OrganizationName, @InviteToken, @Status, @Attempts, @MaxAttempts, @CreatedAt)`
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
		batch.Queue(sql, sqlArgs)
	}

	results := connection.SendBatch(ctx, batch)
	defer func() {
		_ = results.Close()
	}()

	for i := 0; i < len(jobs); i++ {
		_, err := results.Exec()
		if err != nil {
			span.RecordError(err)
			_ = results.Close()
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to enqueue email job %d: %w", i, err))
		}
	}

	zap.L().Info("enqueued email jobs", zap.Int("count", len(jobs)))
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
