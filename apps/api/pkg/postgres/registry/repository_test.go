package registry

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

	"hasir-api/internal/registry"
	"hasir-api/pkg/proto"
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

func createRepositoriesTable(t *testing.T, connString string) {
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

	sql := `CREATE TABLE repositories (
		id VARCHAR PRIMARY KEY,
		name VARCHAR NOT NULL,
		created_by VARCHAR NOT NULL,
		organization_id VARCHAR,
		path VARCHAR NOT NULL,
		visibility visibility NOT NULL DEFAULT 'private',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	)`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createTestRepository(t *testing.T, name string) *registry.RepositoryDTO {
	t.Helper()
	now := time.Now().UTC()
	return &registry.RepositoryDTO{
		Id:         uuid.NewString(),
		Name:       name,
		Path:       "/test/path/" + name,
		CreatedBy:  uuid.NewString(),
		Visibility: proto.VisibilityPrivate,
		CreatedAt:  now,
		UpdatedAt:  &now,
	}
}

func createRepositoriesAndMembersTables(t *testing.T, connString string) {
	t.Helper()

	createRepositoriesTable(t, connString)

	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `CREATE TABLE organization_members (
		id VARCHAR PRIMARY KEY,
		organization_id VARCHAR NOT NULL,
		user_id VARCHAR NOT NULL,
		role VARCHAR NOT NULL,
		joined_at TIMESTAMP NOT NULL
	)`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func TestPgRepository_CreateRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testRepo := createTestRepository(t, "test-repo-"+uuid.NewString())

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbId, dbName, dbPath string
		err = conn.QueryRow(t.Context(),
			"SELECT id, name, path FROM repositories WHERE id = $1", testRepo.Id).
			Scan(&dbId, &dbName, &dbPath)
		require.NoError(t, err)

		assert.Equal(t, testRepo.Id, dbId)
		assert.Equal(t, testRepo.Name, dbName)
		assert.Equal(t, testRepo.Path, dbPath)
	})

	t.Run("verify all fields are stored correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testRepo := createTestRepository(t, "full-fields-repo-"+uuid.NewString())

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbId, dbName, dbOwnerId, dbPath, dbVisibility string
		var dbCreatedAt, dbUpdatedAt time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT id, name, created_by, path, visibility, created_at, updated_at FROM repositories WHERE id = $1", testRepo.Id).
			Scan(&dbId, &dbName, &dbOwnerId, &dbPath, &dbVisibility, &dbCreatedAt, &dbUpdatedAt)
		require.NoError(t, err)

		assert.Equal(t, testRepo.Id, dbId)
		assert.Equal(t, testRepo.Name, dbName)
		assert.Equal(t, testRepo.CreatedBy, dbOwnerId)
		assert.Equal(t, testRepo.Path, dbPath)
		assert.Equal(t, string(testRepo.Visibility), dbVisibility)
		assert.WithinDuration(t, time.Now().UTC(), dbCreatedAt, 5*time.Second)
		assert.WithinDuration(t, time.Now().UTC(), dbUpdatedAt, 5*time.Second)
	})
}

func TestPgRepository_GetRepositoryByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testRepo := createTestRepository(t, "get-by-name-"+uuid.NewString())
		testRepo.Visibility = proto.VisibilityPublic

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		found, err := repo.GetRepositoryByName(t.Context(), testRepo.Name)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, testRepo.Id, found.Id)
		assert.Equal(t, testRepo.Name, found.Name)
		assert.Equal(t, testRepo.Path, found.Path)
		assert.Equal(t, testRepo.Visibility, found.Visibility)
	})

	t.Run("not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		_, err = repo.GetRepositoryByName(t.Context(), "nonexistent-repo-"+uuid.NewString())
		require.ErrorIs(t, err, registry.ErrRepositoryNotFound)
	})

	t.Run("deleted repository not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testRepo := createTestRepository(t, "deleted-repo-"+uuid.NewString())

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE repositories SET deleted_at = NOW() WHERE id = $1", testRepo.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = repo.GetRepositoryByName(t.Context(), testRepo.Name)
		require.ErrorIs(t, err, registry.ErrRepositoryNotFound)
	})

	t.Run("verify all fields are retrieved correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testRepo := createTestRepository(t, "full-fields-repo-"+uuid.NewString())
		testRepo.Visibility = proto.VisibilityPublic

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		found, err := repo.GetRepositoryByName(t.Context(), testRepo.Name)
		require.NoError(t, err)
		require.NotNil(t, found)

		assert.Equal(t, testRepo.Id, found.Id)
		assert.Equal(t, testRepo.Name, found.Name)
		assert.Equal(t, testRepo.CreatedBy, found.CreatedBy)
		assert.Equal(t, testRepo.Path, found.Path)
		assert.Equal(t, testRepo.Visibility, found.Visibility)
		assert.WithinDuration(t, time.Now().UTC(), found.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, time.Now().UTC(), *found.UpdatedAt, 5*time.Second)
	})
}

