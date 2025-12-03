package organization

import (
	"hasir-api/pkg/proto"
	"testing"
	"time"

	"connectrpc.com/connect"
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

func createTestOrganization(t *testing.T, name string, visibility proto.Visibility) *OrganizationDTO {
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

		testOrg := createTestOrganization(t, "test-org-"+uuid.NewString(), proto.VisibilityPrivate)

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

		testOrg := createTestOrganization(t, "public-org-"+uuid.NewString(), proto.VisibilityPublic)

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
		testOrg1 := createTestOrganization(t, orgName, proto.VisibilityPrivate)
		testOrg2 := createTestOrganization(t, orgName, proto.VisibilityPrivate)

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

		testOrg := createTestOrganization(t, "full-fields-org-"+uuid.NewString(), proto.VisibilityPublic)

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

		testOrg := createTestOrganization(t, "get-by-name-"+uuid.NewString(), proto.VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		found, err := repo.GetOrganizationByName(t.Context(), testOrg.Name)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, testOrg.Id, found.Id)
		assert.Equal(t, testOrg.Name, found.Name)
		assert.Equal(t, proto.VisibilityPrivate, found.Visibility)
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

		testOrg := createTestOrganization(t, "deleted-org-"+uuid.NewString(), proto.VisibilityPrivate)

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

		testOrg := createTestOrganization(t, "full-fields-org-"+uuid.NewString(), proto.VisibilityPublic)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		found, err := repo.GetOrganizationByName(t.Context(), testOrg.Name)
		require.NoError(t, err)
		require.NotNil(t, found)

		assert.Equal(t, testOrg.Id, found.Id)
		assert.Equal(t, testOrg.Name, found.Name)
		assert.Equal(t, proto.VisibilityPublic, found.Visibility)
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

		testOrg1 := createTestOrganization(t, "list-org-1-"+uuid.NewString(), proto.VisibilityPrivate)
		testOrg2 := createTestOrganization(t, "list-org-2-"+uuid.NewString(), proto.VisibilityPublic)

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
				assert.Equal(t, proto.VisibilityPrivate, o.Visibility)
			}
			if o.Id == testOrg2.Id {
				foundOrg2 = true
				assert.Equal(t, testOrg2.Name, o.Name)
				assert.Equal(t, proto.VisibilityPublic, o.Visibility)
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

		activeOrg := createTestOrganization(t, "active-org-"+uuid.NewString(), proto.VisibilityPrivate)
		deletedOrg := createTestOrganization(t, "deleted-org-"+uuid.NewString(), proto.VisibilityPrivate)

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

		olderOrg := createTestOrganization(t, "older-org-"+uuid.NewString(), proto.VisibilityPrivate)
		newerOrg := createTestOrganization(t, "newer-org-"+uuid.NewString(), proto.VisibilityPrivate)

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

		testOrg := createTestOrganization(t, "full-fields-org-"+uuid.NewString(), proto.VisibilityPublic)

		err = repo.CreateOrganization(t.Context(), testOrg)
		require.NoError(t, err)

		orgs, err := repo.GetOrganizations(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, orgs)
		require.Len(t, *orgs, 1)

		found := (*orgs)[0]
		assert.Equal(t, testOrg.Id, found.Id)
		assert.Equal(t, testOrg.Name, found.Name)
		assert.Equal(t, proto.VisibilityPublic, found.Visibility)
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

func createOrganizationInvitesTable(t *testing.T, connString string) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	createUsersTable(t, connString)

	sql := `CREATE TABLE IF NOT EXISTS organization_invites (
		id VARCHAR(36) PRIMARY KEY,
		organization_id VARCHAR(36) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
		email VARCHAR(255) NOT NULL,
		token VARCHAR(64) NOT NULL UNIQUE,
		invited_by VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role VARCHAR(20) NOT NULL DEFAULT 'author',
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		accepted_at TIMESTAMP WITH TIME ZONE,
		CONSTRAINT chk_invite_status CHECK (status IN ('pending', 'accepted', 'expired', 'cancelled')),
		CONSTRAINT chk_invite_role CHECK (role IN ('reader', 'author', 'owner'))
	)`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createTestInvite(t *testing.T, orgId, email, token, invitedBy string, role MemberRole) *OrganizationInviteDTO {
	t.Helper()
	now := time.Now().UTC()
	return &OrganizationInviteDTO{
		Id:             uuid.NewString(),
		OrganizationId: orgId,
		Email:          email,
		Token:          token,
		InvitedBy:      invitedBy,
		Role:           role,
		Status:         InviteStatusPending,
		CreatedAt:      now,
		ExpiresAt:      now.AddDate(0, 0, 7),
	}
}

func TestPgRepository_CreateInvitesAndEnqueueEmailJobs(t *testing.T) {
	t.Run("success with single invite and job", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createOrganizationInvitesTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "inviter", "inviter@example.com")
		insertTestUser(t, connString, user)

		invite := createTestInvite(t, org.Id, "invitee@example.com", "token123", user.Id, MemberRoleAuthor)
		job := createTestEmailJob(t, invite.Id, org.Id, "invitee@example.com", org.Name, invite.Token)

		err = repo.CreateInvitesAndEnqueueEmailJobs(t.Context(), []*OrganizationInviteDTO{invite}, []*EmailJobDTO{job})
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var inviteId, inviteEmail, inviteToken string
		err = conn.QueryRow(t.Context(),
			"SELECT id, email, token FROM organization_invites WHERE id = $1", invite.Id).
			Scan(&inviteId, &inviteEmail, &inviteToken)
		require.NoError(t, err)
		assert.Equal(t, invite.Id, inviteId)
		assert.Equal(t, invite.Email, inviteEmail)
		assert.Equal(t, invite.Token, inviteToken)

		var jobId, jobEmail, jobStatus string
		err = conn.QueryRow(t.Context(),
			"SELECT id, email, status FROM email_jobs WHERE id = $1", job.Id).
			Scan(&jobId, &jobEmail, &jobStatus)
		require.NoError(t, err)
		assert.Equal(t, job.Id, jobId)
		assert.Equal(t, job.Email, jobEmail)
		assert.Equal(t, "pending", jobStatus)
	})

	t.Run("success with multiple invites and jobs", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createOrganizationInvitesTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "inviter", "inviter@example.com")
		insertTestUser(t, connString, user)

		invites := []*OrganizationInviteDTO{
			createTestInvite(t, org.Id, "user1@example.com", "token1", user.Id, MemberRoleAuthor),
			createTestInvite(t, org.Id, "user2@example.com", "token2", user.Id, MemberRoleReader),
		}

		jobs := []*EmailJobDTO{
			createTestEmailJob(t, invites[0].Id, org.Id, "user1@example.com", org.Name, invites[0].Token),
			createTestEmailJob(t, invites[1].Id, org.Id, "user2@example.com", org.Name, invites[1].Token),
		}

		err = repo.CreateInvitesAndEnqueueEmailJobs(t.Context(), invites, jobs)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var inviteCount int
		err = conn.QueryRow(t.Context(), "SELECT COUNT(*) FROM organization_invites").Scan(&inviteCount)
		require.NoError(t, err)
		assert.Equal(t, 2, inviteCount)

		var jobCount int
		err = conn.QueryRow(t.Context(), "SELECT COUNT(*) FROM email_jobs").Scan(&jobCount)
		require.NoError(t, err)
		assert.Equal(t, 2, jobCount)
	})

	t.Run("error when invites and jobs length mismatch", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createOrganizationInvitesTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "inviter", "inviter@example.com")
		insertTestUser(t, connString, user)

		invite := createTestInvite(t, org.Id, "user@example.com", "token1", user.Id, MemberRoleAuthor)
		job1 := createTestEmailJob(t, invite.Id, org.Id, "user@example.com", org.Name, invite.Token)
		job2 := createTestEmailJob(t, invite.Id, org.Id, "user@example.com", org.Name, invite.Token)

		err = repo.CreateInvitesAndEnqueueEmailJobs(t.Context(), []*OrganizationInviteDTO{invite}, []*EmailJobDTO{job1, job2})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must have the same length")
	})

	t.Run("error when invite token already exists", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createOrganizationInvitesTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "inviter", "inviter@example.com")
		insertTestUser(t, connString, user)

		invite1 := createTestInvite(t, org.Id, "user1@example.com", "duplicate-token", user.Id, MemberRoleAuthor)
		job1 := createTestEmailJob(t, invite1.Id, org.Id, "user1@example.com", org.Name, invite1.Token)

		err = repo.CreateInvitesAndEnqueueEmailJobs(t.Context(), []*OrganizationInviteDTO{invite1}, []*EmailJobDTO{job1})
		require.NoError(t, err)

		invite2 := createTestInvite(t, org.Id, "user2@example.com", "duplicate-token", user.Id, MemberRoleAuthor)
		job2 := createTestEmailJob(t, invite2.Id, org.Id, "user2@example.com", org.Name, invite2.Token)

		err = repo.CreateInvitesAndEnqueueEmailJobs(t.Context(), []*OrganizationInviteDTO{invite2}, []*EmailJobDTO{job2})
		require.Error(t, err)
		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeAlreadyExists, connectErr.Code())
	})

	t.Run("success with empty invites", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createOrganizationInvitesTable(t, connString)
		createEmailJobsTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		err = repo.CreateInvitesAndEnqueueEmailJobs(t.Context(), []*OrganizationInviteDTO{}, []*EmailJobDTO{})
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

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		for _, job := range []*EmailJobDTO{pendingJob1, pendingJob2, completedJob} {
			_, err = conn.Exec(t.Context(), `
				INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, 0, 3, NOW())
			`, job.Id, job.InviteId, job.OrganizationId, job.Email, job.OrganizationName, job.InviteToken, job.Status)
			require.NoError(t, err)
		}

		_, err = conn.Exec(t.Context(), "UPDATE email_jobs SET status = 'completed' WHERE id = $1", completedJob.Id)
		require.NoError(t, err)

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

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		for _, job := range jobs {
			_, err = conn.Exec(t.Context(), `
				INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, 3, NOW())
			`, job.Id, job.InviteId, job.OrganizationId, job.Email, job.OrganizationName, job.InviteToken)
			require.NoError(t, err)
		}

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

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = conn.Exec(t.Context(), `
			INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, 3, NOW())
		`, job.Id, job.InviteId, job.OrganizationId, job.Email, job.OrganizationName, job.InviteToken)
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

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = conn.Exec(t.Context(), `
			INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'processing', 0, 3, NOW())
		`, job.Id, job.InviteId, job.OrganizationId, job.Email, job.OrganizationName, job.InviteToken)
		require.NoError(t, err)

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

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = conn.Exec(t.Context(), `
			INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'processing', 0, 3, NOW())
		`, job.Id, job.InviteId, job.OrganizationId, job.Email, job.OrganizationName, job.InviteToken)
		require.NoError(t, err)

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

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = conn.Exec(t.Context(), `
			INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'processing', 0, 3, NOW())
		`, job.Id, job.InviteId, job.OrganizationId, job.Email, job.OrganizationName, job.InviteToken)
		require.NoError(t, err)

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

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = conn.Exec(t.Context(), `
			INSERT INTO email_jobs (id, invite_id, organization_id, email, organization_name, invite_token, status, attempts, max_attempts, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, 3, NOW())
		`, job.Id, job.InviteId, job.OrganizationId, job.Email, job.OrganizationName, job.InviteToken)
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

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), proto.VisibilityPrivate)
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

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), proto.VisibilityPrivate)
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

		activeOrg := createTestOrganization(t, "active-org-"+uuid.NewString(), proto.VisibilityPrivate)
		deletedOrg := createTestOrganization(t, "deleted-org-"+uuid.NewString(), proto.VisibilityPrivate)

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

		org1 := createTestOrganization(t, "org-1-"+uuid.NewString(), proto.VisibilityPrivate)
		org2 := createTestOrganization(t, "org-2-"+uuid.NewString(), proto.VisibilityPrivate)

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

