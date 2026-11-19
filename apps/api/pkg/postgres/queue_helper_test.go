package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildJobStatusUpdateQuery(t *testing.T) {
	now := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	errMsg := "smtp timeout"

	t.Run("processing sets processed_at and increments attempts", func(t *testing.T) {
		sql, args := BuildJobStatusUpdateQuery(EmailJobQueries, "job-1", JobStatusProcessing, nil, now)
		assert.Equal(t, EmailJobQueries.Processing, sql)
		assert.Equal(t, "job-1", args["Id"])
		assert.Equal(t, JobStatusProcessing, args["Status"])
		assert.Equal(t, now, args["ProcessedAt"])
	})

	t.Run("completed sets completed_at", func(t *testing.T) {
		sql, args := BuildJobStatusUpdateQuery(SdkTriggerJobQueries, "job-2", JobStatusCompleted, nil, now)
		assert.Equal(t, SdkTriggerJobQueries.Completed, sql)
		assert.Equal(t, now, args["CompletedAt"])
	})

	t.Run("failed sets error_message", func(t *testing.T) {
		sql, args := BuildJobStatusUpdateQuery(SdkGenerationJobQueries, "job-3", JobStatusFailed, &errMsg, now)
		assert.Equal(t, SdkGenerationJobQueries.Failed, sql)
		assert.Equal(t, &errMsg, args["ErrorMessage"])
	})

	t.Run("pending resets status from processing", func(t *testing.T) {
		sql, args := BuildJobStatusUpdateQuery(EmailJobQueries, "job-4", JobStatusPending, nil, now)
		assert.Equal(t, EmailJobQueries.Pending, sql)
		assert.Equal(t, JobStatusPending, args["Status"])
	})

	t.Run("unknown status uses default query", func(t *testing.T) {
		sql, args := BuildJobStatusUpdateQuery(EmailJobQueries, "job-5", "custom", nil, now)
		assert.Equal(t, EmailJobQueries.Default, sql)
		assert.Equal(t, "custom", args["Status"])
	})
}

func TestHandleJobRetryOrFail(t *testing.T) {
	t.Run("resets to pending when attempts below max", func(t *testing.T) {
		var pendingCalled, failCalled bool
		HandleJobRetryOrFail(
			1, 3, errors.New("transient"),
			func() error { pendingCalled = true; return nil },
			func(string) error { failCalled = true; return nil },
		)
		assert.True(t, pendingCalled)
		assert.False(t, failCalled)
	})

	t.Run("marks failed with error message when attempts reach max", func(t *testing.T) {
		var pendingCalled bool
		var gotErr string
		HandleJobRetryOrFail(
			3, 3, errors.New("permanent failure"),
			func() error { pendingCalled = true; return nil },
			func(msg string) error { gotErr = msg; return nil },
		)
		assert.False(t, pendingCalled)
		assert.Equal(t, "permanent failure", gotErr)
	})

	t.Run("logs without panic when update callbacks return error", func(t *testing.T) {
		HandleJobRetryOrFail(
			1, 3, errors.New("boom"),
			func() error { return errors.New("db error") },
			func(string) error { return nil },
		)
		HandleJobRetryOrFail(
			3, 3, errors.New("boom"),
			func() error { return nil },
			func(string) error { return errors.New("db error") },
		)
	})
}