func TestPgRepository_GetRepositories(t *testing.T) {
	t.Run("success with multiple repositories", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testRepo1 := createTestRepository(t, "list-repo-1-"+uuid.NewString())
		testRepo2 := createTestRepository(t, "list-repo-2-"+uuid.NewString())
		testRepo2.Visibility = proto.VisibilityPublic

		err = repo.CreateRepository(t.Context(), testRepo1)
		require.NoError(t, err)

		err = repo.CreateRepository(t.Context(), testRepo2)
		require.NoError(t, err)

		repos, err := repo.GetRepositories(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Len(t, *repos, 2)

		foundRepo1 := false
		foundRepo2 := false
		for _, r := range *repos {
			if r.Id == testRepo1.Id {
				foundRepo1 = true
				assert.Equal(t, testRepo1.Name, r.Name)
				assert.Equal(t, testRepo1.Visibility, r.Visibility)
			}
			if r.Id == testRepo2.Id {
				foundRepo2 = true
				assert.Equal(t, testRepo2.Name, r.Name)
				assert.Equal(t, testRepo2.Visibility, r.Visibility)
			}
		}
		assert.True(t, foundRepo1, "testRepo1 should be found in results")
		assert.True(t, foundRepo2, "testRepo2 should be found in results")
	})

	t.Run("success with empty result", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		repos, err := repo.GetRepositories(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Empty(t, *repos)
	})

	t.Run("excludes deleted repositories", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		activeRepo := createTestRepository(t, "active-repo-"+uuid.NewString())
		deletedRepo := createTestRepository(t, "deleted-repo-"+uuid.NewString())

		err = repo.CreateRepository(t.Context(), activeRepo)
		require.NoError(t, err)

		err = repo.CreateRepository(t.Context(), deletedRepo)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE repositories SET deleted_at = NOW() WHERE id = $1", deletedRepo.Id)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		repos, err := repo.GetRepositories(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Len(t, *repos, 1)
		assert.Equal(t, activeRepo.Id, (*repos)[0].Id)
	})

	t.Run("returns repositories ordered by created_at desc", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		olderRepo := createTestRepository(t, "older-repo-"+uuid.NewString())
		newerRepo := createTestRepository(t, "newer-repo-"+uuid.NewString())

		err = repo.CreateRepository(t.Context(), olderRepo)
		require.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		err = repo.CreateRepository(t.Context(), newerRepo)
		require.NoError(t, err)

		repos, err := repo.GetRepositories(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Len(t, *repos, 2)

		assert.Equal(t, newerRepo.Id, (*repos)[0].Id)
		assert.Equal(t, olderRepo.Id, (*repos)[1].Id)
	})

	t.Run("verify all fields are retrieved correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesTable(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		testRepo := createTestRepository(t, "full-fields-repo-"+uuid.NewString())
		testRepo.Visibility = proto.VisibilityPublic

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		repos, err := repo.GetRepositories(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Len(t, *repos, 1)

		found := (*repos)[0]
		assert.Equal(t, testRepo.Id, found.Id)
		assert.Equal(t, testRepo.Name, found.Name)
		assert.Equal(t, testRepo.CreatedBy, found.CreatedBy)
		assert.Equal(t, testRepo.Path, found.Path)
		assert.Equal(t, testRepo.Visibility, found.Visibility)
		assert.WithinDuration(t, time.Now().UTC(), found.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, time.Now().UTC(), *found.UpdatedAt, 5*time.Second)
	})
}

func TestPgRepository_GetRepositoriesByUser(t *testing.T) {
	t.Run("returns repositories for a specific user via organization membership", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createRepositoriesAndMembersTables(t, connString)

		repo, pool := setupTestRepository(t, connString)
		defer pool.Close()

		userID := uuid.NewString()
		orgID1 := uuid.NewString()
		orgID2 := uuid.NewString()
		otherOrgID := uuid.NewString()

		repo1 := createTestRepository(t, "user-repo-1-"+uuid.NewString())
		repo1.OrganizationId = orgID1
		repo2 := createTestRepository(t, "user-repo-2-"+uuid.NewString())
		repo2.OrganizationId = orgID2
		repoOther := createTestRepository(t, "other-repo-"+uuid.NewString())
		repoOther.OrganizationId = otherOrgID

		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repoOther)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = conn.Exec(t.Context(),
			`INSERT INTO organization_members (id, organization_id, user_id, role, joined_at)
			 VALUES ($1, $2, $3, 'owner', NOW()),
			        ($4, $5, $3, 'author', NOW())`,
			uuid.NewString(), orgID1, userID,
			uuid.NewString(), orgID2,
		)
		require.NoError(t, err)

		repos, err := repo.GetRepositoriesByUser(t.Context(), userID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Len(t, *repos, 2)
	})
}
