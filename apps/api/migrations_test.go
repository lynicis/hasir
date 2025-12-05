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

		// Verify migration version
		version, dirty, err := m.Version()
		assert.NoError(t, err)
		assert.False(t, dirty)
		assert.Equal(t, uint(11), version, "Expected migration version to be 11")
	})

	t.Run("idempotent - running migrations twice should not fail", func(t *testing.T) {
		container := setupPostgresContainer(t)
		defer func() {
			err := container.Terminate(context.Background())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(context.Background())
		require.NoError(t, err)

		// First migration
		m := setupMigration(t, connString)
		err = m.Up()
		assert.NoError(t, err)
		_, _ = m.Close()

		// Second migration (should return ErrNoChange)
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

		// Verify columns
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

		// Verify primary key
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

		// Verify table exists
		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'api_keys'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table api_keys should exist")

		// Verify foreign key to users
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

		// Verify indexes
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

		// Verify table exists
		var exists bool
		err = conn.QueryRow(context.Background(),
			`SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'ssh_keys'
			)`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table ssh_keys should exist")

		// Verify foreign key to users
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

		// Verify indexes
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

		// Verify unique constraint on public_key
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

		// Apply all migrations
		err = m.Up()
		require.NoError(t, err)

		// Verify tables exist
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

		// Rollback all migrations
		err = m.Down()
		assert.NoError(t, err)

		// Verify tables don't exist
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

		// Verify view doesn't exist
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

		// Step 1: Apply first migration (users table)
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

		// Step 2: Apply second migration (organizations table)
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

		// Step down one migration
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

		// Try to insert an API key with a non-existent user_id (should fail)
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

		// Insert a user
		userId := "test-user-id"
		_, err = conn.Exec(context.Background(),
			`INSERT INTO users (id, username, email, password, created_at)
			VALUES ($1, 'testuser', 'test@example.com', 'password', NOW())`, userId)
		require.NoError(t, err)

		// Insert an API key for the user
		_, err = conn.Exec(context.Background(),
			`INSERT INTO api_keys (user_id, name, key, created_at)
			VALUES ($1, 'test-key', 'test-key-value', NOW())`, userId)
		require.NoError(t, err)

		// Verify API key exists
		var count int
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM api_keys WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Delete the user
		_, err = conn.Exec(context.Background(),
			`DELETE FROM users WHERE id = $1`, userId)
		require.NoError(t, err)

		// Verify API key was cascade deleted
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

		// Insert a user
		userId := "test-user-id"
		_, err = conn.Exec(context.Background(),
			`INSERT INTO users (id, username, email, password, created_at)
			VALUES ($1, 'testuser', 'test@example.com', 'password', NOW())`, userId)
		require.NoError(t, err)

		// Insert an SSH key for the user
		_, err = conn.Exec(context.Background(),
			`INSERT INTO ssh_keys (user_id, name, public_key, created_at)
			VALUES ($1, 'test-key', 'ssh-rsa AAAAB3NzaC1...', NOW())`, userId)
		require.NoError(t, err)

		// Verify SSH key exists
		var count int
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM ssh_keys WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Delete the user
		_, err = conn.Exec(context.Background(),
			`DELETE FROM users WHERE id = $1`, userId)
		require.NoError(t, err)

		// Verify SSH key was cascade deleted
		err = conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM ssh_keys WHERE user_id = $1`, userId).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "SSH key should be cascade deleted when user is deleted")
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
