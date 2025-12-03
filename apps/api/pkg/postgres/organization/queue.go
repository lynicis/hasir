package organization

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"hasir-api/internal/organization"
	"hasir-api/pkg/email"
)

type EmailJobQueue struct {
	connectionPool *pgxpool.Pool
	tracer         trace.Tracer
	stopChan       chan struct{}
	stopOnce       sync.Once
	processorWg    sync.WaitGroup
}

func NewEmailJobQueue(connectionPool *pgxpool.Pool, tracer trace.Tracer) *EmailJobQueue {
	return &EmailJobQueue{
		connectionPool: connectionPool,
		tracer:         tracer,
		stopChan:       make(chan struct{}),
	}
}

func (q *EmailJobQueue) Start(ctx context.Context, emailService email.Service, batchSize int, pollInterval time.Duration) {
	q.processorWg.Add(1)
	go func() {
		defer q.processorWg.Done()
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		zap.L().Info("email job processor started",
			zap.Int("batchSize", batchSize),
			zap.Duration("pollInterval", pollInterval))

		for {
			select {
			case <-q.stopChan:
				zap.L().Info("email job processor stopping")
				return
			case <-ticker.C:
				q.processEmailJobs(ctx, emailService, batchSize)
			}
		}
	}()
}

func (q *EmailJobQueue) Stop() {
	q.stopOnce.Do(func() {
		close(q.stopChan)
		q.processorWg.Wait()
		zap.L().Info("email job processor stopped")
	})
}

func (q *EmailJobQueue) EnqueueEmailJobs(ctx context.Context, jobs []*organization.EmailJobDTO) error {
	if len(jobs) == 0 {
		return nil
	}

	connection, err := q.connectionPool.Acquire(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	}
	defer connection.Release()

	tx, err := connection.Begin(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to begin transaction"))
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	jobSQL := `INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, attempts, max_attempts, status, created_at)
			VALUES (@Id, @InviteId, @OrganizationId, @Email, @OrganizationName, @InviteToken, @Attempts, @MaxAttempts, @Status, @CreatedAt)`

	jobBatch := &pgx.Batch{}
	for _, job := range jobs {
		jobArgs := pgx.NamedArgs{
			"Id":               job.Id,
			"InviteId":         job.InviteId,
			"OrganizationId":   job.OrganizationId,
			"Email":            job.Email,
			"OrganizationName": job.OrganizationName,
			"InviteToken":      job.InviteToken,
			"Attempts":         job.Attempts,
			"MaxAttempts":      job.MaxAttempts,
			"Status":           job.Status,
			"CreatedAt":        job.CreatedAt,
		}
		jobBatch.Queue(jobSQL, jobArgs)
	}

	jobResults := tx.SendBatch(ctx, jobBatch)

	for i := 0; i < len(jobs); i++ {
		_, err := jobResults.Exec()
		if err != nil {
			_ = jobResults.Close()
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to enqueue job %d: %w", i, err))
		}
	}

	if err := jobResults.Close(); err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to close job batch results: %w", err))
	}

	if err := tx.Commit(ctx); err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to commit transaction"))
	}
	committed = true

	zap.L().Info("email jobs enqueued successfully", zap.Int("count", len(jobs)))

	return nil
}

func (q *EmailJobQueue) processEmailJobs(ctx context.Context, emailService email.Service, batchSize int) {
	jobs, err := q.GetPendingEmailJobs(ctx, batchSize)
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
				if updateErr := q.UpdateEmailJobStatus(ctx, job.Id, organization.EmailJobStatusPending, nil); updateErr != nil {
					zap.L().Error("failed to reset job to pending", zap.Error(updateErr))
				}
			} else {
				errorMsg := err.Error()
				if updateErr := q.UpdateEmailJobStatus(ctx, job.Id, organization.EmailJobStatusFailed, &errorMsg); updateErr != nil {
					zap.L().Error("failed to mark job as failed", zap.Error(updateErr))
				}
			}
			continue
		}

		if err := q.UpdateEmailJobStatus(ctx, job.Id, organization.EmailJobStatusCompleted, nil); err != nil {
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

func (q *EmailJobQueue) GetPendingEmailJobs(ctx context.Context, limit int) ([]*organization.EmailJobDTO, error) {
	var span trace.Span
	ctx, span = q.tracer.Start(ctx, "GetPendingEmailJobs", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "limit",
			Value: attribute.IntValue(limit),
		},
	))
	defer span.End()

	connection, err := q.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
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

	jobs, err := pgx.CollectRows(rows, pgx.RowToStructByName[organization.EmailJobDTO])
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect email job rows"))
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to commit transaction"))
	}
	committed = true

	result := make([]*organization.EmailJobDTO, len(jobs))
	for i := range jobs {
		result[i] = &jobs[i]
	}

	return result, nil
}

func (q *EmailJobQueue) UpdateEmailJobStatus(ctx context.Context, jobId string, status organization.EmailJobStatus, errorMsg *string) error {
	var span trace.Span
	ctx, span = q.tracer.Start(ctx, "UpdateEmailJobStatus", trace.WithAttributes(
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

	connection, err := q.connectionPool.Acquire(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	}
	defer connection.Release()

	now := time.Now().UTC()
	var sql string
	var sqlArgs pgx.NamedArgs

	switch status {
	case organization.EmailJobStatusProcessing:
		sql = `UPDATE email_jobs SET status = @Status, processed_at = @ProcessedAt, attempts = attempts + 1 WHERE id = @Id AND status = 'pending'`
		sqlArgs = pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"ProcessedAt": now,
		}
	case organization.EmailJobStatusCompleted:
		sql = `UPDATE email_jobs SET status = @Status, completed_at = @CompletedAt WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"CompletedAt": now,
		}
	case organization.EmailJobStatusFailed:
		sql = `UPDATE email_jobs SET status = @Status, error_message = @ErrorMessage WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":           jobId,
			"Status":       status,
			"ErrorMessage": errorMsg,
		}
	case organization.EmailJobStatusPending:
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