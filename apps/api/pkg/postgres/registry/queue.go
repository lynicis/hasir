package registry

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

	"hasir-api/internal/registry"
)

type SdkGenerationJobQueue struct {
	connectionPool *pgxpool.Pool
	tracer         trace.Tracer
	stopChan       chan struct{}
	stopOnce       sync.Once
	processorWg    sync.WaitGroup
}

func NewSdkGenerationJobQueue(connectionPool *pgxpool.Pool, tracer trace.Tracer) *SdkGenerationJobQueue {
	return &SdkGenerationJobQueue{
		connectionPool: connectionPool,
		tracer:         tracer,
		stopChan:       make(chan struct{}),
	}
}

func (q *SdkGenerationJobQueue) Start(
	ctx context.Context,
	sdkGenerator registry.SdkGenerator,
	triggerProcessor registry.SdkTriggerProcessor,
	batchSize int,
	pollInterval time.Duration,
) {
	q.processorWg.Add(1)
	go func() {
		defer q.processorWg.Done()
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		zap.L().Info("sdk generation job processor started",
			zap.Int("batchSize", batchSize),
			zap.Duration("pollInterval", pollInterval))

		for {
			select {
			case <-q.stopChan:
				zap.L().Info("sdk generation job processor stopping")
				return
			case <-ticker.C:
				q.processSdkTriggerJobs(ctx, triggerProcessor, batchSize)
				q.processSdkGenerationJobs(ctx, sdkGenerator, batchSize)
			}
		}
	}()
}

func (q *SdkGenerationJobQueue) Stop() {
	q.stopOnce.Do(func() {
		close(q.stopChan)
		q.processorWg.Wait()
		zap.L().Info("sdk generation job processor stopped")
	})
}

func (q *SdkGenerationJobQueue) EnqueueSdkGenerationJobs(ctx context.Context, jobs []*registry.SdkGenerationJobDTO) error {
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

	jobSQL := `INSERT INTO sdk_generation_jobs (id, repository_id, commit_hash, sdk, status, attempts, max_attempts, created_at)
			VALUES (@Id, @RepositoryId, @CommitHash, @Sdk, @Status, @Attempts, @MaxAttempts, @CreatedAt)`

	jobBatch := &pgx.Batch{}
	for _, job := range jobs {
		jobArgs := pgx.NamedArgs{
			"Id":           job.Id,
			"RepositoryId": job.RepositoryId,
			"CommitHash":   job.CommitHash,
			"Sdk":          job.Sdk,
			"Status":       job.Status,
			"Attempts":     job.Attempts,
			"MaxAttempts":  job.MaxAttempts,
			"CreatedAt":    job.CreatedAt,
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

	zap.L().Info("sdk generation jobs enqueued successfully", zap.Int("count", len(jobs)))

	return nil
}

func (q *SdkGenerationJobQueue) EnqueueSdkTriggerJob(ctx context.Context, job *registry.SdkTriggerJobDTO) error {
	connection, err := q.connectionPool.Acquire(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	}
	defer connection.Release()

	sql := `INSERT INTO sdk_trigger_jobs (id, repository_id, repo_path, status, attempts, max_attempts, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = connection.Exec(ctx, sql,
		job.Id,
		job.RepositoryId,
		job.RepoPath,
		job.Status,
		job.Attempts,
		job.MaxAttempts,
		job.CreatedAt,
	)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to enqueue sdk trigger job: %w", err))
	}

	zap.L().Info("sdk trigger job enqueued successfully",
		zap.String("jobId", job.Id),
		zap.String("repositoryId", job.RepositoryId))

	return nil
}

func (q *SdkGenerationJobQueue) GetPendingSdkTriggerJobs(ctx context.Context, limit int) ([]*registry.SdkTriggerJobDTO, error) {
	var span trace.Span
	ctx, span = q.tracer.Start(ctx, "GetPendingSdkTriggerJobs", trace.WithAttributes(
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

	sql := `UPDATE sdk_trigger_jobs
			SET status = 'processing', processed_at = NOW(), attempts = attempts + 1
			WHERE id IN (
				SELECT id FROM sdk_trigger_jobs
				WHERE status = 'pending'
				ORDER BY created_at ASC
				LIMIT $1
				FOR UPDATE SKIP LOCKED
			)
			RETURNING id, repository_id, repo_path, status, attempts, max_attempts, created_at, processed_at, completed_at, error_message`

	rows, err := tx.Query(ctx, sql, limit)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query and update pending sdk trigger jobs"))
	}
	defer rows.Close()

	jobs, err := pgx.CollectRows(rows, pgx.RowToStructByName[registry.SdkTriggerJobDTO])
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect sdk trigger job rows"))
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to commit transaction"))
	}
	committed = true

	result := make([]*registry.SdkTriggerJobDTO, len(jobs))
	for i := range jobs {
		result[i] = &jobs[i]
	}

	return result, nil
}

func (q *SdkGenerationJobQueue) UpdateSdkTriggerJobStatus(ctx context.Context, jobId string, status registry.SdkGenerationJobStatus, errorMsg *string) error {
	var span trace.Span
	ctx, span = q.tracer.Start(ctx, "UpdateSdkTriggerJobStatus", trace.WithAttributes(
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
	case registry.SdkGenerationJobStatusProcessing:
		sql = `UPDATE sdk_trigger_jobs SET status = @Status, processed_at = @ProcessedAt, attempts = attempts + 1 WHERE id = @Id AND status = 'pending'`
		sqlArgs = pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"ProcessedAt": now,
		}
	case registry.SdkGenerationJobStatusCompleted:
		sql = `UPDATE sdk_trigger_jobs SET status = @Status, completed_at = @CompletedAt WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"CompletedAt": now,
		}
	case registry.SdkGenerationJobStatusFailed:
		sql = `UPDATE sdk_trigger_jobs SET status = @Status, error_message = @ErrorMessage WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":           jobId,
			"Status":       status,
			"ErrorMessage": errorMsg,
		}
	case registry.SdkGenerationJobStatusPending:
		sql = `UPDATE sdk_trigger_jobs SET status = @Status WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":     jobId,
			"Status": status,
		}
	default:
		sql = `UPDATE sdk_trigger_jobs SET status = @Status WHERE id = @Id`
		sqlArgs = pgx.NamedArgs{
			"Id":     jobId,
			"Status": status,
		}
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to update sdk trigger job status"))
	}

	if result.RowsAffected() == 0 {
		return connect.NewError(connect.CodeNotFound, errors.New("sdk trigger job not found or status mismatch"))
	}

	return nil
}

