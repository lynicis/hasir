package registry

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

func setupTestGitRepository(t *testing.T, files map[string]string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "test-git-repo-*")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	repo, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	for filePath, content := range files {
		fullPath := filepath.Join(tempDir, filePath)

		dir := filepath.Dir(fullPath)
		if dir != tempDir {
			err = os.MkdirAll(dir, 0755)
			require.NoError(t, err)
		}

		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)

		_, err = worktree.Add(filePath)
		require.NoError(t, err)
	}

	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	return tempDir
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

func TestPgRepository_DeleteRepository(t *testing.T) {
	t.Run("successfully deletes repository by setting deleted_at", func(t *testing.T) {
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

		testRepo := createTestRepository(t, "test-repo")
		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		err = repo.DeleteRepository(t.Context(), testRepo.Id)
		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var deletedAt *time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT deleted_at FROM repositories WHERE id = $1",
			testRepo.Id,
		).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt, "deleted_at should be set")
	})

	t.Run("returns error when repository not found", func(t *testing.T) {
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

		nonExistentID := uuid.NewString()
		err = repo.DeleteRepository(t.Context(), nonExistentID)
		assert.Error(t, err)
		assert.Equal(t, ErrRepositoryNotFound, err)
	})

	t.Run("returns error when trying to delete already deleted repository", func(t *testing.T) {
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

		testRepo := createTestRepository(t, "test-repo")
		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		err = repo.DeleteRepository(t.Context(), testRepo.Id)
		require.NoError(t, err)

		err = repo.DeleteRepository(t.Context(), testRepo.Id)
		assert.Error(t, err)
		assert.Equal(t, ErrRepositoryNotFound, err)
	})
}

func TestPgRepository_DeleteRepositoriesByOrganizationId(t *testing.T) {
	t.Run("successfully deletes all repositories for organization", func(t *testing.T) {
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

		orgID := uuid.NewString()

		repo1 := createTestRepository(t, "repo1")
		repo1.OrganizationId = orgID
		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)

		repo2 := createTestRepository(t, "repo2")
		repo2.OrganizationId = orgID
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)

		repo3 := createTestRepository(t, "repo3")
		repo3.OrganizationId = orgID
		err = repo.CreateRepository(t.Context(), repo3)
		require.NoError(t, err)

		err = repo.DeleteRepositoriesByOrganizationId(t.Context(), orgID)
		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var count int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM repositories WHERE organization_id = $1 AND deleted_at IS NOT NULL",
			orgID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 3, count, "all 3 repositories should be marked as deleted")

		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM repositories WHERE organization_id = $1 AND deleted_at IS NULL",
			orgID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "no repositories should remain active")
	})

	t.Run("does not delete repositories from other organizations", func(t *testing.T) {
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

		orgID1 := uuid.NewString()
		orgID2 := uuid.NewString()

		repo1 := createTestRepository(t, "org1-repo1")
		repo1.OrganizationId = orgID1
		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)

		repo2 := createTestRepository(t, "org1-repo2")
		repo2.OrganizationId = orgID1
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)

		repo3 := createTestRepository(t, "org2-repo1")
		repo3.OrganizationId = orgID2
		err = repo.CreateRepository(t.Context(), repo3)
		require.NoError(t, err)

		err = repo.DeleteRepositoriesByOrganizationId(t.Context(), orgID1)
		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var deletedCount int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM repositories WHERE organization_id = $1 AND deleted_at IS NOT NULL",
			orgID1,
		).Scan(&deletedCount)
		require.NoError(t, err)
		assert.Equal(t, 2, deletedCount, "both org1 repositories should be deleted")

		var activeCount int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM repositories WHERE organization_id = $1 AND deleted_at IS NULL",
			orgID2,
		).Scan(&activeCount)
		require.NoError(t, err)
		assert.Equal(t, 1, activeCount, "org2 repository should still be active")
	})

	t.Run("handles empty organization gracefully", func(t *testing.T) {
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

		nonExistentOrgID := uuid.NewString()
		err = repo.DeleteRepositoriesByOrganizationId(t.Context(), nonExistentOrgID)
		assert.NoError(t, err, "deleting from empty organization should not error")
	})

	t.Run("does not re-delete already deleted repositories", func(t *testing.T) {
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

		orgID := uuid.NewString()

		repo1 := createTestRepository(t, "repo1")
		repo1.OrganizationId = orgID
		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)

		repo2 := createTestRepository(t, "repo2")
		repo2.OrganizationId = orgID
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)

		err = repo.DeleteRepositoriesByOrganizationId(t.Context(), orgID)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var firstDeletedAt time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT deleted_at FROM repositories WHERE id = $1",
			repo1.Id,
		).Scan(&firstDeletedAt)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		err = repo.DeleteRepositoriesByOrganizationId(t.Context(), orgID)
		assert.NoError(t, err, "second delete should succeed but not change anything")

		var secondDeletedAt time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT deleted_at FROM repositories WHERE id = $1",
			repo1.Id,
		).Scan(&secondDeletedAt)
		require.NoError(t, err)
		assert.Equal(t, firstDeletedAt, secondDeletedAt, "deleted_at should not change on second delete")
	})

	t.Run("batch delete is atomic", func(t *testing.T) {
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

		orgID := uuid.NewString()

		repo1 := createTestRepository(t, "repo1")
		repo1.OrganizationId = orgID
		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)

		repo2 := createTestRepository(t, "repo2")
		repo2.OrganizationId = orgID
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)

		err = repo.DeleteRepositoriesByOrganizationId(t.Context(), orgID)
		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var deletedAt1, deletedAt2 time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT deleted_at FROM repositories WHERE id = $1",
			repo1.Id,
		).Scan(&deletedAt1)
		require.NoError(t, err)

		err = conn.QueryRow(t.Context(),
			"SELECT deleted_at FROM repositories WHERE id = $1",
			repo2.Id,
		).Scan(&deletedAt2)
		require.NoError(t, err)

		timeDiff := deletedAt1.Sub(deletedAt2)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		assert.Less(t, timeDiff, 1*time.Second, "all repositories should be deleted with same timestamp (atomic operation)")
	})
}

