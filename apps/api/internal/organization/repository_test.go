package organization

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	pgDb       = "test"
	pgUsername = "test"
	pgPassword = "test"
)

func setupPgContainer(t *testing.T) *postgres.PostgresContainer {
	t.Helper()

	postgresContainer, err := postgres.Run(t.Context(),
		"postgres:16-alpine",
		postgres.WithDatabase(pgDb),
		postgres.WithUsername(pgUsername),
		postgres.WithPassword(pgPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	require.NoError(t, err)

	return postgresContainer
}

func setupTestRepository(t *testing.T, connString string) (*PgRepository, *pgxpool.Pool) {
	t.Helper()

	pool, err := pgxpool.New(t.Context(), connString)
	require.NoError(t, err)

	repo := &PgRepository{
		connectionPool: pool,
		tracer:         noop.NewTracerProvider().Tracer("test"),
	}

	return repo, pool
}

func createOrganizationsTable(t *testing.T, connString string) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	enumSQL := `DO $$ BEGIN
		CREATE TYPE visibility AS ENUM ('private', 'public');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$`
	_, err = conn.Exec(t.Context(), enumSQL)
	require.NoError(t, err)

	sql := `CREATE TABLE organizations (
		id VARCHAR PRIMARY KEY,
		name VARCHAR NOT NULL UNIQUE,
		visibility visibility NOT NULL DEFAULT 'private',
		created_by VARCHAR NOT NULL,
		created_at TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	)`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createTestOrganization(t *testing.T, name string, visibility Visibility) *OrganizationDTO {
	t.Helper()
	now := time.Now().UTC()
	return &OrganizationDTO{
		Id:         uuid.NewString(),
		Name:       name,
		Visibility: visibility,
		CreatedBy:  uuid.NewString(),
		CreatedAt:  now,
	}
}

func TestPgRepository_CreateOrganization(t *testing.T) {
	t.Run("success with private visibility", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testOrg := createTestOrganization(t, "test-org-"+uuid.NewString(), VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbId, dbName, dbVisibility, dbCreatedBy string
		err = conn.QueryRow(t.Context(),
			"SELECT id, name, visibility, created_by FROM organizations WHERE id = $1", testOrg.Id).
			Scan(&dbId, &dbName, &dbVisibility, &dbCreatedBy)
		require.NoError(t, err)

		assert.Equal(t, testOrg.Id, dbId)
		assert.Equal(t, testOrg.Name, dbName)
		assert.Equal(t, "private", dbVisibility)
		assert.Equal(t, testOrg.CreatedBy, dbCreatedBy)
	})

	t.Run("success with public visibility", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testOrg := createTestOrganization(t, "public-org-"+uuid.NewString(), VisibilityPublic)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbVisibility string
		err = conn.QueryRow(t.Context(),
			"SELECT visibility FROM organizations WHERE id = $1", testOrg.Id).
			Scan(&dbVisibility)
		require.NoError(t, err)

		assert.Equal(t, "public", dbVisibility)
	})

	t.Run("duplicate name returns already exists error", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgName := "duplicate-org-" + uuid.NewString()
		testOrg1 := createTestOrganization(t, orgName, VisibilityPrivate)
		testOrg2 := createTestOrganization(t, orgName, VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), testOrg1)
		require.NoError(t, err)

		err = repo.CreateOrganization(t.Context(), testOrg2)
		require.ErrorIs(t, err, ErrOrganizationAlreadyExists)
	})

	t.Run("verify all fields are stored correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testOrg := createTestOrganization(t, "full-fields-org-"+uuid.NewString(), VisibilityPublic)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbId, dbName, dbVisibility, dbCreatedBy string
		var dbCreatedAt time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT id, name, visibility, created_by, created_at FROM organizations WHERE id = $1", testOrg.Id).
			Scan(&dbId, &dbName, &dbVisibility, &dbCreatedBy, &dbCreatedAt)
		require.NoError(t, err)

		assert.Equal(t, testOrg.Id, dbId)
		assert.Equal(t, testOrg.Name, dbName)
		assert.Equal(t, "public", dbVisibility)
		assert.Equal(t, testOrg.CreatedBy, dbCreatedBy)
		assert.WithinDuration(t, time.Now().UTC(), dbCreatedAt, 5*time.Second)
	})
}