func (q *SdkGenerationJobQueue) processSdkTriggerJobs(ctx context.Context, triggerProcessor registry.SdkTriggerProcessor, batchSize int) {
	jobs, err := q.GetPendingSdkTriggerJobs(ctx, batchSize)
	if err != nil {
		zap.L().Error("failed to get pending sdk trigger jobs", zap.Error(err))
		return
	}

	if len(jobs) == 0 {
		return
	}

	zap.L().Info("processing sdk trigger jobs", zap.Int("count", len(jobs)))

	for _, job := range jobs {
		err := triggerProcessor.ProcessSdkTrigger(ctx, job.RepositoryId, job.RepoPath)
		if err != nil {
			zap.L().Error("failed to process sdk trigger",
				zap.Error(err),
				zap.String("jobId", job.Id),
				zap.String("repositoryId", job.RepositoryId))

			if job.Attempts < job.MaxAttempts {
				if updateErr := q.UpdateSdkTriggerJobStatus(ctx, job.Id, registry.SdkGenerationJobStatusPending, nil); updateErr != nil {
					zap.L().Error("failed to reset trigger job to pending", zap.Error(updateErr))
				}
			} else {
				errorMsg := err.Error()
				if updateErr := q.UpdateSdkTriggerJobStatus(ctx, job.Id, registry.SdkGenerationJobStatusFailed, &errorMsg); updateErr != nil {
					zap.L().Error("failed to mark trigger job as failed", zap.Error(updateErr))
				}
			}
			continue
		}

		if err := q.UpdateSdkTriggerJobStatus(ctx, job.Id, registry.SdkGenerationJobStatusCompleted, nil); err != nil {
			zap.L().Error("failed to update trigger job status to completed",
				zap.Error(err),
				zap.String("jobId", job.Id))
			continue
		}

		zap.L().Info("sdk trigger job completed successfully",
			zap.String("jobId", job.Id),
			zap.String("repositoryId", job.RepositoryId))
	}
}

