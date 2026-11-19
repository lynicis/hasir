package postgres

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

type JobTableQueries struct {
	Processing  string
	Completed   string
	Failed      string
	Pending     string
	Default     string
	UpdateError string
	NotFound    string
}

var (
	EmailJobQueries = JobTableQueries{
		Processing:  `UPDATE email_jobs SET status = @Status, processed_at = @ProcessedAt, attempts = attempts + 1 WHERE id = @Id AND status = 'pending'`,
		Completed:   `UPDATE email_jobs SET status = @Status, completed_at = @CompletedAt WHERE id = @Id AND status = 'processing'`,
		Failed:      `UPDATE email_jobs SET status = @Status, error_message = @ErrorMessage WHERE id = @Id AND status = 'processing'`,
		Pending:     `UPDATE email_jobs SET status = @Status WHERE id = @Id AND status = 'processing'`,
		Default:     `UPDATE email_jobs SET status = @Status WHERE id = @Id`,
		UpdateError: "failed to update email job status",
		NotFound:    "email job not found or status mismatch",
	}
	SdkTriggerJobQueries = JobTableQueries{
		Processing:  `UPDATE sdk_trigger_jobs SET status = @Status, processed_at = @ProcessedAt, attempts = attempts + 1 WHERE id = @Id AND status = 'pending'`,
		Completed:   `UPDATE sdk_trigger_jobs SET status = @Status, completed_at = @CompletedAt WHERE id = @Id AND status = 'processing'`,
		Failed:      `UPDATE sdk_trigger_jobs SET status = @Status, error_message = @ErrorMessage WHERE id = @Id AND status = 'processing'`,
		Pending:     `UPDATE sdk_trigger_jobs SET status = @Status WHERE id = @Id AND status = 'processing'`,
		Default:     `UPDATE sdk_trigger_jobs SET status = @Status WHERE id = @Id`,
		UpdateError: "failed to update sdk trigger job status",
		NotFound:    "sdk trigger job not found or status mismatch",
	}
	SdkGenerationJobQueries = JobTableQueries{
		Processing:  `UPDATE sdk_generation_jobs SET status = @Status, processed_at = @ProcessedAt, attempts = attempts + 1 WHERE id = @Id AND status = 'pending'`,
		Completed:   `UPDATE sdk_generation_jobs SET status = @Status, completed_at = @CompletedAt WHERE id = @Id AND status = 'processing'`,
		Failed:      `UPDATE sdk_generation_jobs SET status = @Status, error_message = @ErrorMessage WHERE id = @Id AND status = 'processing'`,
		Pending:     `UPDATE sdk_generation_jobs SET status = @Status WHERE id = @Id AND status = 'processing'`,
		Default:     `UPDATE sdk_generation_jobs SET status = @Status WHERE id = @Id`,
		UpdateError: "failed to update sdk generation job status",
		NotFound:    "sdk generation job not found or status mismatch",
	}
)

func BuildJobStatusUpdateQuery(queries JobTableQueries, jobId, status string, errorMsg *string, now time.Time) (string, pgx.NamedArgs) {
	switch status {
	case JobStatusProcessing:
		return queries.Processing, pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"ProcessedAt": now,
		}
	case JobStatusCompleted:
		return queries.Completed, pgx.NamedArgs{
			"Id":          jobId,
			"Status":      status,
			"CompletedAt": now,
		}
	case JobStatusFailed:
		return queries.Failed, pgx.NamedArgs{
			"Id":           jobId,
			"Status":       status,
			"ErrorMessage": errorMsg,
		}
	case JobStatusPending:
		return queries.Pending, pgx.NamedArgs{
			"Id":     jobId,
			"Status": status,
		}
	default:
		return queries.Default, pgx.NamedArgs{
			"Id":     jobId,
			"Status": status,
		}
	}
}

func UpdateJobStatus(
	ctx context.Context,
	pool *pgxpool.Pool,
	tracer trace.Tracer,
	spanName string,
	queries JobTableQueries,
	jobId string,
	status string,
	errorMsg *string,
) error {
	var span trace.Span
	ctx, span = tracer.Start(ctx, spanName, trace.WithAttributes(
		attribute.String("jobId", jobId),
		attribute.String("status", status),
	))
	defer span.End()

	connection, err := pool.Acquire(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	}
	defer connection.Release()

	sql, sqlArgs := BuildJobStatusUpdateQuery(queries, jobId, status, errorMsg, time.Now().UTC())

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New(queries.UpdateError))
	}

	if result.RowsAffected() == 0 {
		return connect.NewError(connect.CodeNotFound, errors.New(queries.NotFound))
	}

	return nil
}

func HandleJobRetryOrFail(
	attempts int,
	maxAttempts int,
	jobErr error,
	resetPending func() error,
	markFailed func(errorMsg string) error,
) {
	if attempts < maxAttempts {
		if updateErr := resetPending(); updateErr != nil {
			zap.L().Error("failed to reset job to pending", zap.Error(updateErr))
		}
		return
	}

	if updateErr := markFailed(jobErr.Error()); updateErr != nil {
		zap.L().Error("failed to mark job as failed", zap.Error(updateErr))
	}
}