func TestPgRepository_GetOrganizationByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testOrg := createTestOrganization(t, "get-by-name-"+uuid.NewString(), VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		found, err := repo.GetOrganizationByName(t.Context(), testOrg.Name)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, testOrg.Id, found.Id)
		assert.Equal(t, testOrg.Name, found.Name)
		assert.Equal(t, VisibilityPrivate, found.Visibility)
	})

	t.Run("not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		_, err = repo.GetOrganizationByName(t.Context(), "nonexistent-org-"+uuid.NewString())
		require.ErrorIs(t, err, ErrOrganizationNotFound)
	})

	t.Run("deleted organization not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testOrg := createTestOrganization(t, "deleted-org-"+uuid.NewString(), VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE organizations SET deleted_at = NOW() WHERE id = $1", testOrg.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = repo.GetOrganizationByName(t.Context(), testOrg.Name)
		require.ErrorIs(t, err, ErrOrganizationNotFound)
	})

	t.Run("verify all fields are retrieved correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testOrg := createTestOrganization(t, "full-fields-org-"+uuid.NewString(), VisibilityPublic)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		found, err := repo.GetOrganizationByName(t.Context(), testOrg.Name)
		require.NoError(t, err)
		require.NotNil(t, found)

		assert.Equal(t, testOrg.Id, found.Id)
		assert.Equal(t, testOrg.Name, found.Name)
		assert.Equal(t, VisibilityPublic, found.Visibility)
		assert.Equal(t, testOrg.CreatedBy, found.CreatedBy)
		assert.WithinDuration(t, time.Now().UTC(), found.CreatedAt, 5*time.Second)
	})
}

func TestPgRepository_GetOrganizations(t *testing.T) {
	t.Run("success with multiple organizations", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testOrg1 := createTestOrganization(t, "list-org-1-"+uuid.NewString(), VisibilityPrivate)
		testOrg2 := createTestOrganization(t, "list-org-2-"+uuid.NewString(), VisibilityPublic)

		err = repo.CreateOrganization(t.Context(), testOrg1)
		require.NoError(t, err)

		err = repo.CreateOrganization(t.Context(), testOrg2)
		require.NoError(t, err)

		orgs, err := repo.GetOrganizations(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, orgs)
		require.Len(t, *orgs, 2)

		foundOrg1 := false
		foundOrg2 := false
		for _, o := range *orgs {
			if o.Id == testOrg1.Id {
				foundOrg1 = true
				assert.Equal(t, testOrg1.Name, o.Name)
				assert.Equal(t, VisibilityPrivate, o.Visibility)
			}
			if o.Id == testOrg2.Id {
				foundOrg2 = true
				assert.Equal(t, testOrg2.Name, o.Name)
				assert.Equal(t, VisibilityPublic, o.Visibility)
			}
		}
		assert.True(t, foundOrg1, "testOrg1 should be found in results")
		assert.True(t, foundOrg2, "testOrg2 should be found in results")
	})

	t.Run("success with empty result", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgs, err := repo.GetOrganizations(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, orgs)
		require.Empty(t, *orgs)
	})

	t.Run("excludes deleted organizations", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		activeOrg := createTestOrganization(t, "active-org-"+uuid.NewString(), VisibilityPrivate)
		deletedOrg := createTestOrganization(t, "deleted-org-"+uuid.NewString(), VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), activeOrg)
		require.NoError(t, err)

		err = repo.CreateOrganization(t.Context(), deletedOrg)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE organizations SET deleted_at = NOW() WHERE id = $1", deletedOrg.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		orgs, err := repo.GetOrganizations(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, orgs)
		require.Len(t, *orgs, 1)
		assert.Equal(t, activeOrg.Id, (*orgs)[0].Id)
	})

	t.Run("returns organizations ordered by created_at desc", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		olderOrg := createTestOrganization(t, "older-org-"+uuid.NewString(), VisibilityPrivate)
		newerOrg := createTestOrganization(t, "newer-org-"+uuid.NewString(), VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), olderOrg)
		require.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		err = repo.CreateOrganization(t.Context(), newerOrg)
		require.NoError(t, err)

		orgs, err := repo.GetOrganizations(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, orgs)
		require.Len(t, *orgs, 2)

		assert.Equal(t, newerOrg.Id, (*orgs)[0].Id)
		assert.Equal(t, olderOrg.Id, (*orgs)[1].Id)
	})

	t.Run("verify all fields are retrieved correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testOrg := createTestOrganization(t, "full-fields-org-"+uuid.NewString(), VisibilityPublic)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		orgs, err := repo.GetOrganizations(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, orgs)
		require.Len(t, *orgs, 1)

		found := (*orgs)[0]
		assert.Equal(t, testOrg.Id, found.Id)
		assert.Equal(t, testOrg.Name, found.Name)
		assert.Equal(t, VisibilityPublic, found.Visibility)
		assert.Equal(t, testOrg.CreatedBy, found.CreatedBy)
		assert.WithinDuration(t, time.Now().UTC(), found.CreatedAt, 5*time.Second)
	})
}