func TestPgRepository_GetFilePreview(t *testing.T) {
	t.Run("successfully retrieves file content", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		content := "# Test Repository\n\nThis is a test file."
		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md": content,
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "README.md")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, content, response.Content)
		assert.Equal(t, "text/markdown", response.MimeType)
		assert.Equal(t, int64(len(content)), response.Size)
	})

	t.Run("returns correct mime type for different file extensions", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		tests := []struct {
			fileName     string
			content      string
			expectedMime string
		}{
			{"file.txt", "plain text", "text/plain"},
			{"config.json", `{"key": "value"}`, "application/json"},
			{"config.yaml", "key: value", "application/yaml"},
			{"config.yml", "key: value", "application/yaml"},
			{"style.css", "body { margin: 0; }", "text/css"},
			{"script.js", "console.log('test');", "application/javascript"},
			{"app.ts", "const x: string = 'test';", "application/typescript"},
			{"main.go", "package main", "text/x-go"},
			{"script.py", "print('hello')", "text/x-python"},
			{"Main.java", "public class Main {}", "text/x-java"},
			{"schema.proto", "syntax = \"proto3\";", "text/x-protobuf"},
			{"query.sql", "SELECT * FROM users;", "application/sql"},
			{"index.html", "<html></html>", "text/html"},
		}

		for _, tt := range tests {
			testRepoPath := setupTestGitRepository(t, map[string]string{
				tt.fileName: tt.content,
			})

			response, err := repo.GetFilePreview(t.Context(), testRepoPath, tt.fileName)
			require.NoError(t, err, "failed for file: %s", tt.fileName)
			assert.Equal(t, tt.expectedMime, response.MimeType, "incorrect mime type for: %s", tt.fileName)
			assert.Equal(t, tt.content, response.Content, "incorrect content for: %s", tt.fileName)
		}
	})

	t.Run("returns correct size for file", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		content := "This is a test file with specific length."
		testRepoPath := setupTestGitRepository(t, map[string]string{
			"test.txt": content,
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "test.txt")
		require.NoError(t, err)
		assert.Equal(t, int64(len(content)), response.Size)
	})

	t.Run("handles file in subdirectory", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"src/main.go": "package main\n\nfunc main() {}",
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "src/main.go")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "package main\n\nfunc main() {}", response.Content)
		assert.Equal(t, "text/x-go", response.MimeType)
	})

	t.Run("handles deeply nested file", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"src/internal/handlers/auth/login.go": "package auth",
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "src/internal/handlers/auth/login.go")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "package auth", response.Content)
		assert.Equal(t, "text/x-go", response.MimeType)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md": "# Test",
		})

		_, err := repo.GetFilePreview(t.Context(), testRepoPath, "non-existent.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
	})

	t.Run("returns error for invalid repository path", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		_, err := repo.GetFilePreview(t.Context(), "/invalid/path/to/repo", "README.md")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open git repository")
	})

	t.Run("handles empty file", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"empty.txt": "",
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "empty.txt")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "", response.Content)
		assert.Equal(t, int64(0), response.Size)
		assert.Equal(t, "text/plain", response.MimeType)
	})

	t.Run("handles file without extension", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"Makefile": "all:\n\tgo build",
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "Makefile")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "all:\n\tgo build", response.Content)
		assert.Equal(t, "text/plain", response.MimeType)
	})

	t.Run("handles dockerfile extension", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"app.dockerfile": "FROM node:18",
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "app.dockerfile")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "text/x-dockerfile", response.MimeType)
		assert.Equal(t, "FROM node:18", response.Content)
	})

	t.Run("defaults to text/plain for unknown extension", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"file.unknown": "some content",
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "file.unknown")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "text/plain", response.MimeType)
	})

	t.Run("detects binary file", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		binaryContent := string([]byte{0x00, 0x01, 0x02, 0x03})
		testRepoPath := setupTestGitRepository(t, map[string]string{
			"binary.bin": binaryContent,
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "binary.bin")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "application/octet-stream", response.MimeType)
	})

	t.Run("handles large file content", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		largeContent := ""
		for i := 0; i < 1000; i++ {
			largeContent += "This is line " + string(rune(i)) + "\n"
		}

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"large.txt": largeContent,
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "large.txt")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, largeContent, response.Content)
		assert.Equal(t, int64(len(largeContent)), response.Size)
	})

	t.Run("handles multiple files and retrieves correct one", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"file1.txt": "content1",
			"file2.txt": "content2",
			"file3.txt": "content3",
		})

		response, err := repo.GetFilePreview(t.Context(), testRepoPath, "file2.txt")
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "content2", response.Content)
	})

	t.Run("returns error when file is actually a directory", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"src/main.go": "package main",
		})

		_, err := repo.GetFilePreview(t.Context(), testRepoPath, "src")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
	})
}

