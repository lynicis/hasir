package registry

import (
	"context"
	"time"
)

type SdkGenerator interface {
	GenerateSDK(ctx context.Context, repositoryId, commitHash string, sdk SDK) error
}

type SdkTriggerProcessor interface {
	ProcessSdkTrigger(ctx context.Context, repositoryId, repoPath string) error
}

type SdkGenerationQueue interface {
	Start(ctx context.Context, sdkGenerator SdkGenerator, triggerProcessor SdkTriggerProcessor, batchSize int, pollInterval time.Duration)
	Stop()
	EnqueueSdkGenerationJobs(ctx context.Context, jobs []*SdkGenerationJobDTO) error
	EnqueueSdkTriggerJob(ctx context.Context, job *SdkTriggerJobDTO) error
	GetPendingSdkGenerationJobs(ctx context.Context, limit int) ([]*SdkGenerationJobDTO, error)
	GetPendingSdkTriggerJobs(ctx context.Context, limit int) ([]*SdkTriggerJobDTO, error)
	UpdateSdkGenerationJobStatus(ctx context.Context, jobId string, status SdkGenerationJobStatus, errorMsg *string) error
	UpdateSdkTriggerJobStatus(ctx context.Context, jobId string, status SdkGenerationJobStatus, errorMsg *string) error
}