func createEmailJobsTable(t *testing.T, connString string) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `CREATE TABLE IF NOT EXISTS email_jobs (
		id VARCHAR(36) PRIMARY KEY,
		invite_id VARCHAR(36) NOT NULL,
		organization_id VARCHAR(36) NOT NULL,
		email VARCHAR(255) NOT NULL,
		organization_name VARCHAR(255) NOT NULL,
		invite_token VARCHAR(64) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		attempts INT NOT NULL DEFAULT 0,
		max_attempts INT NOT NULL DEFAULT 3,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		processed_at TIMESTAMP WITH TIME ZONE,
		completed_at TIMESTAMP WITH TIME ZONE,
		error_message TEXT,
		CONSTRAINT chk_email_job_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
	)`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createTestEmailJob(t *testing.T, inviteId, orgId, email, orgName, token string) *EmailJobDTO {
	t.Helper()
	now := time.Now().UTC()
	return &EmailJobDTO{
		Id:               uuid.NewString(),
		InviteId:         inviteId,
		OrganizationId:   orgId,
		Email:            email,
		OrganizationName: orgName,
		InviteToken:      token,
		Status:           EmailJobStatusPending,
		Attempts:         0,
		MaxAttempts:      3,
		CreatedAt:        now,
	}
}

func TestPgRepository_EnqueueEmailJobs(t *testing.T) {
	t.Run("success with single job", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()
		inviteId := uuid.NewString()
		job := createTestEmailJob(t, inviteId, orgId, "test@example.com", "Test Org", "token123")

		err = repo.EnqueueEmailJobs(t.Context(), []*EmailJobDTO{job})
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbId, dbEmail, dbStatus string
		var dbAttempts, dbMaxAttempts int
		err = conn.QueryRow(t.Context(),
			"SELECT id, email, status, attempts, max_attempts FROM email_jobs WHERE id = $1", job.Id).
			Scan(&dbId, &dbEmail, &dbStatus, &dbAttempts, &dbMaxAttempts)
		require.NoError(t, err)

		assert.Equal(t, job.Id, dbId)
		assert.Equal(t, job.Email, dbEmail)
		assert.Equal(t, "pending", dbStatus)
		assert.Equal(t, 0, dbAttempts)
		assert.Equal(t, 3, dbMaxAttempts)
	})

	t.Run("success with multiple jobs", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()
		jobs := []*EmailJobDTO{
			createTestEmailJob(t, uuid.NewString(), orgId, "user1@example.com", "Test Org", "token1"),
			createTestEmailJob(t, uuid.NewString(), orgId, "user2@example.com", "Test Org", "token2"),
			createTestEmailJob(t, uuid.NewString(), orgId, "user3@example.com", "Test Org", "token3"),
		}

		err = repo.EnqueueEmailJobs(t.Context(), jobs)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var count int
		err = conn.QueryRow(t.Context(), "SELECT COUNT(*) FROM email_jobs").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("success with empty jobs", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		err = repo.EnqueueEmailJobs(t.Context(), []*EmailJobDTO{})
		require.NoError(t, err)
	})
}