func createUsersTable(t *testing.T, connString string) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		username VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	)`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createOrganizationMembersTable(t *testing.T, connString string) {
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

	sql := `CREATE TABLE IF NOT EXISTS organization_members (
		id VARCHAR(36) PRIMARY KEY,
		organization_id VARCHAR(36) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role VARCHAR(20) NOT NULL DEFAULT 'reader',
		joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		CONSTRAINT chk_member_role CHECK (role IN ('owner', 'author', 'reader')),
		CONSTRAINT uq_organization_member UNIQUE (organization_id, user_id)
	)`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createOrganizationMembersView(t *testing.T, connString string) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `CREATE OR REPLACE VIEW organization_members_view AS
		SELECT
			om.id,
			om.organization_id,
			om.user_id,
			om.role,
			om.joined_at,
			u.username,
			u.email
		FROM organization_members om
		INNER JOIN users u ON om.user_id = u.id
		WHERE u.deleted_at IS NULL`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createTestUser(t *testing.T, username, email string) *struct {
	Id        string
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
} {
	t.Helper()
	now := time.Now().UTC()
	return &struct {
		Id        string
		Username  string
		Email     string
		Password  string
		CreatedAt time.Time
	}{
		Id:        uuid.NewString(),
		Username:  username,
		Email:     email,
		Password:  "hashed-password",
		CreatedAt: now,
	}
}

func insertTestUser(t *testing.T, connString string, user *struct {
	Id        string
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
}) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		_ = conn.Close(t.Context())
	}()

	_, err = conn.Exec(t.Context(),
		"INSERT INTO users (id, username, email, password, created_at) VALUES ($1, $2, $3, $4, $5)",
		user.Id, user.Username, user.Email, user.Password, user.CreatedAt)
	require.NoError(t, err)
}

func createTestMember(t *testing.T, organizationId, userId string, role MemberRole) *OrganizationMemberDTO {
	t.Helper()
	now := time.Now().UTC()
	return &OrganizationMemberDTO{
		Id:             uuid.NewString(),
		OrganizationId: organizationId,
		UserId:         userId,
		Role:           role,
		JoinedAt:       now,
	}
}

func insertTestMember(t *testing.T, connString string, member *OrganizationMemberDTO) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		_ = conn.Close(t.Context())
	}()

	_, err = conn.Exec(t.Context(),
		"INSERT INTO organization_members (id, organization_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4, $5)",
		member.Id, member.OrganizationId, member.UserId, member.Role, member.JoinedAt)
	require.NoError(t, err)
}

func TestPgRepository_GetMembers(t *testing.T) {
	t.Run("success with multiple members", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "test-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user1 := createTestUser(t, "user1", "user1@example.com")
		user2 := createTestUser(t, "user2", "user2@example.com")
		user3 := createTestUser(t, "user3", "user3@example.com")

		insertTestUser(t, connString, user1)
		insertTestUser(t, connString, user2)
		insertTestUser(t, connString, user3)

		member1 := createTestMember(t, org.Id, user1.Id, MemberRoleOwner)
		member2 := createTestMember(t, org.Id, user2.Id, MemberRoleAuthor)
		member3 := createTestMember(t, org.Id, user3.Id, MemberRoleReader)

		insertTestMember(t, connString, member1)
		time.Sleep(10 * time.Millisecond)
		insertTestMember(t, connString, member2)
		time.Sleep(10 * time.Millisecond)
		insertTestMember(t, connString, member3)

		members, usernames, emails, err := repo.GetMembers(t.Context(), org.Id)
		require.NoError(t, err)
		require.Len(t, members, 3)
		require.Len(t, usernames, 3)
		require.Len(t, emails, 3)

		assert.Equal(t, member1.Id, members[0].Id)
		assert.Equal(t, user1.Username, usernames[0])
		assert.Equal(t, user1.Email, emails[0])
		assert.Equal(t, MemberRoleOwner, members[0].Role)

		assert.Equal(t, member2.Id, members[1].Id)
		assert.Equal(t, user2.Username, usernames[1])
		assert.Equal(t, user2.Email, emails[1])
		assert.Equal(t, MemberRoleAuthor, members[1].Role)

		assert.Equal(t, member3.Id, members[2].Id)
		assert.Equal(t, user3.Username, usernames[2])
		assert.Equal(t, user3.Email, emails[2])
		assert.Equal(t, MemberRoleReader, members[2].Role)
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
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "empty-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		members, usernames, emails, err := repo.GetMembers(t.Context(), org.Id)
		require.NoError(t, err)
		require.Empty(t, members)
		require.Empty(t, usernames)
		require.Empty(t, emails)
	})

	t.Run("success with single member", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "single-member-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "singleuser", "single@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org.Id, user.Id, MemberRoleOwner)
		insertTestMember(t, connString, member)

		members, usernames, emails, err := repo.GetMembers(t.Context(), org.Id)
		require.NoError(t, err)
		require.Len(t, members, 1)
		require.Len(t, usernames, 1)
		require.Len(t, emails, 1)

		assert.Equal(t, member.Id, members[0].Id)
		assert.Equal(t, member.OrganizationId, members[0].OrganizationId)
		assert.Equal(t, member.UserId, members[0].UserId)
		assert.Equal(t, member.Role, members[0].Role)
		assert.Equal(t, user.Username, usernames[0])
		assert.Equal(t, user.Email, emails[0])
	})

	t.Run("excludes deleted users", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "org-with-deleted-user-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		activeUser := createTestUser(t, "activeuser", "active@example.com")
		deletedUser := createTestUser(t, "deleteduser", "deleted@example.com")

		insertTestUser(t, connString, activeUser)
		insertTestUser(t, connString, deletedUser)

		activeMember := createTestMember(t, org.Id, activeUser.Id, MemberRoleOwner)
		deletedMember := createTestMember(t, org.Id, deletedUser.Id, MemberRoleAuthor)

		insertTestMember(t, connString, activeMember)
		insertTestMember(t, connString, deletedMember)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE users SET deleted_at = NOW() WHERE id = $1", deletedUser.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		members, usernames, emails, err := repo.GetMembers(t.Context(), org.Id)
		require.NoError(t, err)
		require.Len(t, members, 1)
		require.Len(t, usernames, 1)
		require.Len(t, emails, 1)

		assert.Equal(t, activeMember.Id, members[0].Id)
		assert.Equal(t, activeUser.Username, usernames[0])
		assert.Equal(t, activeUser.Email, emails[0])
	})

	t.Run("returns members ordered by joined_at ASC", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "ordered-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user1 := createTestUser(t, "firstuser", "first@example.com")
		user2 := createTestUser(t, "seconduser", "second@example.com")
		user3 := createTestUser(t, "thirduser", "third@example.com")

		insertTestUser(t, connString, user1)
		insertTestUser(t, connString, user2)
		insertTestUser(t, connString, user3)

		firstMember := createTestMember(t, org.Id, user1.Id, MemberRoleOwner)
		insertTestMember(t, connString, firstMember)

		time.Sleep(50 * time.Millisecond)

		secondMember := createTestMember(t, org.Id, user2.Id, MemberRoleAuthor)
		insertTestMember(t, connString, secondMember)

		time.Sleep(50 * time.Millisecond)

		thirdMember := createTestMember(t, org.Id, user3.Id, MemberRoleReader)
		insertTestMember(t, connString, thirdMember)

		members, _, _, err := repo.GetMembers(t.Context(), org.Id)
		require.NoError(t, err)
		require.Len(t, members, 3)

		assert.Equal(t, firstMember.Id, members[0].Id)
		assert.Equal(t, secondMember.Id, members[1].Id)
		assert.Equal(t, thirdMember.Id, members[2].Id)

		assert.True(t, members[0].JoinedAt.Before(members[1].JoinedAt) || members[0].JoinedAt.Equal(members[1].JoinedAt))
		assert.True(t, members[1].JoinedAt.Before(members[2].JoinedAt) || members[1].JoinedAt.Equal(members[2].JoinedAt))
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
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "full-fields-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "fulluser", "full@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org.Id, user.Id, MemberRoleAuthor)
		insertTestMember(t, connString, member)

		members, usernames, emails, err := repo.GetMembers(t.Context(), org.Id)
		require.NoError(t, err)
		require.Len(t, members, 1)

		found := members[0]
		assert.Equal(t, member.Id, found.Id)
		assert.Equal(t, member.OrganizationId, found.OrganizationId)
		assert.Equal(t, member.UserId, found.UserId)
		assert.Equal(t, member.Role, found.Role)
		assert.WithinDuration(t, time.Now().UTC(), found.JoinedAt, 5*time.Second)
		assert.Equal(t, user.Username, usernames[0])
		assert.Equal(t, user.Email, emails[0])
	})

	t.Run("only returns members for specified organization", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org1 := createTestOrganization(t, "org1-"+uuid.NewString(), proto.VisibilityPrivate)
		org2 := createTestOrganization(t, "org2-"+uuid.NewString(), proto.VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), org1)
		require.NoError(t, err)
		err = repo.CreateOrganization(t.Context(), org2)
		require.NoError(t, err)

		user1 := createTestUser(t, "org1user", "org1@example.com")
		user2 := createTestUser(t, "org2user", "org2@example.com")

		insertTestUser(t, connString, user1)
		insertTestUser(t, connString, user2)

		member1 := createTestMember(t, org1.Id, user1.Id, MemberRoleOwner)
		member2 := createTestMember(t, org2.Id, user2.Id, MemberRoleOwner)

		insertTestMember(t, connString, member1)
		insertTestMember(t, connString, member2)

		members, usernames, emails, err := repo.GetMembers(t.Context(), org1.Id)
		require.NoError(t, err)
		require.Len(t, members, 1)

		assert.Equal(t, member1.Id, members[0].Id)
		assert.Equal(t, user1.Username, usernames[0])
		assert.Equal(t, user1.Email, emails[0])
	})

	t.Run("handles all member roles correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "roles-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		ownerUser := createTestUser(t, "owner", "owner@example.com")
		authorUser := createTestUser(t, "author", "author@example.com")
		readerUser := createTestUser(t, "reader", "reader@example.com")

		insertTestUser(t, connString, ownerUser)
		insertTestUser(t, connString, authorUser)
		insertTestUser(t, connString, readerUser)

		ownerMember := createTestMember(t, org.Id, ownerUser.Id, MemberRoleOwner)
		authorMember := createTestMember(t, org.Id, authorUser.Id, MemberRoleAuthor)
		readerMember := createTestMember(t, org.Id, readerUser.Id, MemberRoleReader)

		insertTestMember(t, connString, ownerMember)
		insertTestMember(t, connString, authorMember)
		insertTestMember(t, connString, readerMember)

		members, _, _, err := repo.GetMembers(t.Context(), org.Id)
		require.NoError(t, err)
		require.Len(t, members, 3)

		roleMap := make(map[MemberRole]bool)
		for _, m := range members {
			roleMap[m.Role] = true
		}

		assert.True(t, roleMap[MemberRoleOwner])
		assert.True(t, roleMap[MemberRoleAuthor])
		assert.True(t, roleMap[MemberRoleReader])
	})
}

func TestPgRepository_UpdateMemberRole(t *testing.T) {
	t.Run("success - update reader to author", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "update-role-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "testuser", "test@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org.Id, user.Id, MemberRoleReader)
		insertTestMember(t, connString, member)

		err = repo.UpdateMemberRole(t.Context(), org.Id, user.Id, MemberRoleAuthor)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbRole string
		err = conn.QueryRow(t.Context(),
			"SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user.Id).Scan(&dbRole)
		require.NoError(t, err)
		assert.Equal(t, string(MemberRoleAuthor), dbRole)
	})

	t.Run("success - update author to owner", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "promote-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "authoruser", "author@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org.Id, user.Id, MemberRoleAuthor)
		insertTestMember(t, connString, member)

		err = repo.UpdateMemberRole(t.Context(), org.Id, user.Id, MemberRoleOwner)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbRole string
		err = conn.QueryRow(t.Context(),
			"SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user.Id).Scan(&dbRole)
		require.NoError(t, err)
		assert.Equal(t, string(MemberRoleOwner), dbRole)
	})

	t.Run("success - update owner to reader", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "demote-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "owneruser", "owner@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org.Id, user.Id, MemberRoleOwner)
		insertTestMember(t, connString, member)

		err = repo.UpdateMemberRole(t.Context(), org.Id, user.Id, MemberRoleReader)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbRole string
		err = conn.QueryRow(t.Context(),
			"SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user.Id).Scan(&dbRole)
		require.NoError(t, err)
		assert.Equal(t, string(MemberRoleReader), dbRole)
	})

	t.Run("success - update role for specific member in multi-member organization", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "multi-member-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user1 := createTestUser(t, "user1", "user1@example.com")
		user2 := createTestUser(t, "user2", "user2@example.com")
		user3 := createTestUser(t, "user3", "user3@example.com")

		insertTestUser(t, connString, user1)
		insertTestUser(t, connString, user2)
		insertTestUser(t, connString, user3)

		member1 := createTestMember(t, org.Id, user1.Id, MemberRoleOwner)
		member2 := createTestMember(t, org.Id, user2.Id, MemberRoleAuthor)
		member3 := createTestMember(t, org.Id, user3.Id, MemberRoleReader)

		insertTestMember(t, connString, member1)
		insertTestMember(t, connString, member2)
		insertTestMember(t, connString, member3)

		err = repo.UpdateMemberRole(t.Context(), org.Id, user2.Id, MemberRoleOwner)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var role1, role2, role3 string
		err = conn.QueryRow(t.Context(),
			"SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user1.Id).Scan(&role1)
		require.NoError(t, err)

		err = conn.QueryRow(t.Context(),
			"SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user2.Id).Scan(&role2)
		require.NoError(t, err)

		err = conn.QueryRow(t.Context(),
			"SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user3.Id).Scan(&role3)
		require.NoError(t, err)

		assert.Equal(t, string(MemberRoleOwner), role1)
		assert.Equal(t, string(MemberRoleOwner), role2)
		assert.Equal(t, string(MemberRoleReader), role3)
	})

	t.Run("member not found - non-existent user", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "notfound-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		nonExistentUserId := uuid.NewString()

		err = repo.UpdateMemberRole(t.Context(), org.Id, nonExistentUserId, MemberRoleAuthor)
		require.Error(t, err)
		assert.Equal(t, ErrMemberNotFound, err)
	})

	t.Run("member not found - user not in organization", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org1 := createTestOrganization(t, "org1-"+uuid.NewString(), proto.VisibilityPrivate)
		org2 := createTestOrganization(t, "org2-"+uuid.NewString(), proto.VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), org1)
		require.NoError(t, err)
		err = repo.CreateOrganization(t.Context(), org2)
		require.NoError(t, err)

		user := createTestUser(t, "otherorguser", "other@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org1.Id, user.Id, MemberRoleOwner)
		insertTestMember(t, connString, member)

		err = repo.UpdateMemberRole(t.Context(), org2.Id, user.Id, MemberRoleAuthor)
		require.Error(t, err)
		assert.Equal(t, ErrMemberNotFound, err)
	})

	t.Run("success - update to same role", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "same-role-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "sameuser", "same@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org.Id, user.Id, MemberRoleAuthor)
		insertTestMember(t, connString, member)

		err = repo.UpdateMemberRole(t.Context(), org.Id, user.Id, MemberRoleAuthor)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbRole string
		err = conn.QueryRow(t.Context(),
			"SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user.Id).Scan(&dbRole)
		require.NoError(t, err)
		assert.Equal(t, string(MemberRoleAuthor), dbRole)
	})
}

func TestPgRepository_DeleteMember(t *testing.T) {
	t.Run("success - delete reader member", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "delete-reader-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "readeruser", "reader@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org.Id, user.Id, MemberRoleReader)
		insertTestMember(t, connString, member)

		err = repo.DeleteMember(t.Context(), org.Id, user.Id)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var count int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user.Id).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("success - delete author member", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "delete-author-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user := createTestUser(t, "authoruser", "author@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org.Id, user.Id, MemberRoleAuthor)
		insertTestMember(t, connString, member)

		err = repo.DeleteMember(t.Context(), org.Id, user.Id)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var count int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user.Id).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("success - delete owner member when multiple owners exist", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "delete-owner-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		owner1 := createTestUser(t, "owner1", "owner1@example.com")
		owner2 := createTestUser(t, "owner2", "owner2@example.com")
		insertTestUser(t, connString, owner1)
		insertTestUser(t, connString, owner2)

		member1 := createTestMember(t, org.Id, owner1.Id, MemberRoleOwner)
		member2 := createTestMember(t, org.Id, owner2.Id, MemberRoleOwner)
		insertTestMember(t, connString, member1)
		insertTestMember(t, connString, member2)

		err = repo.DeleteMember(t.Context(), org.Id, owner1.Id)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var count int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, owner1.Id).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		var remainingCount int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM organization_members WHERE organization_id = $1",
			org.Id).Scan(&remainingCount)
		require.NoError(t, err)
		assert.Equal(t, 1, remainingCount)
	})

	t.Run("success - delete member from multi-member organization", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "multi-delete-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user1 := createTestUser(t, "user1", "user1@example.com")
		user2 := createTestUser(t, "user2", "user2@example.com")
		user3 := createTestUser(t, "user3", "user3@example.com")

		insertTestUser(t, connString, user1)
		insertTestUser(t, connString, user2)
		insertTestUser(t, connString, user3)

		member1 := createTestMember(t, org.Id, user1.Id, MemberRoleOwner)
		member2 := createTestMember(t, org.Id, user2.Id, MemberRoleAuthor)
		member3 := createTestMember(t, org.Id, user3.Id, MemberRoleReader)

		insertTestMember(t, connString, member1)
		insertTestMember(t, connString, member2)
		insertTestMember(t, connString, member3)

		err = repo.DeleteMember(t.Context(), org.Id, user2.Id)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var count int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM organization_members WHERE organization_id = $1 AND user_id = $2",
			org.Id, user2.Id).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		var totalCount int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM organization_members WHERE organization_id = $1",
			org.Id).Scan(&totalCount)
		require.NoError(t, err)
		assert.Equal(t, 2, totalCount)
	})

	t.Run("member not found - non-existent user", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "notfound-delete-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		nonExistentUserId := uuid.NewString()

		err = repo.DeleteMember(t.Context(), org.Id, nonExistentUserId)
		require.Error(t, err)
		assert.Equal(t, ErrMemberNotFound, err)
	})

	t.Run("member not found - user not in organization", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org1 := createTestOrganization(t, "org1-"+uuid.NewString(), proto.VisibilityPrivate)
		org2 := createTestOrganization(t, "org2-"+uuid.NewString(), proto.VisibilityPrivate)

		err = repo.CreateOrganization(t.Context(), org1)
		require.NoError(t, err)
		err = repo.CreateOrganization(t.Context(), org2)
		require.NoError(t, err)

		user := createTestUser(t, "otherorguser", "other@example.com")
		insertTestUser(t, connString, user)

		member := createTestMember(t, org1.Id, user.Id, MemberRoleOwner)
		insertTestMember(t, connString, member)

		err = repo.DeleteMember(t.Context(), org2.Id, user.Id)
		require.Error(t, err)
		assert.Equal(t, ErrMemberNotFound, err)
	})

	t.Run("success - verify other members remain after deletion", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createOrganizationsTable(t, connString)
		createUsersTable(t, connString)
		createOrganizationMembersTable(t, connString)
		createOrganizationMembersView(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		org := createTestOrganization(t, "verify-remaining-org-"+uuid.NewString(), proto.VisibilityPrivate)
		err = repo.CreateOrganization(t.Context(), org)
		require.NoError(t, err)

		user1 := createTestUser(t, "keepuser1", "keep1@example.com")
		user2 := createTestUser(t, "deleteuser", "delete@example.com")
		user3 := createTestUser(t, "keepuser2", "keep2@example.com")

		insertTestUser(t, connString, user1)
		insertTestUser(t, connString, user2)
		insertTestUser(t, connString, user3)

		member1 := createTestMember(t, org.Id, user1.Id, MemberRoleOwner)
		member2 := createTestMember(t, org.Id, user2.Id, MemberRoleAuthor)
		member3 := createTestMember(t, org.Id, user3.Id, MemberRoleReader)

		insertTestMember(t, connString, member1)
		insertTestMember(t, connString, member2)
		insertTestMember(t, connString, member3)

		err = repo.DeleteMember(t.Context(), org.Id, user2.Id)
		require.NoError(t, err)

		members, _, _, err := repo.GetMembers(t.Context(), org.Id)
		require.NoError(t, err)
		require.Len(t, members, 2)

		userIds := make(map[string]bool)
		for _, m := range members {
			userIds[m.UserId] = true
		}

		assert.True(t, userIds[user1.Id])
		assert.False(t, userIds[user2.Id])
		assert.True(t, userIds[user3.Id])
	})
}
