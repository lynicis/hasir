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

	// Create the visibility enum type
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

		// Verify the organization was created
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

		// Verify the organization was created
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

		// Soft delete the organization
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

		// Soft delete one organization
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

		// Small delay to ensure different created_at timestamps
		time.Sleep(50 * time.Millisecond)

		err = repo.CreateOrganization(t.Context(), newerOrg)
		require.NoError(t, err)

		orgs, err := repo.GetOrganizations(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, orgs)
		require.Len(t, *orgs, 2)

		// Newer org should be first (DESC order)
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
