package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"hasir/api/internal/registry"
)

func initTempGitRepo(t *testing.T) (string, string) {
	t.Helper()
	repoDir := t.TempDir()

	repo, err := git.PlainInit(repoDir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	subDir := filepath.Join(repoDir, "proto", "user", "v1")
	require.NoError(t, os.MkdirAll(subDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "user.proto"), []byte("syntax = \"proto3\";\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Hello\n"), 0o600))

	_, err = wt.Add(".")
	require.NoError(t, err)

	hash, err := wt.Commit("initial proto commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Alice",
			Email: "alice@example.com",
			When:  time.Now().UTC(),
		},
	})
	require.NoError(t, err)

	return repoDir, hash.String()
}

func TestPgRepository_GitOperations(t *testing.T) {
	repoDir, commitHash := initTempGitRepo(t)
	repo := &PgRepository{
		tracer: noop.NewTracerProvider().Tracer("test"),
	}
	ctx := context.Background()

	t.Run("GetRecentCommit returns latest commit", func(t *testing.T) {
		commit, err := repo.GetRecentCommit(ctx, repoDir)
		require.NoError(t, err)
		assert.Equal(t, commitHash, commit.Id)
		assert.Equal(t, "initial proto commit", commit.Message)
		assert.Equal(t, "Alice", commit.User.Username)
	})

	t.Run("GetCommits paginates commits", func(t *testing.T) {
		commits, total, err := repo.GetCommits(ctx, repoDir, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		require.Len(t, commits, 1)
		assert.Equal(t, commitHash, commits[0].Id)
	})

	t.Run("GetFileTree returns nested nodes and supports subPath", func(t *testing.T) {
		treeResp, err := repo.GetFileTree(ctx, repoDir, nil)
		require.NoError(t, err)
		require.NotEmpty(t, treeResp.Nodes)

		sub := "proto/user/v1"
		subResp, err := repo.GetFileTree(ctx, repoDir, &sub)
		require.NoError(t, err)
		require.Len(t, subResp.Nodes, 1)
		assert.Equal(t, "user.proto", subResp.Nodes[0].Name)
		assert.Equal(t, "proto/user/v1/user.proto", subResp.Nodes[0].Path)
	})

	t.Run("GetFilePreview returns file content and detected mime type", func(t *testing.T) {
		preview, err := repo.GetFilePreview(ctx, repoDir, "proto/user/v1/user.proto")
		require.NoError(t, err)
		assert.Equal(t, "syntax = \"proto3\";\n", preview.Content)
		assert.Equal(t, "text/x-protobuf", preview.MimeType)
	})

	t.Run("GetFilePreview returns NotFound for missing file", func(t *testing.T) {
		_, err := repo.GetFilePreview(ctx, repoDir, "missing.proto")
		require.Error(t, err)
	})

	t.Run("GetCommits wraps ErrEmptyRepository for an empty git repo", func(t *testing.T) {
		emptyDir := t.TempDir()
		_, err := git.PlainInit(emptyDir, true)
		require.NoError(t, err)

		_, _, err = repo.GetCommits(ctx, emptyDir, 1, 10)
		require.Error(t, err)
		assert.ErrorIs(t, err, registry.ErrEmptyRepository)
	})

	t.Run("GetCommits does not wrap ErrEmptyRepository when storage path is missing", func(t *testing.T) {
		_, _, err := repo.GetCommits(ctx, filepath.Join(t.TempDir(), "missing"), 1, 10)
		require.Error(t, err)
		assert.NotErrorIs(t, err, registry.ErrEmptyRepository)
	})
}

func TestDetectMimeType(t *testing.T) {
	assert.Equal(t, "text/x-protobuf", detectMimeType("a/b/c.proto", []byte("syntax")))
	assert.Equal(t, "application/octet-stream", detectMimeType("binfile", []byte{0, 1, 2}))
	assert.Equal(t, "text/plain", detectMimeType("plainfile", []byte("hello")))
}
