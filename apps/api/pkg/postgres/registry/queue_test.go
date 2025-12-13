package registry

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"hasir-api/internal/registry"
)

type mockSdkGenerator struct {
	generateSDKFunc func(ctx context.Context, repositoryId, commitHash string, sdk registry.SDK) error
}

func (m *mockSdkGenerator) GenerateSDK(ctx context.Context, repositoryId, commitHash string, sdk registry.SDK) error {
	if m.generateSDKFunc != nil {
		return m.generateSDKFunc(ctx, repositoryId, commitHash, sdk)
	}
	return nil
}

func setupQueueTestEnvironment(t *testing.T) (*SdkGenerationJobQueue, *pgxpool.Pool, func()) {
	t.Helper()

	postgresContainer := setupPgContainer(t)

	connString, err := postgresContainer.ConnectionString(t.Context())
	require.NoError(t, err)

	createSdkGenerationJobsTable(t, connString)

	pool, err := pgxpool.New(t.Context(), connString)
	require.NoError(t, err)

	queue := NewSdkGenerationJobQueue(pool, noop.NewTracerProvider().Tracer("test"))

	cleanup := func() {
		pool.Close()
		if err := postgresContainer.Terminate(t.Context()); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}

	return queue, pool, cleanup
}

func createSdkGenerationJobsTable(t *testing.T, connString string) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sdkEnumSQL := `DO $$ BEGIN
		CREATE TYPE sdk_type AS ENUM (
			'GO_PROTOBUF',
			'GO_CONNECTRPC',
			'GO_GRPC',
			'JS_BUFBUILD_ES',
			'JS_PROTOBUF',
			'JS_CONNECTRPC'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$`
	_, err = conn.Exec(t.Context(), sdkEnumSQL)
	require.NoError(t, err)

	jobStatusEnumSQL := `DO $$ BEGIN
		CREATE TYPE sdk_generation_job_status AS ENUM (
			'pending',
			'processing',
			'completed',
			'failed'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$`
	_, err = conn.Exec(t.Context(), jobStatusEnumSQL)
	require.NoError(t, err)

	tableSQL := `CREATE TABLE IF NOT EXISTS sdk_generation_jobs (
		id VARCHAR(36) PRIMARY KEY,
		repository_id VARCHAR(36) NOT NULL,
		commit_hash VARCHAR(40) NOT NULL,
		sdk sdk_type NOT NULL,
		status sdk_generation_job_status NOT NULL DEFAULT 'pending',
		attempts INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 5,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		processed_at TIMESTAMP,
		completed_at TIMESTAMP,
		error_message TEXT
	)`
	_, err = conn.Exec(t.Context(), tableSQL)
	require.NoError(t, err)
}

func TestEnqueueSdkGenerationJobs(t *testing.T) {
	queue, pool, cleanup := setupQueueTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name      string
		jobs      []*registry.SdkGenerationJobDTO
		wantError bool
	}{
		{
			name:      "empty jobs list",
			jobs:      []*registry.SdkGenerationJobDTO{},
			wantError: false,
		},
		{
			name: "single job",
			jobs: []*registry.SdkGenerationJobDTO{
				{
					Id:           uuid.NewString(),
					RepositoryId: uuid.NewString(),
					CommitHash:   "abc123",
					Sdk:          registry.SdkGoProtobuf,
					Status:       registry.SdkGenerationJobStatusPending,
					Attempts:     0,
					MaxAttempts:  5,
					CreatedAt:    time.Now().UTC(),
				},
			},
			wantError: false,
		},
		{
			name: "multiple jobs",
			jobs: []*registry.SdkGenerationJobDTO{
				{
					Id:           uuid.NewString(),
					RepositoryId: uuid.NewString(),
					CommitHash:   "abc123",
					Sdk:          registry.SdkGoProtobuf,
					Status:       registry.SdkGenerationJobStatusPending,
					Attempts:     0,
					MaxAttempts:  5,
					CreatedAt:    time.Now().UTC(),
				},
				{
					Id:           uuid.NewString(),
					RepositoryId: uuid.NewString(),
					CommitHash:   "def456",
					Sdk:          registry.SdkGoConnectRpc,
					Status:       registry.SdkGenerationJobStatusPending,
					Attempts:     0,
					MaxAttempts:  5,
					CreatedAt:    time.Now().UTC(),
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queue.EnqueueSdkGenerationJobs(t.Context(), tt.jobs)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if len(tt.jobs) > 0 {
				var count int
				err = pool.QueryRow(t.Context(), "SELECT COUNT(*) FROM sdk_generation_jobs").Scan(&count)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, count, len(tt.jobs))
			}
		})
	}
}

