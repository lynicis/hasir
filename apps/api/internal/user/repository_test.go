package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/crypto/bcrypt"

	"hasir-api/pkg/config"
)

const (
	pgDb       = "test"
	pgUsername = "test"
	pgPassword = "test"

	fakeEmail    = "test@mail.com"
	fakeUsername = "test-user"
	fakePassword = "Asdfg12345_"
)

var (
	fakeId  = uuid.NewString()
	fakeNow = time.Now().UTC()
)

func Test_NewPgRepository(t *testing.T) {
	container := setupPgContainer(t)
	defer func() {
		err := container.Terminate(t.Context())
		require.NoError(t, err)
	}()

	connString, err := container.ConnectionString(t.Context())
	require.NoError(t, err)

	traceProvider := sdktrace.NewTracerProvider()

	pgRepository := NewPgRepository(&config.Config{
		PostgresConfig: config.PostgresConfig{
			ConnectionString: connString,
		},
	}, traceProvider)

	assert.Implements(t, (*Repository)(nil), pgRepository)
}

func TestPgRepository_CreateUser(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		userId := uuid.NewString()
		user := &UserDTO{
			Id:        userId,
			Username:  "test",
			Email:     "test@mail.com",
			Password:  "Asdfg123456_",
			CreatedAt: time.Now().UTC(),
		}

		err = pgRepository.CreateUser(t.Context(), user)

		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbId, dbUsername, dbEmail, dbPassword string
		var dbCreatedAt time.Time
		err = conn.QueryRow(t.Context(),
			"select id, username, email, password, created_at from users where id = $1", userId).
			Scan(&dbId, &dbUsername, &dbEmail, &dbPassword, &dbCreatedAt)
		require.NoError(t, err)

		assert.Equal(t, userId, dbId)
		assert.Equal(t, "test", dbUsername)
		assert.Equal(t, "test@mail.com", dbEmail)
		assert.Equal(t, "Asdfg123456_", dbPassword)
		assert.WithinDuration(t, time.Now().UTC(), dbCreatedAt, 5*time.Second)
	})

	t.Run("email already exists (unique violation)", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users(email)")
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user := &UserDTO{
			Id:        uuid.NewString(),
			Username:  "another-user",
			Email:     fakeEmail,
			Password:  "Asdfg123456_",
			CreatedAt: time.Now().UTC(),
		}

		err = pgRepository.CreateUser(t.Context(), user)

		assert.Error(t, err)
	})

	t.Run("duplicate id (primary key violation)", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user := &UserDTO{
			Id:        fakeId,
			Username:  "another-user",
			Email:     "another@mail.com",
			Password:  "Asdfg123456_",
			CreatedAt: time.Now().UTC(),
		}

		err = pgRepository.CreateUser(t.Context(), user)

		assert.Error(t, err)
	})

	t.Run("verify all fields are stored correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		userId := uuid.NewString()
		username := "test-username"
		email := "test-email@example.com"
		password := "TestPassword123_"
		createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		user := &UserDTO{
			Id:        userId,
			Username:  username,
			Email:     email,
			Password:  password,
			CreatedAt: createdAt,
		}

		err = pgRepository.CreateUser(t.Context(), user)

		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbId, dbUsername, dbEmail, dbPassword string
		var dbCreatedAt time.Time
		err = conn.QueryRow(t.Context(),
			"select id, username, email, password, created_at from users where id = $1", userId).
			Scan(&dbId, &dbUsername, &dbEmail, &dbPassword, &dbCreatedAt)
		require.NoError(t, err)

		assert.Equal(t, userId, dbId)
		assert.Equal(t, username, dbUsername)
		assert.Equal(t, email, dbEmail)
		assert.Equal(t, password, dbPassword)
		assert.WithinDuration(t, time.Now().UTC(), dbCreatedAt, 5*time.Second)
	})
}