func TestPgRepository_GetPendingEmailJobs(t *testing.T) {
	t.Run("success returns pending jobs", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()

		pendingJob1 := createTestEmailJob(t, uuid.NewString(), orgId, "pending1@example.com", "Test Org", "token1")
		pendingJob2 := createTestEmailJob(t, uuid.NewString(), orgId, "pending2@example.com", "Test Org", "token2")

		completedJob := createTestEmailJob(t, uuid.NewString(), orgId, "completed@example.com", "Test Org", "token3")
		completedJob.Status = EmailJobStatusCompleted

		err = repo.EnqueueEmailJobs(t.Context(), []*EmailJobDTO{pendingJob1, pendingJob2, completedJob})
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE email_jobs SET status = 'completed' WHERE id = $1", completedJob.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		jobs, err := repo.GetPendingEmailJobs(t.Context(), 10)
		require.NoError(t, err)
		require.Len(t, jobs, 2)

		for _, job := range jobs {
			assert.Equal(t, EmailJobStatusProcessing, job.Status)
			assert.NotNil(t, job.ProcessedAt)
			assert.Equal(t, 1, job.Attempts)
		}
	})

	t.Run("success with limit", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()
		jobs := make([]*EmailJobDTO, 5)
		for i := 0; i < 5; i++ {
			jobs[i] = createTestEmailJob(t, uuid.NewString(), orgId,
				"user"+string(rune('0'+i))+"@example.com", "Test Org", "token"+string(rune('0'+i)))
		}

		err = repo.EnqueueEmailJobs(t.Context(), jobs)
		require.NoError(t, err)

		retrievedJobs, err := repo.GetPendingEmailJobs(t.Context(), 2)
		require.NoError(t, err)
		require.Len(t, retrievedJobs, 2)
	})

	t.Run("success with no pending jobs", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		jobs, err := repo.GetPendingEmailJobs(t.Context(), 10)
		require.NoError(t, err)
		require.Empty(t, jobs)
	})

	t.Run("jobs are atomically updated to processing", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()
		job := createTestEmailJob(t, uuid.NewString(), orgId, "test@example.com", "Test Org", "token1")

		err = repo.EnqueueEmailJobs(t.Context(), []*EmailJobDTO{job})
		require.NoError(t, err)

		jobs, err := repo.GetPendingEmailJobs(t.Context(), 10)
		require.NoError(t, err)
		require.Len(t, jobs, 1)

		assert.Equal(t, EmailJobStatusProcessing, jobs[0].Status)
		assert.NotNil(t, jobs[0].ProcessedAt)
		assert.Equal(t, 1, jobs[0].Attempts)

		jobs2, err := repo.GetPendingEmailJobs(t.Context(), 10)
		require.NoError(t, err)
		require.Empty(t, jobs2)
	})
}

func TestPgRepository_UpdateEmailJobStatus(t *testing.T) {
	t.Run("success update to completed", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()
		job := createTestEmailJob(t, uuid.NewString(), orgId, "test@example.com", "Test Org", "token1")
		job.Status = EmailJobStatusProcessing

		err = repo.EnqueueEmailJobs(t.Context(), []*EmailJobDTO{job})
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE email_jobs SET status = 'processing' WHERE id = $1", job.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		err = repo.UpdateEmailJobStatus(t.Context(), job.Id, EmailJobStatusCompleted, nil)
		require.NoError(t, err)

		var dbStatus string
		var dbCompletedAt *time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT status, completed_at FROM email_jobs WHERE id = $1", job.Id).
			Scan(&dbStatus, &dbCompletedAt)
		require.NoError(t, err)

		assert.Equal(t, "completed", dbStatus)
		assert.NotNil(t, dbCompletedAt)
	})

	t.Run("success update to failed", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()
		job := createTestEmailJob(t, uuid.NewString(), orgId, "test@example.com", "Test Org", "token1")

		err = repo.EnqueueEmailJobs(t.Context(), []*EmailJobDTO{job})
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE email_jobs SET status = 'processing' WHERE id = $1", job.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		errorMsg := "SMTP connection failed"
		err = repo.UpdateEmailJobStatus(t.Context(), job.Id, EmailJobStatusFailed, &errorMsg)
		require.NoError(t, err)

		var dbStatus string
		var dbErrorMessage *string
		err = conn.QueryRow(t.Context(),
			"SELECT status, error_message FROM email_jobs WHERE id = $1", job.Id).
			Scan(&dbStatus, &dbErrorMessage)
		require.NoError(t, err)

		assert.Equal(t, "failed", dbStatus)
		require.NotNil(t, dbErrorMessage)
		assert.Equal(t, errorMsg, *dbErrorMessage)
	})

	t.Run("success update to pending for retry", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()
		job := createTestEmailJob(t, uuid.NewString(), orgId, "test@example.com", "Test Org", "token1")

		err = repo.EnqueueEmailJobs(t.Context(), []*EmailJobDTO{job})
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE email_jobs SET status = 'processing' WHERE id = $1", job.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		err = repo.UpdateEmailJobStatus(t.Context(), job.Id, EmailJobStatusPending, nil)
		require.NoError(t, err)

		var dbStatus string
		err = conn.QueryRow(t.Context(),
			"SELECT status FROM email_jobs WHERE id = $1", job.Id).
			Scan(&dbStatus)
		require.NoError(t, err)

		assert.Equal(t, "pending", dbStatus)
	})

	t.Run("fails when job not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		err = repo.UpdateEmailJobStatus(t.Context(), "non-existent-id", EmailJobStatusCompleted, nil)
		require.Error(t, err)
	})

	t.Run("fails when status mismatch (optimistic locking)", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		orgId := uuid.NewString()
		job := createTestEmailJob(t, uuid.NewString(), orgId, "test@example.com", "Test Org", "token1")

		err = repo.EnqueueEmailJobs(t.Context(), []*EmailJobDTO{job})
		require.NoError(t, err)

		err = repo.UpdateEmailJobStatus(t.Context(), job.Id, EmailJobStatusCompleted, nil)
		require.Error(t, err)
	})
}

