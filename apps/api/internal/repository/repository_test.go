package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setupTestRepository(t *testing.T) (*PgRepository, *pgxpool.Pool, func()) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)

	repo := &PgRepository{
		connectionPool: pool,
		tracer:         noop.NewTracerProvider().Tracer("test"),
	}

	cleanup := func() {
		pool.Close()
	}

	return repo, pool, cleanup
}

func cleanupTestRepository(t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()
	_, _ = pool.Exec(context.Background(), "DELETE FROM repositories WHERE id = $1", id)
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
		UpdatedAt: now,
	}
}

func TestPgRepository_CreateRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, pool, cleanup := setupTestRepository(t)
		defer cleanup()

		ctx := context.Background()
		testRepo := createTestRepository(t, "test-repo-"+uuid.NewString())

		err := repo.CreateRepository(ctx, testRepo)
		require.NoError(t, err)

		cleanupTestRepository(t, pool, testRepo.Id)
	})

	t.Run("duplicate name returns already exists error", func(t *testing.T) {
		repo, pool, cleanup := setupTestRepository(t)
		defer cleanup()

		ctx := context.Background()
		repoName := "duplicate-repo-" + uuid.NewString()
		testRepo1 := createTestRepository(t, repoName)
		testRepo2 := createTestRepository(t, repoName)

		err := repo.CreateRepository(ctx, testRepo1)
		require.NoError(t, err)

		err = repo.CreateRepository(ctx, testRepo2)
		require.ErrorIs(t, err, ErrRepositoryAlreadyExists)

		cleanupTestRepository(t, pool, testRepo1.Id)
	})
}

func TestPgRepository_GetRepositoryByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, pool, cleanup := setupTestRepository(t)
		defer cleanup()

		ctx := context.Background()
		testRepo := createTestRepository(t, "get-by-name-"+uuid.NewString())

		err := repo.CreateRepository(ctx, testRepo)
		require.NoError(t, err)

		found, err := repo.GetRepositoryByName(ctx, testRepo.Name)
		require.NoError(t, err)
		require.Equal(t, testRepo.Id, found.Id)
		require.Equal(t, testRepo.Name, found.Name)

		cleanupTestRepository(t, pool, testRepo.Id)
	})

	t.Run("not found", func(t *testing.T) {
		repo, _, cleanup := setupTestRepository(t)
		defer cleanup()

		ctx := context.Background()

		_, err := repo.GetRepositoryByName(ctx, "nonexistent-repo-"+uuid.NewString())
		require.ErrorIs(t, err, ErrRepositoryNotFound)
	})

	t.Run("deleted repository not found", func(t *testing.T) {
		repo, pool, cleanup := setupTestRepository(t)
		defer cleanup()

		ctx := context.Background()
		testRepo := createTestRepository(t, "deleted-repo-"+uuid.NewString())

		err := repo.CreateRepository(ctx, testRepo)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, "UPDATE repositories SET deleted_at = NOW() WHERE id = $1", testRepo.Id)
		require.NoError(t, err)

		_, err = repo.GetRepositoryByName(ctx, testRepo.Name)
		require.ErrorIs(t, err, ErrRepositoryNotFound)

		cleanupTestRepository(t, pool, testRepo.Id)
	})
}