func TestGetPendingSdkGenerationJobs(t *testing.T) {
	queue, pool, cleanup := setupQueueTestEnvironment(t)
	defer cleanup()

	now := time.Now().UTC()
	jobs := []*registry.SdkGenerationJobDTO{
		{
			Id:           uuid.NewString(),
			RepositoryId: uuid.NewString(),
			CommitHash:   "abc123",
			Sdk:          registry.SdkGoProtobuf,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now,
		},
		{
			Id:           uuid.NewString(),
			RepositoryId: uuid.NewString(),
			CommitHash:   "def456",
			Sdk:          registry.SdkGoConnectRpc,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now.Add(1 * time.Second),
		},
		{
			Id:           uuid.NewString(),
			RepositoryId: uuid.NewString(),
			CommitHash:   "ghi789",
			Sdk:          registry.SdkGoGrpc,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now.Add(2 * time.Second),
		},
	}

	err := queue.EnqueueSdkGenerationJobs(t.Context(), jobs)
	require.NoError(t, err)

	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{
			name:      "get one job",
			limit:     1,
			wantCount: 1,
		},
		{
			name:      "get two jobs",
			limit:     2,
			wantCount: 2,
		},
		{
			name:      "get all jobs",
			limit:     10,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pool.Exec(t.Context(), "UPDATE sdk_generation_jobs SET status = 'pending', attempts = 0")
			require.NoError(t, err)

			fetchedJobs, err := queue.GetPendingSdkGenerationJobs(t.Context(), tt.limit)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCount, len(fetchedJobs))

			for _, job := range fetchedJobs {
				assert.Equal(t, registry.SdkGenerationJobStatusProcessing, job.Status)
				assert.Equal(t, 1, job.Attempts)
				assert.NotNil(t, job.ProcessedAt)
			}
		})
	}
}

func TestGetPendingSdkGenerationJobs_FIFO_Order(t *testing.T) {
	queue, _, cleanup := setupQueueTestEnvironment(t)
	defer cleanup()

	now := time.Now().UTC()
	jobs := []*registry.SdkGenerationJobDTO{
		{
			Id:           "job-1",
			RepositoryId: uuid.NewString(),
			CommitHash:   "abc123",
			Sdk:          registry.SdkGoProtobuf,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now,
		},
		{
			Id:           "job-2",
			RepositoryId: uuid.NewString(),
			CommitHash:   "def456",
			Sdk:          registry.SdkGoConnectRpc,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now.Add(1 * time.Second),
		},
	}

	err := queue.EnqueueSdkGenerationJobs(t.Context(), jobs)
	require.NoError(t, err)

	fetchedJobs, err := queue.GetPendingSdkGenerationJobs(t.Context(), 2)
	require.NoError(t, err)

	assert.Equal(t, 2, len(fetchedJobs))
	assert.Equal(t, "job-1", fetchedJobs[0].Id)
	assert.Equal(t, "job-2", fetchedJobs[1].Id)
}

func TestUpdateSdkGenerationJobStatus(t *testing.T) {
	t.Run("processing to completed", func(t *testing.T) {
		queue, _, cleanup := setupQueueTestEnvironment(t)
		defer cleanup()

		now := time.Now().UTC()
		job := &registry.SdkGenerationJobDTO{
			Id:           uuid.NewString(),
			RepositoryId: uuid.NewString(),
			CommitHash:   "abc123",
			Sdk:          registry.SdkGoProtobuf,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now,
		}

		err := queue.EnqueueSdkGenerationJobs(t.Context(), []*registry.SdkGenerationJobDTO{job})
		require.NoError(t, err)

		jobs, err := queue.GetPendingSdkGenerationJobs(t.Context(), 1)
		require.NoError(t, err)
		require.Len(t, jobs, 1)

		err = queue.UpdateSdkGenerationJobStatus(t.Context(), jobs[0].Id, registry.SdkGenerationJobStatusCompleted, nil)
		assert.NoError(t, err)
	})

	t.Run("processing to failed with error message", func(t *testing.T) {
		queue, _, cleanup := setupQueueTestEnvironment(t)
		defer cleanup()

		now := time.Now().UTC()
		job := &registry.SdkGenerationJobDTO{
			Id:           uuid.NewString(),
			RepositoryId: uuid.NewString(),
			CommitHash:   "abc123",
			Sdk:          registry.SdkGoProtobuf,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now,
		}

		err := queue.EnqueueSdkGenerationJobs(t.Context(), []*registry.SdkGenerationJobDTO{job})
		require.NoError(t, err)

		jobs, err := queue.GetPendingSdkGenerationJobs(t.Context(), 1)
		require.NoError(t, err)
		require.Len(t, jobs, 1)

		errorMsg := "generation failed"
		err = queue.UpdateSdkGenerationJobStatus(t.Context(), jobs[0].Id, registry.SdkGenerationJobStatusFailed, &errorMsg)
		assert.NoError(t, err)
	})

	t.Run("processing to pending (retry)", func(t *testing.T) {
		queue, _, cleanup := setupQueueTestEnvironment(t)
		defer cleanup()

		now := time.Now().UTC()
		job := &registry.SdkGenerationJobDTO{
			Id:           uuid.NewString(),
			RepositoryId: uuid.NewString(),
			CommitHash:   "abc123",
			Sdk:          registry.SdkGoProtobuf,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now,
		}

		err := queue.EnqueueSdkGenerationJobs(t.Context(), []*registry.SdkGenerationJobDTO{job})
		require.NoError(t, err)

		jobs, err := queue.GetPendingSdkGenerationJobs(t.Context(), 1)
		require.NoError(t, err)
		require.Len(t, jobs, 1)

		err = queue.UpdateSdkGenerationJobStatus(t.Context(), jobs[0].Id, registry.SdkGenerationJobStatusPending, nil)
		assert.NoError(t, err)
	})
}

