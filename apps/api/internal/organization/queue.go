package organization

import (
	"context"
	"time"

	"hasir-api/pkg/email"
)

type Queue interface {
	Start(ctx context.Context, emailService email.Service, batchSize int, pollInterval time.Duration)
	Stop()
	EnqueueEmailJobs(ctx context.Context, jobs []*EmailJobDTO) error
	GetPendingEmailJobs(ctx context.Context, limit int) ([]*EmailJobDTO, error)
	UpdateEmailJobStatus(ctx context.Context, jobId string, status EmailJobStatus, errorMsg *string) error
}