func (q *SdkGenerationJobQueue) processSdkGenerationJobs(ctx context.Context, sdkGenerator registry.SdkGenerator, batchSize int) {
	jobs, err := q.GetPendingSdkGenerationJobs(ctx, batchSize)
	if err != nil {
		zap.L().Error("failed to get pending sdk generation jobs", zap.Error(err))
		return
	}

	if len(jobs) == 0 {
		return
	}

	zap.L().Info("processing sdk generation jobs", zap.Int("count", len(jobs)))

	for _, job := range jobs {
		err := sdkGenerator.GenerateSDK(ctx, job.RepositoryId, job.CommitHash, job.Sdk)
		if err != nil {
			zap.L().Error("failed to generate SDK",
				zap.Error(err),
				zap.String("jobId", job.Id),
				zap.String("repositoryId", job.RepositoryId),
				zap.String("commitHash", job.CommitHash),
				zap.String("sdk", string(job.Sdk)))

			if job.Attempts < job.MaxAttempts {
				if updateErr := q.UpdateSdkGenerationJobStatus(ctx, job.Id, registry.SdkGenerationJobStatusPending, nil); updateErr != nil {
					zap.L().Error("failed to reset job to pending", zap.Error(updateErr))
				}
			} else {
				errorMsg := err.Error()
				if updateErr := q.UpdateSdkGenerationJobStatus(ctx, job.Id, registry.SdkGenerationJobStatusFailed, &errorMsg); updateErr != nil {
					zap.L().Error("failed to mark job as failed", zap.Error(updateErr))
				}
			}
			continue
		}

		if err := q.UpdateSdkGenerationJobStatus(ctx, job.Id, registry.SdkGenerationJobStatusCompleted, nil); err != nil {
			zap.L().Error("failed to update job status to completed",
				zap.Error(err),
				zap.String("jobId", job.Id))
			continue
		}

		zap.L().Info("sdk generation job completed successfully",
			zap.String("jobId", job.Id),
			zap.String("repositoryId", job.RepositoryId),
			zap.String("commitHash", job.CommitHash),
			zap.String("sdk", string(job.Sdk)))
	}
}

func (q *SdkGenerationJobQueue) GetPendingSdkGenerationJobs(ctx context.Context, limit int) ([]*registry.SdkGenerationJobDTO, error) {
	var span trace.Span
	ctx, span = q.tracer.Start(ctx, "GetPendingSdkGenerationJobs", trace.WithAttributes(
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

	sql := `UPDATE sdk_generation_jobs
			SET status = 'processing', processed_at = NOW(), attempts = attempts + 1
			WHERE id IN (
				SELECT id FROM sdk_generation_jobs
				WHERE status = 'pending'
				ORDER BY created_at ASC
				LIMIT $1
				FOR UPDATE SKIP LOCKED
			)
			RETURNING id, repository_id, commit_hash, sdk, status, attempts, max_attempts, created_at, processed_at, completed_at, error_message`

	rows, err := tx.Query(ctx, sql, limit)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query and update pending sdk generation jobs"))
	}
	defer rows.Close()

	jobs, err := pgx.CollectRows(rows, pgx.RowToStructByName[registry.SdkGenerationJobDTO])
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect sdk generation job rows"))
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to commit transaction"))
	}
	committed = true

	result := make([]*registry.SdkGenerationJobDTO, len(jobs))
	for i := range jobs {
		result[i] = &jobs[i]
	}

	return result, nil
}

func (q *SdkGenerationJobQueue) UpdateSdkGenerationJobStatus(ctx context.Context, jobId string, status registry.SdkGenerationJobStatus, errorMsg *string) error {
	var span trace.Span
	ctx, span = q.tracer.Start(ctx, "UpdateSdkGenerationJobStatus", trace.WithAttributes(
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
	case registry.SdkGenerationJobStatusProcessing:
		sql = `UPDATE sdk_generation_jobs SET status = @Status, processed_at = @ProcessedAt, attempts = attempts + 1 WHERE id = @Id AND status = 'pending'`
		sqlArgs = pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"ProcessedAt": now,
		}
	case registry.SdkGenerationJobStatusCompleted:
		sql = `UPDATE sdk_generation_jobs SET status = @Status, completed_at = @CompletedAt WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"CompletedAt": now,
		}
	case registry.SdkGenerationJobStatusFailed:
		sql = `UPDATE sdk_generation_jobs SET status = @Status, error_message = @ErrorMessage WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":           jobId,
			"Status":       status,
			"ErrorMessage": errorMsg,
		}
	case registry.SdkGenerationJobStatusPending:
		sql = `UPDATE sdk_generation_jobs SET status = @Status WHERE id = @Id AND status = 'processing'`
		sqlArgs = pgx.NamedArgs{
			"Id":     jobId,
			"Status": status,
		}
	default:
		sql = `UPDATE sdk_generation_jobs SET status = @Status WHERE id = @Id`
		sqlArgs = pgx.NamedArgs{
			"Id":     jobId,
			"Status": status,
		}
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to update sdk generation job status"))
	}

	if result.RowsAffected() == 0 {
		return connect.NewError(connect.CodeNotFound, errors.New("sdk generation job not found or status mismatch"))
	}

	return nil
}