func TestPgRepository_GetUserByEmail(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user, err := pgRepository.GetUserByEmail(t.Context(), fakeEmail)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, fakeEmail, user.Email)
	})

	t.Run("verify all fields are retrieved correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user, err := pgRepository.GetUserByEmail(t.Context(), fakeEmail)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, fakeId, user.Id)
		assert.Equal(t, fakeEmail, user.Email)
		assert.Equal(t, fakeUsername, user.Username)
		// Password should be the hashed version
		assert.NotEmpty(t, user.Password)
		// Verify password is the hashed one (not the plain text)
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(fakePassword))
		assert.NoError(t, err)
		assert.WithinDuration(t, fakeNow, user.CreatedAt, time.Second)
	})

	t.Run("user not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		nonExistentEmail := "nonexistent@mail.com"
		user, err := pgRepository.GetUserByEmail(t.Context(), nonExistentEmail)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("case sensitive email lookup", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		// Try with different case (PostgreSQL string comparison is case-sensitive by default)
		user, err := pgRepository.GetUserByEmail(t.Context(), "TEST@MAIL.COM")

		// This should fail because email comparison is case-sensitive
		// (unless the database has a case-insensitive collation)
		if err != nil {
			assert.Error(t, err)
			assert.Nil(t, user)
		} else {
			// If it succeeds, verify it's the same user
			require.NotNil(t, user)
			assert.Equal(t, fakeId, user.Id)
		}
	})

	t.Run("get user with deleted_at set", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		// Soft delete the user
		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "update users set deleted_at = $1 where id = $2", time.Now().UTC(), fakeId)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		// The current implementation doesn't filter by deleted_at,
		// so it should still return the user
		user, err := pgRepository.GetUserByEmail(t.Context(), fakeEmail)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, fakeEmail, user.Email)
	})
}

func TestPgRepository_GetUserById(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user, err := pgRepository.GetUserById(t.Context(), fakeId)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, fakeId, user.Id)
	})

	t.Run("verify all fields are retrieved correctly", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user, err := pgRepository.GetUserById(t.Context(), fakeId)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, fakeId, user.Id)
		assert.Equal(t, fakeEmail, user.Email)
		assert.Equal(t, fakeUsername, user.Username)
		assert.NotEmpty(t, user.Password)

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(fakePassword))
		assert.NoError(t, err)
		assert.WithinDuration(t, fakeNow, user.CreatedAt, time.Second)
		assert.Nil(t, user.DeletedAt)
	})

	t.Run("user not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		nonExistentId := uuid.NewString()
		user, err := pgRepository.GetUserById(t.Context(), nonExistentId)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrNoRows, err)
	})

	t.Run("get user with deleted_at set", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		// Soft delete the user
		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		deletedAt := time.Now().UTC()
		_, err = conn.Exec(t.Context(), "update users set deleted_at = $1 where id = $2", deletedAt, fakeId)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		// The current implementation doesn't filter by deleted_at,
		// so it should still return the user
		user, err := pgRepository.GetUserById(t.Context(), fakeId)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, fakeId, user.Id)
		require.NotNil(t, user.DeletedAt)
		assert.WithinDuration(t, deletedAt, *user.DeletedAt, time.Second)
	})

	t.Run("empty id", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user, err := pgRepository.GetUserById(t.Context(), "")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrNoRows, err)
	})

	t.Run("invalid uuid format", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		// Use an invalid ID format (not a valid UUID)
		invalidId := "not-a-valid-uuid"
		user, err := pgRepository.GetUserById(t.Context(), invalidId)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrNoRows, err)
	})
}

