package main

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testDB       = "test"
	testUsername = "test"
	testPassword = "test"
)

func TestMigrations(t *testing.T) {
	t.Run("apply all migrations successfully", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		assert.NoError(t, err)

		version, dirty, err := m.Version()
		assert.NoError(t, err)
		assert.False(t, dirty)
		assert.Equal(t, uint(17), version, "Expected migration version to be 17")
	})

	t.Run("idempotent - running migrations twice should not fail", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)
		m := setupMigration(t, connString)
		err = m.Up()
		assert.NoError(t, err)
		_, _ = m.Close()
		m = setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, migrate.ErrNoChange))
	})

	t.Run("verify all tables are created", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		expectedTables := []string{
			"users",
			"organizations",
			"refresh_tokens",
			"repositories",
			"organization_invites",
			"organization_members",
			"email_jobs",
			"api_keys",
			"ssh_keys",
			"sdk_preferences",
			"password_reset_tokens",
			"sdk_generation_jobs",
		}

		for _, tableName := range expectedTables {
			var exists bool
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = $1
				)`, tableName).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "Table %s should exist", tableName)
		}
	})

	t.Run("verify organization_members_view is created", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.views
				WHERE table_schema = 'public'
				AND table_name = 'organization_members_view'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "View organization_members_view should exist")
	})

	t.Run("verify users table schema", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()
		expectedColumns := map[string]string{
			"id":         "character varying",
			"username":   "character varying",
			"email":      "character varying",
			"password":   "character varying",
			"created_at": "timestamp with time zone",
			"deleted_at": "timestamp with time zone",
		}

		for columnName, expectedType := range expectedColumns {
			var dataType string
			err = conn.QueryRow(context.Background(),
				`SELECT data_type
				FROM information_schema.columns
				WHERE table_name = 'users'
				AND column_name = $1`, columnName).Scan(&dataType)
			require.NoError(t, err, "Column %s should exist", columnName)
			assert.Equal(t, expectedType, dataType, "Column %s should have type %s", columnName, expectedType)
		}
		var constraintType string
		err = conn.QueryRow(context.Background(),
			`SELECT constraint_type
			FROM information_schema.table_constraints
			WHERE table_name = 'users'
			AND constraint_type = 'PRIMARY KEY'`).Scan(&constraintType)
		require.NoError(t, err)
		assert.Equal(t, "PRIMARY KEY", constraintType)
	})

	t.Run("verify api_keys table schema and indexes", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()
		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'api_keys'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table api_keys should exist")
		var fkExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				WHERE tc.table_name = 'api_keys'
				AND tc.constraint_type = 'FOREIGN KEY'
				AND kcu.column_name = 'user_id'
			)`).Scan(&fkExists)
		require.NoError(t, err)
		assert.True(t, fkExists, "Foreign key on user_id should exist")
		expectedIndexes := []string{
			"idx_api_keys_user_id",
			"idx_api_keys_key",
			"idx_api_keys_deleted_at",
		}

		for _, indexName := range expectedIndexes {
			var indexExists bool
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT 1
					FROM pg_indexes
					WHERE tablename = 'api_keys'
					AND indexname = $1
				)`, indexName).Scan(&indexExists)
			require.NoError(t, err)
			assert.True(t, indexExists, "Index %s should exist", indexName)
		}
	})

	t.Run("verify ssh_keys table schema and indexes", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()
		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'ssh_keys'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table ssh_keys should exist")
		var fkExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				WHERE tc.table_name = 'ssh_keys'
				AND tc.constraint_type = 'FOREIGN KEY'
				AND kcu.column_name = 'user_id'
			)`).Scan(&fkExists)
		require.NoError(t, err)
		assert.True(t, fkExists, "Foreign key on user_id should exist")
		expectedIndexes := []string{
			"idx_ssh_keys_user_id",
			"idx_ssh_keys_deleted_at",
		}

		for _, indexName := range expectedIndexes {
			var indexExists bool
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT 1
					FROM pg_indexes
					WHERE tablename = 'ssh_keys'
					AND indexname = $1
				)`, indexName).Scan(&indexExists)
			require.NoError(t, err)
			assert.True(t, indexExists, "Index %s should exist", indexName)
		}
		var uniqueExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE tablename = 'ssh_keys'
				AND indexdef LIKE '%UNIQUE%'
				AND indexdef LIKE '%public_key%'
			)`).Scan(&uniqueExists)
		require.NoError(t, err)
		assert.True(t, uniqueExists, "Unique constraint on public_key should exist")
	})

	t.Run("verify sdk_preferences table schema and indexes", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'sdk_preferences'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table sdk_preferences should exist")

		expectedColumns := map[string]string{
			"id":            "character varying",
			"repository_id": "character varying",
			"sdk":           "USER-DEFINED",
			"status":        "boolean",
			"created_at":    "timestamp with time zone",
			"updated_at":    "timestamp with time zone",
		}

		for columnName, expectedType := range expectedColumns {
			var dataType string
			err = conn.QueryRow(context.Background(),
				`SELECT data_type
				FROM information_schema.columns
				WHERE table_name = 'sdk_preferences'
				AND column_name = $1`, columnName).Scan(&dataType)
			require.NoError(t, err, "Column %s should exist", columnName)
			assert.Equal(t, expectedType, dataType, "Column %s should have type %s", columnName, expectedType)
		}

		var fkExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				WHERE tc.table_name = 'sdk_preferences'
				AND tc.constraint_type = 'FOREIGN KEY'
				AND kcu.column_name = 'repository_id'
			)`).Scan(&fkExists)
		require.NoError(t, err)
		assert.True(t, fkExists, "Foreign key on repository_id should exist")

		expectedIndexes := []string{
			"idx_sdk_preferences_repository_id",
			"idx_sdk_preferences_sdk",
		}

		for _, indexName := range expectedIndexes {
			var indexExists bool
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT 1
					FROM pg_indexes
					WHERE tablename = 'sdk_preferences'
					AND indexname = $1
				)`, indexName).Scan(&indexExists)
			require.NoError(t, err)
			assert.True(t, indexExists, "Index %s should exist", indexName)
		}

		var uniqueExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE tablename = 'sdk_preferences'
				AND indexdef LIKE '%UNIQUE%'
				AND indexdef LIKE '%repository_id%'
				AND indexdef LIKE '%sdk%'
			)`).Scan(&uniqueExists)
		require.NoError(t, err)
		assert.True(t, uniqueExists, "Unique constraint on (repository_id, sdk) should exist")

		var cascadeExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM information_schema.referential_constraints rc
				WHERE rc.constraint_name IN (
					SELECT constraint_name
					FROM information_schema.key_column_usage
					WHERE table_name = 'sdk_preferences'
					AND column_name = 'repository_id'
				)
				AND rc.delete_rule = 'CASCADE'
			)`).Scan(&cascadeExists)
		require.NoError(t, err)
		assert.True(t, cascadeExists, "ON DELETE CASCADE should be set for repository_id foreign key")
	})

	t.Run("verify sdk_generation_jobs table schema and indexes", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'sdk_generation_jobs'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table sdk_generation_jobs should exist")

		expectedColumns := map[string]string{
			"id":            "character varying",
			"repository_id": "character varying",
			"commit_hash":   "character varying",
			"sdk":           "USER-DEFINED",
			"status":        "character varying",
			"attempts":      "integer",
			"max_attempts":  "integer",
			"created_at":    "timestamp with time zone",
			"processed_at":  "timestamp with time zone",
			"completed_at":  "timestamp with time zone",
			"error_message": "text",
		}

		for columnName, expectedType := range expectedColumns {
			var dataType string
			err = conn.QueryRow(context.Background(),
				`SELECT data_type
				FROM information_schema.columns
				WHERE table_name = 'sdk_generation_jobs'
				AND column_name = $1`, columnName).Scan(&dataType)
			require.NoError(t, err, "Column %s should exist", columnName)
			assert.Equal(t, expectedType, dataType, "Column %s should have type %s", columnName, expectedType)
		}

		var fkExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				WHERE tc.table_name = 'sdk_generation_jobs'
				AND tc.constraint_type = 'FOREIGN KEY'
				AND kcu.column_name = 'repository_id'
			)`).Scan(&fkExists)
		require.NoError(t, err)
		assert.True(t, fkExists, "Foreign key on repository_id should exist")

		expectedIndexes := []string{
			"idx_sdk_generation_jobs_repository_id",
			"idx_sdk_generation_jobs_status",
			"idx_sdk_generation_jobs_commit_hash",
			"idx_sdk_generation_jobs_created_at",
		}

		for _, indexName := range expectedIndexes {
			var indexExists bool
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT 1
					FROM pg_indexes
					WHERE tablename = 'sdk_generation_jobs'
					AND indexname = $1
				)`, indexName).Scan(&indexExists)
			require.NoError(t, err)
			assert.True(t, indexExists, "Index %s should exist", indexName)
		}

		var cascadeExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM information_schema.referential_constraints rc
				WHERE rc.constraint_name IN (
					SELECT constraint_name
					FROM information_schema.key_column_usage
					WHERE table_name = 'sdk_generation_jobs'
					AND column_name = 'repository_id'
				)
				AND rc.delete_rule = 'CASCADE'
			)`).Scan(&cascadeExists)
		require.NoError(t, err)
		assert.True(t, cascadeExists, "ON DELETE CASCADE should be set for repository_id foreign key")
	})

	t.Run("verify password_reset_tokens table schema and indexes", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'password_reset_tokens'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table password_reset_tokens should exist")

		expectedColumns := map[string]string{
			"id":         "uuid",
			"user_id":    "character varying",
			"token":      "character varying",
			"expires_at": "timestamp with time zone",
			"created_at": "timestamp with time zone",
			"used_at":    "timestamp with time zone",
		}

		for columnName, expectedType := range expectedColumns {
			var dataType string
			err = conn.QueryRow(context.Background(),
				`SELECT data_type
				FROM information_schema.columns
				WHERE table_name = 'password_reset_tokens'
				AND column_name = $1`, columnName).Scan(&dataType)
			require.NoError(t, err, "Column %s should exist", columnName)
			assert.Equal(t, expectedType, dataType, "Column %s should have type %s", columnName, expectedType)
		}

		var fkExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				WHERE tc.table_name = 'password_reset_tokens'
				AND tc.constraint_type = 'FOREIGN KEY'
				AND kcu.column_name = 'user_id'
			)`).Scan(&fkExists)
		require.NoError(t, err)
		assert.True(t, fkExists, "Foreign key on user_id should exist")

		expectedIndexes := []string{
			"idx_password_reset_tokens_token",
			"idx_password_reset_tokens_user_id",
			"idx_password_reset_tokens_expires_at",
		}

		for _, indexName := range expectedIndexes {
			var indexExists bool
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT 1
					FROM pg_indexes
					WHERE tablename = 'password_reset_tokens'
					AND indexname = $1
				)`, indexName).Scan(&indexExists)
			require.NoError(t, err)
			assert.True(t, indexExists, "Index %s should exist", indexName)
		}

		var uniqueExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE tablename = 'password_reset_tokens'
				AND indexdef LIKE '%UNIQUE%'
				AND indexdef LIKE '%token%'
			)`).Scan(&uniqueExists)
		require.NoError(t, err)
		assert.True(t, uniqueExists, "Unique constraint on token should exist")

		var cascadeExists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM information_schema.referential_constraints rc
				WHERE rc.constraint_name IN (
					SELECT constraint_name
					FROM information_schema.key_column_usage
					WHERE table_name = 'password_reset_tokens'
					AND column_name = 'user_id'
				)
				AND rc.delete_rule = 'CASCADE'
			)`).Scan(&cascadeExists)
		require.NoError(t, err)
		assert.True(t, cascadeExists, "ON DELETE CASCADE should be set for user_id foreign key")
	})

	t.Run("rollback all migrations successfully", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'users'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Users table should exist before rollback")

		err = m.Down()
		assert.NoError(t, err)

		expectedTables := []string{
			"users",
			"organizations",
			"refresh_tokens",
			"repositories",
			"organization_invites",
			"organization_members",
			"email_jobs",
			"api_keys",
			"ssh_keys",
			"sdk_preferences",
			"password_reset_tokens",
			"sdk_generation_jobs",
		}

		for _, tableName := range expectedTables {
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = $1
				)`, tableName).Scan(&exists)
			require.NoError(t, err)
			assert.False(t, exists, "Table %s should not exist after rollback", tableName)
		}
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.views
				WHERE table_schema = 'public'
				AND table_name = 'organization_members_view'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "View organization_members_view should not exist after rollback")
	})

	t.Run("step up and step down migrations", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()
		err = m.Steps(1)
		require.NoError(t, err)

		version, _, err := m.Version()
		require.NoError(t, err)
		assert.Equal(t, uint(1), version)

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'users'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Users table should exist after first migration")
		err = m.Steps(1)
		require.NoError(t, err)

		version, _, err = m.Version()
		require.NoError(t, err)
		assert.Equal(t, uint(2), version)

		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'organizations'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Organizations table should exist after second migration")
		err = m.Steps(-1)
		require.NoError(t, err)

		version, _, err = m.Version()
		require.NoError(t, err)
		assert.Equal(t, uint(1), version)

		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'organizations'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "Organizations table should not exist after stepping down")
	})

	t.Run("verify foreign key constraints work correctly", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()
		_, err = conn.Exec(context.Background(),
			`INSERT INTO api_keys (user_id, name, key, created_at)
			VALUES ('non-existent-user-id', 'test-key', 'test-key-value', NOW())`)
		assert.Error(t, err, "Should fail to insert API key with non-existent user_id")
		assert.Contains(t, err.Error(), "foreign key", "Error should mention foreign key constraint")
	})

	t.Run("verify ON DELETE CASCADE works for api_keys", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()
		userId := "test-user-id"
		_, err = conn.Exec(context.Background(),
			`INSERT INTO users (id, username, email, password, created_at)
			VALUES ($1, 'testuser', 'test@example.com', 'password', NOW())`, userId)
		require.NoError(t, err)
		_, err = conn.Exec(context.Background(),
			`INSERT INTO api_keys (user_id, name, key, created_at)
			VALUES ($1, 'test-key', 'test-key-value', NOW())`, userId)
		require.NoError(t, err)
		var count int
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM api_keys WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		_, err = conn.Exec(context.Background(),
			`DELETE FROM users WHERE id = $1`, userId)
		require.NoError(t, err)
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM api_keys WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "API key should be cascade deleted when user is deleted")
	})

	t.Run("verify ON DELETE CASCADE works for ssh_keys", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()
		userId := "test-user-id"
		_, err = conn.Exec(context.Background(),
			`INSERT INTO users (id, username, email, password, created_at)
			VALUES ($1, 'testuser', 'test@example.com', 'password', NOW())`, userId)
		require.NoError(t, err)
		_, err = conn.Exec(context.Background(),
			`INSERT INTO ssh_keys (user_id, name, public_key, created_at)
			VALUES ($1, 'test-key', 'ssh-rsa AAAAB3NzaC1...', NOW())`, userId)
		require.NoError(t, err)
		var count int
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM ssh_keys WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		_, err = conn.Exec(context.Background(),
			`DELETE FROM users WHERE id = $1`, userId)
		require.NoError(t, err)
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM ssh_keys WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "SSH key should be cascade deleted when user is deleted")
	})

	t.Run("verify ON DELETE CASCADE works for password_reset_tokens", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()
		userId := "test-user-id"
		_, err = conn.Exec(context.Background(),
			`INSERT INTO users (id, username, email, password, created_at)
			VALUES ($1, 'testuser', 'test@example.com', 'password', NOW())`, userId)
		require.NoError(t, err)
		_, err = conn.Exec(context.Background(),
			`INSERT INTO password_reset_tokens (user_id, token, expires_at, created_at)
			VALUES ($1, 'test-token-12345', NOW() + INTERVAL '1 hour', NOW())`, userId)
		require.NoError(t, err)
		var count int
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM password_reset_tokens WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		_, err = conn.Exec(context.Background(),
			`DELETE FROM users WHERE id = $1`, userId)
		require.NoError(t, err)
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM password_reset_tokens WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Password reset token should be cascade deleted when user is deleted")
	})

	t.Run("verify pg_trgm extension is enabled", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM pg_extension
				WHERE extname = 'pg_trgm'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "pg_trgm extension should be enabled")
	})

	t.Run("verify search_items materialized view is created", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM pg_matviews
				WHERE schemaname = 'public'
				AND matviewname = 'search_items'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Materialized view search_items should exist")
	})

	t.Run("verify search_items view has correct columns", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		expectedColumns := []string{
			"id",
			"name",
			"item_type",
			"organization_id",
			"created_at",
			"deleted_at",
		}

		for _, columnName := range expectedColumns {
			var exists bool
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT 1
					FROM pg_attribute a
					JOIN pg_class c ON a.attrelid = c.oid
					JOIN pg_namespace n ON c.relnamespace = n.oid
					WHERE n.nspname = 'public'
					AND c.relname = 'search_items'
					AND a.attname = $1
					AND a.attnum > 0
					AND NOT a.attisdropped
				)`, columnName).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "Column %s should exist in search_items view", columnName)
		}
	})

	t.Run("verify search_items has GIN trigram index", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE tablename = 'search_items'
				AND indexname = 'idx_search_items_name_trgm'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "GIN trigram index on name should exist")

		var indexDef string
		err = conn.QueryRow(context.Background(),
			`SELECT indexdef
			FROM pg_indexes
			WHERE tablename = 'search_items'
			AND indexname = 'idx_search_items_name_trgm'`).Scan(&indexDef)
		require.NoError(t, err)
		assert.Contains(t, indexDef, "gin", "Index should use GIN")
		assert.Contains(t, indexDef, "gin_trgm_ops", "Index should use gin_trgm_ops operator class")
	})

	t.Run("verify search_items has all required indexes", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		expectedIndexes := []string{
			"idx_search_items_name_trgm",
			"idx_search_items_item_type",
			"idx_search_items_organization_id",
			"idx_search_items_deleted_at",
			"idx_search_items_created_at",
			"idx_search_items_id_type",
		}

		for _, indexName := range expectedIndexes {
			var exists bool
			err = conn.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT 1
					FROM pg_indexes
					WHERE tablename = 'search_items'
					AND indexname = $1
				)`, indexName).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "Index %s should exist", indexName)
		}
	})

	t.Run("verify search_items unique index on id and item_type", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var indexDef string
		err = conn.QueryRow(context.Background(),
			`SELECT indexdef
			FROM pg_indexes
			WHERE tablename = 'search_items'
			AND indexname = 'idx_search_items_id_type'`).Scan(&indexDef)
		require.NoError(t, err)
		assert.Contains(t, indexDef, "UNIQUE", "Index should be UNIQUE")
		assert.Contains(t, indexDef, "id", "Index should include id column")
		assert.Contains(t, indexDef, "item_type", "Index should include item_type column")
	})

	t.Run("verify search_items view returns data from organizations", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		_, err = conn.Exec(context.Background(),
			`INSERT INTO users (id, username, email, password, created_at)
			VALUES ('user-id', 'testuser', 'test@example.com', 'password', NOW())`)
		require.NoError(t, err)

		_, err = conn.Exec(context.Background(),
			`INSERT INTO organizations (id, name, visibility, created_by, created_at)
			VALUES ('test-org-id', 'test-org', 'private', 'user-id', NOW())`)
		require.NoError(t, err)

		_, err = conn.Exec(context.Background(), "REFRESH MATERIALIZED VIEW search_items")
		require.NoError(t, err)

		var id, name, itemType string
		err = conn.QueryRow(context.Background(),
			`SELECT id, name, item_type
			FROM search_items
			WHERE id = 'test-org-id'`).Scan(&id, &name, &itemType)
		require.NoError(t, err)
		assert.Equal(t, "test-org-id", id)
		assert.Equal(t, "test-org", name)
		assert.Equal(t, "organization", itemType)
	})

	t.Run("verify search_items view returns data from repositories", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		_, err = conn.Exec(context.Background(),
			`INSERT INTO users (id, username, email, password, created_at)
			VALUES ('user-id', 'testuser', 'test@example.com', 'password', NOW())`)
		require.NoError(t, err)

		_, err = conn.Exec(context.Background(),
			`INSERT INTO organizations (id, name, visibility, created_by, created_at)
			VALUES ('test-org-id', 'test-org', 'private', 'user-id', NOW())`)
		require.NoError(t, err)

		_, err = conn.Exec(context.Background(),
			`INSERT INTO repositories (id, name, visibility, organization_id, created_by, path, created_at)
			VALUES ('test-repo-id', 'test-repo', 'private', 'test-org-id', 'user-id', '/repos/test-repo', NOW())`)
		require.NoError(t, err)

		_, err = conn.Exec(context.Background(), "REFRESH MATERIALIZED VIEW search_items")
		require.NoError(t, err)

		var id, name, itemType, orgId string
		err = conn.QueryRow(context.Background(),
			`SELECT id, name, item_type, organization_id
			FROM search_items
			WHERE id = 'test-repo-id'`).Scan(&id, &name, &itemType, &orgId)
		require.NoError(t, err)
		assert.Equal(t, "test-repo-id", id)
		assert.Equal(t, "test-repo", name)
		assert.Equal(t, "repository", itemType)
		assert.Equal(t, "test-org-id", orgId)
	})

	t.Run("verify pg_trgm similarity search works", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		_, err = conn.Exec(context.Background(),
			`INSERT INTO users (id, username, email, password, created_at)
			VALUES ('user-id', 'testuser', 'test@example.com', 'password', NOW())`)
		require.NoError(t, err)

		_, err = conn.Exec(context.Background(),
			`INSERT INTO organizations (id, name, visibility, created_by, created_at)
			VALUES ('org-1', 'testing-org', 'private', 'user-id', NOW()),
			       ('org-2', 'test', 'private', 'user-id', NOW()),
			       ('org-3', 'production', 'private', 'user-id', NOW())`)
		require.NoError(t, err)

		_, err = conn.Exec(context.Background(), "REFRESH MATERIALIZED VIEW search_items")
		require.NoError(t, err)

		var count int
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*)
			FROM search_items
			WHERE name % 'test'`).Scan(&count)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2, "Should find at least 2 items matching 'test'")

		var topResult string
		err = conn.QueryRow(context.Background(),
			`SELECT name
			FROM search_items
			WHERE name % 'test'
			ORDER BY similarity(name, 'test') DESC
			LIMIT 1`).Scan(&topResult)
		require.NoError(t, err)
		assert.Equal(t, "test", topResult, "Exact match should be ranked highest")
	})

	t.Run("verify search migrations can be rolled back", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		m := setupMigration(t, connString)
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		require.NoError(t, err)

		conn, err := pgx.Connect(context.Background(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(context.Background())
		}()

		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM pg_matviews
				WHERE schemaname = 'public'
				AND matviewname = 'search_items'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "search_items view should exist before rollback")

		err = m.Steps(-4)
		require.NoError(t, err)

		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM pg_matviews
				WHERE schemaname = 'public'
				AND matviewname = 'search_items'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "search_items view should not exist after rollback")

		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT 1
				FROM pg_extension
				WHERE extname = 'pg_trgm'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "pg_trgm extension should not exist after rollback")
	})
}

func setupPostgresContainer(t *testing.T) *postgres.PostgresContainer {
	t.Helper()

	container, err := postgres.Run(
		context.Background(),
		"postgres:16-alpine",
		postgres.WithDatabase(testDB),
		postgres.WithUsername(testUsername),
		postgres.WithPassword(testPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	require.NoError(t, err)

	return container
}

func setupMigration(t *testing.T, connString string) *migrate.Migrate {
	t.Helper()

	var migrationURL string
	if len(connString) > 0 && connString[len(connString)-1] == '?' {
		migrationURL = connString + "sslmode=disable"
	} else {
		migrationURL = connString + "?sslmode=disable"
	}

	m, err := migrate.New(
		"file://migrations",
		migrationURL,
	)
	require.NoError(t, err)

	return m
}