func TestPgRepository_DeleteOrganization(t *testing.T) {
	t.Run("success - soft delete organization", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		found, err := repo.GetOrganizationById(t.Context(), org.Id)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Nil(t, found.DeletedAt)

		err = repo.DeleteOrganization(t.Context(), org.Id)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var deletedAt *time.Time
		err = conn.QueryRow(t.Context(), "SELECT deleted_at FROM organizations WHERE id = $1", org.Id).Scan(&deletedAt)
		require.NoError(t, err)
		require.NotNil(t, deletedAt)

		_, err = repo.GetOrganizationById(t.Context(), org.Id)
		require.Error(t, err)
		require.Equal(t, ErrOrganizationNotFound, err)
	})

	t.Run("organization not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		nonExistentID := uuid.NewString()
		err = repo.DeleteOrganization(t.Context(), nonExistentID)
		require.Error(t, err)
		require.Equal(t, ErrOrganizationNotFound, err)
	})

	t.Run("cannot delete already deleted organization", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		err = repo.DeleteOrganization(t.Context(), org.Id)
		require.NoError(t, err)

		err = repo.DeleteOrganization(t.Context(), org.Id)
		require.Error(t, err)
		require.Equal(t, ErrOrganizationNotFound, err)
	})

	t.Run("deleted organization excluded from GetOrganizations", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		activeOrg := createTestOrganization(t, "active-org-"+uuid.NewString(), VisibilityPrivate)
		deletedOrg := createTestOrganization(t, "deleted-org-"+uuid.NewString(), VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), activeOrg)
		require.NoError(t, err)

		err = repo.CreateOrganization(t.Context(), deletedOrg)
		require.NoError(t, err)

		err = repo.DeleteOrganization(t.Context(), deletedOrg.Id)
		require.NoError(t, err)

		orgs, err := repo.GetOrganizations(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, orgs)
		require.Len(t, *orgs, 1)
		assert.Equal(t, activeOrg.Id, (*orgs)[0].Id)
	})

	t.Run("deleted organization excluded from count", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org1 := createTestOrganization(t, "org-1-"+uuid.NewString(), VisibilityPrivate)
		org2 := createTestOrganization(t, "org-2-"+uuid.NewString(), VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), org1)
		require.NoError(t, err)

		err = repo.CreateOrganization(t.Context(), org2)
		require.NoError(t, err)

		count, err := repo.GetOrganizationsCount(t.Context())
		require.NoError(t, err)
		require.Equal(t, 2, count)

		err = repo.DeleteOrganization(t.Context(), org1.Id)
		require.NoError(t, err)

		count, err = repo.GetOrganizationsCount(t.Context())
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})
}
