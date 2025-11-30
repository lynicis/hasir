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

	sql := `CREATE TABLE repositories (
		id VARCHAR PRIMARY KEY,
		name VARCHAR NOT NULL,
		owner_id VARCHAR NOT NULL,
		organization_id VARCHAR,
		path VARCHAR NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	)`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createTestRepository(t *testing.T, name string) *RepositoryDTO {
	t.Helper()
	now := time.Now().UTC()
	return &RepositoryDTO{
		Id:        uuid.NewString(),
		Name:      name,
		Path:      "/test/path/" + name,
		OwnerId:   uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: &now,
	}
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

		var dbId, dbName, dbOwnerId, dbPath string
		var dbCreatedAt, dbUpdatedAt time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT id, name, owner_id, path, created_at, updated_at FROM repositories WHERE id = $1", testRepo.Id).
			Scan(&dbId, &dbName, &dbOwnerId, &dbPath, &dbCreatedAt, &dbUpdatedAt)
		require.NoError(t, err)

		assert.Equal(t, testRepo.Id, dbId)
		assert.Equal(t, testRepo.Name, dbName)
		assert.Equal(t, testRepo.OwnerId, dbOwnerId)
		assert.Equal(t, testRepo.Path, dbPath)
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

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		found, err := repo.GetRepositoryByName(t.Context(), testRepo.Name)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, testRepo.Id, found.Id)
		assert.Equal(t, testRepo.Name, found.Name)
		assert.Equal(t, testRepo.Path, found.Path)
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
		require.ErrorIs(t, err, ErrRepositoryNotFound)
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
		require.ErrorIs(t, err, ErrRepositoryNotFound)
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

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		found, err := repo.GetRepositoryByName(t.Context(), testRepo.Name)
		require.NoError(t, err)
		require.NotNil(t, found)

		assert.Equal(t, testRepo.Id, found.Id)
		assert.Equal(t, testRepo.Name, found.Name)
		assert.Equal(t, testRepo.OwnerId, found.OwnerId)
		assert.Equal(t, testRepo.Path, found.Path)
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
			}
			if r.Id == testRepo2.Id {
				foundRepo2 = true
				assert.Equal(t, testRepo2.Name, r.Name)
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

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		repos, err := repo.GetRepositories(t.Context(), 1, 10)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Len(t, *repos, 1)

		found := (*repos)[0]
		assert.Equal(t, testRepo.Id, found.Id)
		assert.Equal(t, testRepo.Name, found.Name)
		assert.Equal(t, testRepo.OwnerId, found.OwnerId)
		assert.Equal(t, testRepo.Path, found.Path)
		assert.WithinDuration(t, time.Now().UTC(), found.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, time.Now().UTC(), *found.UpdatedAt, 5*time.Second)
	})
}