func TestSdkGenerationJobQueue_StartStop(t *testing.T) {
	queue, _, cleanup := setupQueueTestEnvironment(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	mockGen := &mockSdkGenerator{
		generateSDKFunc: func(ctx context.Context, repositoryId, commitHash string, sdk registry.SDK) error {
			return nil
		},
	}

	queue.Start(ctx, mockGen, 10, 100*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	queue.Stop()
}

func TestProcessSdkGenerationJobs_Success(t *testing.T) {
	queue, _, cleanup := setupQueueTestEnvironment(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	generatedSDKs := make(map[string]bool)
	mockGen := &mockSdkGenerator{
		generateSDKFunc: func(ctx context.Context, repositoryId, commitHash string, sdk registry.SDK) error {
			generatedSDKs[repositoryId+"-"+commitHash+"-"+string(sdk)] = true
			return nil
		},
	}

	now := time.Now().UTC()
	jobs := []*registry.SdkGenerationJobDTO{
		{
			Id:           uuid.NewString(),
			RepositoryId: "repo-1",
			CommitHash:   "abc123",
			Sdk:          registry.SdkGoProtobuf,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now,
		},
	}

	err := queue.EnqueueSdkGenerationJobs(ctx, jobs)
	require.NoError(t, err)

	queue.Start(ctx, mockGen, 10, 100*time.Millisecond)
	defer queue.Stop()

	time.Sleep(500 * time.Millisecond)

	assert.True(t, generatedSDKs["repo-1-abc123-GO_PROTOBUF"])

	fetchedJobs, err := queue.GetPendingSdkGenerationJobs(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, len(fetchedJobs))
}

func TestProcessSdkGenerationJobs_Retry(t *testing.T) {
	queue, pool, cleanup := setupQueueTestEnvironment(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	attemptCount := 0
	mockGen := &mockSdkGenerator{
		generateSDKFunc: func(ctx context.Context, repositoryId, commitHash string, sdk registry.SDK) error {
			attemptCount++
			return assert.AnError
		},
	}

	now := time.Now().UTC()
	jobId := uuid.NewString()
	jobs := []*registry.SdkGenerationJobDTO{
		{
			Id:           jobId,
			RepositoryId: "repo-1",
			CommitHash:   "abc123",
			Sdk:          registry.SdkGoProtobuf,
			Status:       registry.SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  3,
			CreatedAt:    now,
		},
	}

	err := queue.EnqueueSdkGenerationJobs(ctx, jobs)
	require.NoError(t, err)

	queue.Start(ctx, mockGen, 10, 100*time.Millisecond)
	defer queue.Stop()

	time.Sleep(1 * time.Second)

	var status registry.SdkGenerationJobStatus
	var errorMsg *string
	err = pool.QueryRow(ctx, "SELECT status, error_message FROM sdk_generation_jobs WHERE id = $1", jobId).Scan(&status, &errorMsg)
	require.NoError(t, err)

	assert.Equal(t, registry.SdkGenerationJobStatusFailed, status)
	assert.NotNil(t, errorMsg)
	assert.GreaterOrEqual(t, attemptCount, 3)
}
