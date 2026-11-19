package sdkgenerator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCommandRunner(t *testing.T) {
	runner := NewDefaultCommandRunner()
	ctx := context.Background()

	t.Run("Run succeeds and captures stdout", func(t *testing.T) {
		out, err := runner.Run(ctx, "echo", []string{"hello"}, t.TempDir())
		require.NoError(t, err)
		assert.Equal(t, "hello\n", string(out))
	})

	t.Run("Run returns stderr in error on failure", func(t *testing.T) {
		_, err := runner.Run(ctx, "sh", []string{"-c", "echo boom >&2; exit 1"}, t.TempDir())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "boom")
	})

	t.Run("RunCombined returns output even when command exits non-zero", func(t *testing.T) {
		out, err := runner.RunCombined(ctx, "sh", []string{"-c", "echo partial; exit 1"}, t.TempDir())
		require.Error(t, err)
		assert.Contains(t, string(out), "partial")
	})

	t.Run("Pipe streams stdout from source to destination", func(t *testing.T) {
		dir := t.TempDir()
		err := runner.Pipe(ctx, "echo", []string{"piped"}, dir, "cat", nil, dir)
		require.NoError(t, err)
	})

	t.Run("Pipe returns error when source command fails", func(t *testing.T) {
		dir := t.TempDir()
		err := runner.Pipe(ctx, "sh", []string{"-c", "exit 2"}, dir, "cat", nil, dir)
		require.Error(t, err)
	})

	t.Run("Pipe does not deadlock when destination fails to start with large source output", func(t *testing.T) {
		dir := t.TempDir()
		timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err := runner.Pipe(
			timeoutCtx,
			"dd", []string{"if=/dev/zero", "bs=4096", "count=128"}, dir,
			"nonexistent-hasir-dst-binary", nil, dir,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start nonexistent-hasir-dst-binary")
		assert.NoError(t, timeoutCtx.Err(), "Pipe should return promptly without hitting context deadline")
	})

	t.Run("Pipe does not deadlock when destination exits early with large source output", func(t *testing.T) {
		dir := t.TempDir()
		timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err := runner.Pipe(
			timeoutCtx,
			"dd", []string{"if=/dev/zero", "bs=4096", "count=128"}, dir,
			"sh", []string{"-c", "exit 1"}, dir,
		)
		require.Error(t, err)
		assert.NoError(t, timeoutCtx.Err(), "Pipe should return promptly without hitting context deadline")
	})
}
