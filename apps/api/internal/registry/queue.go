package registry

import (
	"context"
	"time"
)

type SdkGenerator interface {
	GenerateSDK(ctx context.Context, repositoryId, commitHash string, sdk SDK) error
}

type SdkGenerationQueue interface {
	Start(ctx context.Context, sdkGenerator SdkGenerator, batchSize int, pollInterval time.Duration)
	Stop()
	EnqueueSdkGenerationJobs(ctx context.Context, jobs []*SdkGenerationJobDTO) error
	GetPendingSdkGenerationJobs(ctx context.Context, limit int) ([]*SdkGenerationJobDTO, error)
	UpdateSdkGenerationJobStatus(ctx context.Context, jobId string, status SdkGenerationJobStatus, errorMsg *string) error
}