func TestPgRepository_GetRepositoriesByOrganizationId(t *testing.T) {
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

		orgID := uuid.NewString()
		repo1 := createTestRepository(t, "org-repo-1-"+uuid.NewString())
		repo1.OrganizationId = orgID
		repo2 := createTestRepository(t, "org-repo-2-"+uuid.NewString())
		repo2.OrganizationId = orgID

		otherOrgID := uuid.NewString()
		repoOther := createTestRepository(t, "other-org-repo-"+uuid.NewString())
		repoOther.OrganizationId = otherOrgID

		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repoOther)
		require.NoError(t, err)

		repos, err := repo.GetRepositoriesByOrganizationId(t.Context(), orgID)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Len(t, *repos, 2)
	})

	t.Run("returns empty list for non-existent organization", func(t *testing.T) {
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

		repos, err := repo.GetRepositoriesByOrganizationId(t.Context(), uuid.NewString())
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Empty(t, *repos)
	})
}

func TestPgRepository_GetRepositoriesCount(t *testing.T) {
	t.Run("returns correct count", func(t *testing.T) {
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

		repo1 := createTestRepository(t, "count-repo-1-"+uuid.NewString())
		repo2 := createTestRepository(t, "count-repo-2-"+uuid.NewString())
		repo3 := createTestRepository(t, "count-repo-3-"+uuid.NewString())

		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repo3)
		require.NoError(t, err)

		count, err := repo.GetRepositoriesCount(t.Context())
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("returns 0 for empty table", func(t *testing.T) {
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

		count, err := repo.GetRepositoriesCount(t.Context())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
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

		repo1 := createTestRepository(t, "active-repo")
		repo2 := createTestRepository(t, "deleted-repo")

		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)

		err = repo.DeleteRepository(t.Context(), repo2.Id)
		require.NoError(t, err)

		count, err := repo.GetRepositoriesCount(t.Context())
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestPgRepository_GetRepositoriesByUserCount(t *testing.T) {
	t.Run("returns correct count for user", func(t *testing.T) {
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

		repo1 := createTestRepository(t, "user-repo-1")
		repo1.OrganizationId = orgID1
		repo2 := createTestRepository(t, "user-repo-2")
		repo2.OrganizationId = orgID2

		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repo2)
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

		count, err := repo.GetRepositoriesByUserCount(t.Context(), userID)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestPgRepository_GetRepositoriesByUserAndOrganization(t *testing.T) {
	t.Run("returns repositories for user in specific organization", func(t *testing.T) {
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
		orgID := uuid.NewString()
		otherOrgID := uuid.NewString()

		repo1 := createTestRepository(t, "repo-1")
		repo1.OrganizationId = orgID
		repo2 := createTestRepository(t, "repo-2")
		repo2.OrganizationId = orgID
		repoOther := createTestRepository(t, "other-org-repo")
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
			 VALUES ($1, $2, $3, 'owner', NOW())`,
			uuid.NewString(), orgID, userID,
		)
		require.NoError(t, err)

		repos, err := repo.GetRepositoriesByUserAndOrganization(t.Context(), userID, orgID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, repos)
		require.Len(t, *repos, 2)
	})
}

func TestPgRepository_GetRepositoriesByUserAndOrganizationCount(t *testing.T) {
	t.Run("returns correct count", func(t *testing.T) {
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
		orgID := uuid.NewString()

		repo1 := createTestRepository(t, "repo-1")
		repo1.OrganizationId = orgID
		repo2 := createTestRepository(t, "repo-2")
		repo2.OrganizationId = orgID

		err = repo.CreateRepository(t.Context(), repo1)
		require.NoError(t, err)
		err = repo.CreateRepository(t.Context(), repo2)
		require.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		_, err = conn.Exec(t.Context(),
			`INSERT INTO organization_members (id, organization_id, user_id, role, joined_at)
			 VALUES ($1, $2, $3, 'owner', NOW())`,
			uuid.NewString(), orgID, userID,
		)
		require.NoError(t, err)

		count, err := repo.GetRepositoriesByUserAndOrganizationCount(t.Context(), userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestPgRepository_GetRepositoryById(t *testing.T) {
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

		testRepo := createTestRepository(t, "get-by-id-"+uuid.NewString())

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		found, err := repo.GetRepositoryById(t.Context(), testRepo.Id)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, testRepo.Id, found.Id)
		assert.Equal(t, testRepo.Name, found.Name)
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

		_, err = repo.GetRepositoryById(t.Context(), uuid.NewString())
		require.ErrorIs(t, err, ErrRepositoryNotFound)
	})
}

func TestPgRepository_UpdateRepository(t *testing.T) {
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

		testRepo := createTestRepository(t, "original-name")

		err = repo.CreateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		testRepo.Name = "updated-name"
		err = repo.UpdateRepository(t.Context(), testRepo)
		require.NoError(t, err)

		updated, err := repo.GetRepositoryById(t.Context(), testRepo.Id)
		require.NoError(t, err)
		assert.Equal(t, "updated-name", updated.Name)
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

		testRepo := createTestRepository(t, "non-existent")
		err = repo.UpdateRepository(t.Context(), testRepo)
		require.ErrorIs(t, err, ErrRepositoryNotFound)
	})
}

func TestPgRepository_GetFileTree(t *testing.T) {
	t.Run("successfully retrieves file tree from root", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md":  "# Test",
			"main.go":    "package main",
			"config.yml": "key: value",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 3)

		nodeNames := make(map[string]bool)
		for _, node := range response.Nodes {
			nodeNames[node.Name] = true
			assert.Equal(t, registryv1.NodeType_NODE_TYPE_FILE, node.Type)
		}

		assert.True(t, nodeNames["README.md"])
		assert.True(t, nodeNames["main.go"])
		assert.True(t, nodeNames["config.yml"])
	})

	t.Run("successfully retrieves file tree with directories", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md":     "# Test",
			"src/main.go":   "package main",
			"src/utils.go":  "package main",
			"docs/guide.md": "# Guide",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 3)

		var srcDir, docsDir *registryv1.FileTreeNode
		for _, node := range response.Nodes {
			switch node.Name {
			case "src":
				srcDir = node
			case "docs":
				docsDir = node
			}
		}

		require.NotNil(t, srcDir)
		assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, srcDir.Type)
		assert.Equal(t, "src", srcDir.Path)
		require.Len(t, srcDir.Children, 2)

		require.NotNil(t, docsDir)
		assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, docsDir.Type)
		assert.Equal(t, "docs", docsDir.Path)
		require.Len(t, docsDir.Children, 1)
	})

	t.Run("successfully retrieves file tree from subdirectory", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md":    "# Test",
			"src/main.go":  "package main",
			"src/utils.go": "package main",
			"docs/api.md":  "# API",
		})

		subPath := "src"
		response, err := repo.GetFileTree(t.Context(), testRepoPath, &subPath)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 2)

		nodeNames := make(map[string]bool)
		for _, node := range response.Nodes {
			nodeNames[node.Name] = true
			assert.Equal(t, registryv1.NodeType_NODE_TYPE_FILE, node.Type)
			assert.Contains(t, node.Path, "src/")
		}

		assert.True(t, nodeNames["main.go"])
		assert.True(t, nodeNames["utils.go"])
	})

	t.Run("handles deeply nested directory structure", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"src/internal/handlers/auth/login.go":    "package auth",
			"src/internal/handlers/auth/register.go": "package auth",
			"src/internal/handlers/user/profile.go":  "package user",
			"src/pkg/utils/helpers.go":               "package utils",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 1)

		srcNode := response.Nodes[0]
		assert.Equal(t, "src", srcNode.Name)
		assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, srcNode.Type)
		require.Len(t, srcNode.Children, 2)

		var internalNode, pkgNode *registryv1.FileTreeNode
		for _, child := range srcNode.Children {
			switch child.Name {
			case "internal":
				internalNode = child
			case "pkg":
				pkgNode = child
			}
		}

		require.NotNil(t, internalNode)
		assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, internalNode.Type)

		require.NotNil(t, pkgNode)
		assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, pkgNode.Type)
	})

	t.Run("correctly builds paths for nested files", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"src/internal/auth/login.go": "package auth",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)

		srcNode := response.Nodes[0]
		assert.Equal(t, "src", srcNode.Path)

		internalNode := srcNode.Children[0]
		assert.Equal(t, "src/internal", internalNode.Path)

		authNode := internalNode.Children[0]
		assert.Equal(t, "src/internal/auth", authNode.Path)

		loginFile := authNode.Children[0]
		assert.Equal(t, "src/internal/auth/login.go", loginFile.Path)
		assert.Equal(t, registryv1.NodeType_NODE_TYPE_FILE, loginFile.Type)
	})

	t.Run("handles empty repository", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			".gitkeep": "",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 1)
		assert.Equal(t, ".gitkeep", response.Nodes[0].Name)
	})

	t.Run("returns error for invalid repository path", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		_, err := repo.GetFileTree(t.Context(), "/invalid/path/to/repo", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open git repository")
	})

	t.Run("returns error for non-existent subpath", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md": "# Test",
		})

		subPath := "nonexistent/path"
		_, err := repo.GetFileTree(t.Context(), testRepoPath, &subPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path not found")
	})

	t.Run("handles mixed files and directories at same level", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md":      "# Test",
			"main.go":        "package main",
			"src/utils.go":   "package utils",
			"docs/README.md": "# Docs",
			"config.yml":     "key: value",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 5)

		fileCount := 0
		dirCount := 0
		for _, node := range response.Nodes {
			switch node.Type {
			case registryv1.NodeType_NODE_TYPE_FILE:
				fileCount++
			case registryv1.NodeType_NODE_TYPE_DIRECTORY:
				dirCount++
			}
		}

		assert.Equal(t, 3, fileCount, "should have 3 files at root level")
		assert.Equal(t, 2, dirCount, "should have 2 directories at root level")
	})

	t.Run("children list is empty for files", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md": "# Test",
			"main.go":   "package main",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)

		for _, node := range response.Nodes {
			if node.Type == registryv1.NodeType_NODE_TYPE_FILE {
				assert.Nil(t, node.Children, "files should have nil children")
			}
		}
	})

	t.Run("retrieves subdirectory tree with nested structure", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md":                  "# Test",
			"src/internal/auth/login.go": "package auth",
			"src/internal/user/user.go":  "package user",
			"src/pkg/utils.go":           "package pkg",
		})

		subPath := "src/internal"
		response, err := repo.GetFileTree(t.Context(), testRepoPath, &subPath)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 2)

		for _, node := range response.Nodes {
			assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, node.Type)
			assert.Contains(t, node.Path, "src/internal/")
			require.NotNil(t, node.Children)
			require.Len(t, node.Children, 1)
		}
	})

	t.Run("handles empty string subpath as root", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"README.md": "# Test",
			"main.go":   "package main",
		})

		emptySubPath := ""
		response, err := repo.GetFileTree(t.Context(), testRepoPath, &emptySubPath)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 2)
	})

	t.Run("file tree nodes have correct names", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"src/handlers/auth.go": "package handlers",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)

		srcNode := response.Nodes[0]
		assert.Equal(t, "src", srcNode.Name)

		handlersNode := srcNode.Children[0]
		assert.Equal(t, "handlers", handlersNode.Name)

		authFile := handlersNode.Children[0]
		assert.Equal(t, "auth.go", authFile.Name)
	})

	t.Run("handles repository with only directories", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"src/internal/.gitkeep": "",
			"docs/api/.gitkeep":     "",
			"tests/unit/.gitkeep":   "",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 3)

		for _, node := range response.Nodes {
			assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, node.Type)
			require.NotNil(t, node.Children)
		}
	})

	t.Run("handles special characters in file names", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"file-with-dashes.txt":     "content",
			"file_with_underscores.go": "package main",
			"file.multiple.dots.yaml":  "key: value",
		})

		response, err := repo.GetFileTree(t.Context(), testRepoPath, nil)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 3)

		nodeNames := make(map[string]bool)
		for _, node := range response.Nodes {
			nodeNames[node.Name] = true
		}

		assert.True(t, nodeNames["file-with-dashes.txt"])
		assert.True(t, nodeNames["file_with_underscores.go"])
		assert.True(t, nodeNames["file.multiple.dots.yaml"])
	})

	t.Run("retrieves file tree with multiple levels from subpath", func(t *testing.T) {
		repo, pool := setupTestRepository(t, "")
		defer pool.Close()

		testRepoPath := setupTestGitRepository(t, map[string]string{
			"src/internal/handlers/auth/login.go":    "package auth",
			"src/internal/handlers/auth/register.go": "package auth",
			"src/internal/handlers/user/profile.go":  "package user",
		})

		subPath := "src/internal/handlers"
		response, err := repo.GetFileTree(t.Context(), testRepoPath, &subPath)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Nodes, 2)

		for _, node := range response.Nodes {
			assert.True(t, node.Name == "auth" || node.Name == "user")
			assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, node.Type)
			assert.Contains(t, node.Path, "src/internal/handlers/")
		}
	})
}