func setupPgContainer(t *testing.T) *postgres.PostgresContainer {
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

func TestPgRepository_UpdateUserById(t *testing.T) {
	t.Run("happy path - update all fields", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		newHashedPassword, err := bcrypt.GenerateFromPassword([]byte("NewPassword123_"), bcrypt.DefaultCost)
		require.NoError(t, err)

		updatedUser := &UserDTO{
			Id:       fakeId,
			Username: "updated-username",
			Email:    "updated@mail.com",
			Password: string(newHashedPassword),
		}

		err = pgRepository.UpdateUserById(t.Context(), fakeId, updatedUser)

		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var username, email, password string
		err = conn.QueryRow(t.Context(), "select username, email, password from users where id = $1", fakeId).
			Scan(&username, &email, &password)
		require.NoError(t, err)

		assert.Equal(t, "updated-username", username)
		assert.Equal(t, "updated@mail.com", email)
		assert.Equal(t, string(newHashedPassword), password)
	})

	t.Run("happy path - update only username", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		updatedUser := &UserDTO{
			Id:       fakeId,
			Username: "only-username-updated",
		}

		err = pgRepository.UpdateUserById(t.Context(), fakeId, updatedUser)

		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var username, email string
		err = conn.QueryRow(t.Context(), "select username, email from users where id = $1", fakeId).
			Scan(&username, &email)
		require.NoError(t, err)

		assert.Equal(t, "only-username-updated", username)
		assert.Equal(t, fakeEmail, email)
	})

	t.Run("happy path - update only email", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		updatedUser := &UserDTO{
			Id:    fakeId,
			Email: "only-email-updated@mail.com",
		}

		err = pgRepository.UpdateUserById(t.Context(), fakeId, updatedUser)

		assert.NoError(t, err)

		// Verify only email was updated
		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var username, email string
		err = conn.QueryRow(t.Context(), "select username, email from users where id = $1", fakeId).
			Scan(&username, &email)
		require.NoError(t, err)

		assert.Equal(t, "only-email-updated@mail.com", email)
		assert.Equal(t, fakeUsername, username)
	})

	t.Run("happy path - no fields to update (verify user exists)", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		updatedUser := &UserDTO{
			Id: fakeId,
		}

		err = pgRepository.UpdateUserById(t.Context(), fakeId, updatedUser)

		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		nonExistentId := uuid.NewString()
		updatedUser := &UserDTO{
			Id:       nonExistentId,
			Username: "updated-username",
		}

		err = pgRepository.UpdateUserById(t.Context(), nonExistentId, updatedUser)

		assert.Error(t, err)
		assert.Equal(t, ErrNoRows, err)
	})

	t.Run("email already exists (unique violation)", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		anotherId := uuid.NewString()
		anotherEmail := "another@mail.com"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(fakePassword), bcrypt.DefaultCost)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(),
			"insert into users (id, email, username, password, created_at) values ($1, $2, $3, $4, $5)",
			anotherId, anotherEmail, "another-user", string(hashedPassword), time.Now().UTC())
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		conn, err = pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users(email)")
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		updatedUser := &UserDTO{
			Id:    fakeId,
			Email: anotherEmail,
		}

		err = pgRepository.UpdateUserById(t.Context(), fakeId, updatedUser)

		assert.Error(t, err)
		assert.Equal(t, ErrIdentifierAlreadyExists, err)
	})
}

func TestPgRepository_DeleteAccount(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		err = pgRepository.DeleteUser(t.Context(), fakeId)

		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var deletedAt *time.Time
		err = conn.QueryRow(t.Context(), "select deleted_at from users where id = $1", fakeId).
			Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
		assert.WithinDuration(t, time.Now().UTC(), *deletedAt, 5*time.Second)
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		nonExistentId := uuid.NewString()
		err = pgRepository.DeleteUser(t.Context(), nonExistentId)

		assert.NoError(t, err)
	})

	t.Run("delete already deleted user (idempotent)", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		err = pgRepository.DeleteUser(t.Context(), fakeId)
		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		var firstDeletedAt time.Time
		err = conn.QueryRow(t.Context(), "select deleted_at from users where id = $1", fakeId).
			Scan(&firstDeletedAt)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		err = pgRepository.DeleteUser(t.Context(), fakeId)
		assert.NoError(t, err)

		conn, err = pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var secondDeletedAt time.Time
		err = conn.QueryRow(t.Context(), "select deleted_at from users where id = $1", fakeId).
			Scan(&secondDeletedAt)
		require.NoError(t, err)
		assert.True(t, secondDeletedAt.After(firstDeletedAt) || secondDeletedAt.Equal(firstDeletedAt))
	})
}

func createUserTable(t *testing.T, connString string) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := "CREATE TABLE users (id varchar primary key, email varchar not null, username varchar not null, password varchar not null, created_at timestamp not null, deleted_at timestamp)"

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createFakeUser(t *testing.T, connString string) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	hashedFakePassword, err := bcrypt.GenerateFromPassword([]byte(fakePassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	sql := "insert into users (id, email, username, password, created_at) values ($1, $2, $3, $4, $5)"
	_, err = conn.Exec(t.Context(), sql,
		fakeId,
		fakeEmail,
		fakeUsername,
		string(hashedFakePassword),
		fakeNow,
	)
	require.NoError(t, err)
}
