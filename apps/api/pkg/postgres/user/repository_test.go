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

	"hasir-api/internal/user"
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

	assert.Implements(t, (*user.Repository)(nil), pgRepository)
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
		user := &user.UserDTO{
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

		user := &user.UserDTO{
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

		user := &user.UserDTO{
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

		user := &user.UserDTO{
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

		assert.NotEmpty(t, user.Password)

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

		user, err := pgRepository.GetUserByEmail(t.Context(), "TEST@MAIL.COM")

		if err != nil {
			assert.Error(t, err)
			assert.Nil(t, user)
		} else {

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

		updatedUser := &user.UserDTO{
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

		updatedUser := &user.UserDTO{
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

		updatedUser := &user.UserDTO{
			Id:    fakeId,
			Email: "only-email-updated@mail.com",
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

		updatedUser := &user.UserDTO{
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
		updatedUser := &user.UserDTO{
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

		updatedUser := &user.UserDTO{
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

func TestPgRepository_CreateRefreshToken(t *testing.T) {
	t.Run("stores record with jti", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createRefreshTokensTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		jti := uuid.NewString()
		expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

		err = pgRepository.CreateRefreshToken(t.Context(), fakeId, jti, expiresAt)
		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbUserID, dbJti string
		var dbExpiresAt, dbCreatedAt time.Time
		err = conn.QueryRow(t.Context(),
			"select user_id, jti, expires_at, created_at from refresh_tokens where user_id = $1 and jti = $2",
			fakeId, jti).
			Scan(&dbUserID, &dbJti, &dbExpiresAt, &dbCreatedAt)
		require.NoError(t, err)

		assert.Equal(t, fakeId, dbUserID)
		assert.Equal(t, jti, dbJti)
		assert.WithinDuration(t, expiresAt, dbExpiresAt, 5*time.Second)
		assert.WithinDuration(t, time.Now().UTC(), dbCreatedAt, 5*time.Second)
	})

}

func TestPgRepository_GetRefreshTokenByTokenId(t *testing.T) {
	t.Run("returns record by jti", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createRefreshTokensTable(t, connString)

		jti := uuid.NewString()
		expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(),
			"insert into refresh_tokens (user_id, jti, expires_at, created_at) values ($1, $2, $3, $4)",
			fakeId, jti, expiresAt, fakeNow)
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

		rec, err := pgRepository.GetRefreshTokenByTokenId(t.Context(), jti)

		assert.NoError(t, err)
		require.NotNil(t, rec)
		assert.Equal(t, fakeId, rec.UserId)
		assert.Equal(t, jti, rec.Jti)
		assert.WithinDuration(t, expiresAt, rec.ExpiresAt, time.Second)
		assert.WithinDuration(t, fakeNow, rec.CreatedAt, time.Second)
	})

	t.Run("returns ErrRefreshTokenNotFound when jti does not exist", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createRefreshTokensTable(t, connString)

		missingJti := uuid.NewString()

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		rec, err := pgRepository.GetRefreshTokenByTokenId(t.Context(), missingJti)

		assert.Error(t, err)
		assert.Equal(t, ErrRefreshTokenNotFound, err)
		assert.Nil(t, rec)
	})

}

func TestPgRepository_DeleteRefreshToken(t *testing.T) {
	t.Run("removes record by user and jti", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createRefreshTokensTable(t, connString)

		jti := uuid.NewString()

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(),
			"insert into refresh_tokens (user_id, jti, expires_at, created_at) values ($1, $2, $3, $4)",
			fakeId, jti, time.Now().UTC().Add(7*24*time.Hour), fakeNow)
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

		err = pgRepository.DeleteRefreshToken(t.Context(), fakeId, jti)
		assert.NoError(t, err)

		var count int
		err = conn.QueryRow(t.Context(),
			"select count(*) from refresh_tokens where user_id = $1 and jti = $2",
			fakeId, jti).
			Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestPgRepository_CreateApiKey(t *testing.T) {
	t.Run("creates a new API key", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createApiKeysTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		apiKey := "test-api-key-123"
		err = pgRepository.CreateApiKey(t.Context(), fakeId, "test-api-key", apiKey)

		assert.NoError(t, err)
		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var count int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM api_keys WHERE user_id = $1 AND key = $2 AND deleted_at IS NULL",
			fakeId, apiKey).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("returns error for duplicate API key", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createApiKeysTable(t, connString)
		apiKey := "duplicate-api-key"
		keyId := uuid.NewString()
		createFakeApiKey(t, connString, fakeId, keyId, apiKey)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)
		err = pgRepository.CreateApiKey(t.Context(), fakeId, "duplicate-api-key-name", apiKey)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestPgRepository_GetApiKeys(t *testing.T) {
	t.Run("returns all API keys for a user", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createApiKeysTable(t, connString)
		key1Id := uuid.NewString()
		key1Value := "test-key-1"
		createFakeApiKey(t, connString, fakeId, key1Id, key1Value)

		key2Id := uuid.NewString()
		key2Value := "test-key-2"
		createFakeApiKey(t, connString, fakeId, key2Id, key2Value)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		keys, err := pgRepository.GetApiKeys(t.Context(), fakeId, 1, 10)

		assert.NoError(t, err)
		require.NotNil(t, keys)
		assert.Len(t, *keys, 2)
		assert.Equal(t, key2Id, (*keys)[0].Id)
		assert.Equal(t, key2Value, (*keys)[0].Key)
		assert.Equal(t, key1Id, (*keys)[1].Id)
		assert.Equal(t, key1Value, (*keys)[1].Key)
	})

	t.Run("returns empty slice when no API keys exist", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createApiKeysTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		keys, err := pgRepository.GetApiKeys(t.Context(), fakeId, 1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, keys)
		assert.Empty(t, *keys)
	})
}

func TestPgRepository_RevokeApiKey(t *testing.T) {
	t.Run("marks API key as deleted", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createApiKeysTable(t, connString)
		keyId := uuid.NewString()
		keyValue := "test-key-to-revoke"
		createFakeApiKey(t, connString, fakeId, keyId, keyValue)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)
		err = pgRepository.RevokeApiKey(t.Context(), fakeId, keyId)
		assert.NoError(t, err)
		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var deletedAt *time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT deleted_at FROM api_keys WHERE id = $1",
			keyId).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createApiKeysTable(t, connString)

		nonExistentKeyId := uuid.NewString()

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		err = pgRepository.RevokeApiKey(t.Context(), fakeId, nonExistentKeyId)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestPgRepository_CreateSshKey(t *testing.T) {
	t.Run("creates a new SSH key", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createSshKeysTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		publicKey := "ssh-rsa AAAAB3NzaC1yc2E..."
		err = pgRepository.CreateSshKey(t.Context(), fakeId, "test-ssh-key", publicKey)

		assert.NoError(t, err)
		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var count int
		err = conn.QueryRow(t.Context(),
			"SELECT COUNT(*) FROM ssh_keys WHERE user_id = $1 AND public_key = $2 AND deleted_at IS NULL",
			fakeId, publicKey).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestPgRepository_GetSshKeys(t *testing.T) {
	t.Run("returns all SSH keys for a user", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createSshKeysTable(t, connString)
		key1Id := uuid.NewString()
		key1Value := "ssh-rsa key1..."
		createFakeSshKey(t, connString, fakeId, key1Id, key1Value)

		key2Id := uuid.NewString()
		key2Value := "ssh-rsa key2..."
		createFakeSshKey(t, connString, fakeId, key2Id, key2Value)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		keys, err := pgRepository.GetSshKeys(t.Context(), fakeId, 1, 10)

		assert.NoError(t, err)
		require.NotNil(t, keys)
		assert.Len(t, *keys, 2)
		assert.Equal(t, key2Id, (*keys)[0].Id)
		assert.Equal(t, key2Value, (*keys)[0].PublicKey)
		assert.Equal(t, key1Id, (*keys)[1].Id)
		assert.Equal(t, key1Value, (*keys)[1].PublicKey)
	})
}

func TestPgRepository_RevokeSshKey(t *testing.T) {
	t.Run("marks SSH key as deleted", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createSshKeysTable(t, connString)
		keyId := uuid.NewString()
		keyValue := "ssh-rsa key-to-revoke..."
		createFakeSshKey(t, connString, fakeId, keyId, keyValue)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)
		err = pgRepository.RevokeSshKey(t.Context(), fakeId, keyId)
		assert.NoError(t, err)
		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var deletedAt *time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT deleted_at FROM ssh_keys WHERE id = $1",
			keyId).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createSshKeysTable(t, connString)

		nonExistentKeyId := uuid.NewString()

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		err = pgRepository.RevokeSshKey(t.Context(), fakeId, nonExistentKeyId)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
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

func createRefreshTokensTable(t *testing.T, connString string) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := "CREATE TABLE refresh_tokens (user_id varchar not null, jti varchar not null, expires_at timestamp not null, created_at timestamp not null, PRIMARY KEY (user_id, jti))"

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createApiKeysTable(t *testing.T, connString string) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `
		CREATE TABLE api_keys (
			id varchar PRIMARY KEY,
			user_id varchar NOT NULL,
			name varchar NOT NULL,
			key varchar NOT NULL,
			created_at timestamp NOT NULL,
			deleted_at timestamp
		);
		CREATE UNIQUE INDEX idx_api_keys_key ON api_keys(key) WHERE deleted_at IS NULL;
	`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createSshKeysTable(t *testing.T, connString string) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `
		CREATE TABLE ssh_keys (
			id varchar PRIMARY KEY,
			user_id varchar NOT NULL,
			name varchar NOT NULL,
			public_key text NOT NULL,
			created_at timestamp NOT NULL,
			deleted_at timestamp
		);
	`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createFakeApiKey(t *testing.T, connString, userId, keyId, keyValue string) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `
		INSERT INTO api_keys (id, user_id, name, key, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = conn.Exec(t.Context(), sql,
		keyId,
		userId,
		"test-key",
		keyValue,
		time.Now().UTC().Add(-24*time.Hour),
	)
	require.NoError(t, err)
}

func createFakeSshKey(t *testing.T, connString, userId, keyId, publicKey string) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `
		INSERT INTO ssh_keys (id, user_id, name, public_key, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = conn.Exec(t.Context(), sql,
		keyId,
		userId,
		"test-key",
		publicKey,
		time.Now().UTC().Add(-24*time.Hour),
	)
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

func TestPgRepository_GetUserBySshPublicKey(t *testing.T) {
	t.Run("returns user for valid SSH key", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createSshKeysTable(t, connString)

		keyId := uuid.NewString()
		publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABtest"
		createFakeSshKey(t, connString, fakeId, keyId, publicKey)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user, err := pgRepository.GetUserBySshPublicKey(t.Context(), publicKey)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, fakeId, user.Id)
		assert.Equal(t, fakeEmail, user.Email)
		assert.Equal(t, fakeUsername, user.Username)
	})

	t.Run("returns error for non-existent SSH key", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createSshKeysTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user, err := pgRepository.GetUserBySshPublicKey(t.Context(), "ssh-rsa nonexistent")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "ssh key not found")
	})

	t.Run("returns error for revoked SSH key", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createSshKeysTable(t, connString)

		keyId := uuid.NewString()
		publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABrevoked"
		createFakeSshKey(t, connString, fakeId, keyId, publicKey)

		// Revoke the SSH key
		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		_, err = conn.Exec(t.Context(), "UPDATE ssh_keys SET deleted_at = $1 WHERE id = $2", time.Now().UTC(), keyId)
		require.NoError(t, err)
		_ = conn.Close(t.Context())

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		user, err := pgRepository.GetUserBySshPublicKey(t.Context(), publicKey)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "ssh key not found")
	})
}

func TestPgRepository_CreatePasswordResetToken(t *testing.T) {
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
		createPasswordResetTokensTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		token := uuid.NewString()
		expiresAt := time.Now().UTC().Add(1 * time.Hour)

		err = pgRepository.CreatePasswordResetToken(t.Context(), fakeId, token, expiresAt)

		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var dbToken, dbUserId string
		var dbExpiresAt, dbCreatedAt time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT user_id, token, expires_at, created_at FROM password_reset_tokens WHERE token = $1", token).
			Scan(&dbUserId, &dbToken, &dbExpiresAt, &dbCreatedAt)
		require.NoError(t, err)

		assert.Equal(t, fakeId, dbUserId)
		assert.Equal(t, token, dbToken)
		assert.WithinDuration(t, expiresAt, dbExpiresAt, time.Second)
	})

	t.Run("duplicate token returns error", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createPasswordResetTokensTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		token := uuid.NewString()
		expiresAt := time.Now().UTC().Add(1 * time.Hour)

		err = pgRepository.CreatePasswordResetToken(t.Context(), fakeId, token, expiresAt)
		assert.NoError(t, err)

		err = pgRepository.CreatePasswordResetToken(t.Context(), fakeId, token, expiresAt)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reset token already exists")
	})
}

func TestPgRepository_GetPasswordResetToken(t *testing.T) {
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
		createPasswordResetTokensTable(t, connString)

		token := uuid.NewString()
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		createFakePasswordResetToken(t, connString, fakeId, token, expiresAt, nil)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		resetToken, err := pgRepository.GetPasswordResetToken(t.Context(), token)

		assert.NoError(t, err)
		assert.NotNil(t, resetToken)
		assert.Equal(t, fakeId, resetToken.UserId)
		assert.Equal(t, token, resetToken.Token)
		assert.WithinDuration(t, expiresAt, resetToken.ExpiresAt, time.Second)
		assert.Nil(t, resetToken.UsedAt)
	})

	t.Run("token not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createPasswordResetTokensTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		resetToken, err := pgRepository.GetPasswordResetToken(t.Context(), "nonexistent-token")

		assert.Error(t, err)
		assert.Nil(t, resetToken)
		assert.Contains(t, err.Error(), "reset token not found")
	})

	t.Run("returns used token", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createPasswordResetTokensTable(t, connString)

		token := uuid.NewString()
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		usedAt := time.Now().UTC().Add(-10 * time.Minute)
		createFakePasswordResetToken(t, connString, fakeId, token, expiresAt, &usedAt)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		resetToken, err := pgRepository.GetPasswordResetToken(t.Context(), token)

		assert.NoError(t, err)
		assert.NotNil(t, resetToken)
		assert.NotNil(t, resetToken.UsedAt)
		assert.WithinDuration(t, usedAt, *resetToken.UsedAt, time.Second)
	})
}

func TestPgRepository_MarkPasswordResetTokenAsUsed(t *testing.T) {
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
		createPasswordResetTokensTable(t, connString)

		token := uuid.NewString()
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		createFakePasswordResetToken(t, connString, fakeId, token, expiresAt, nil)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		err = pgRepository.MarkPasswordResetTokenAsUsed(t.Context(), token)

		assert.NoError(t, err)

		conn, err := pgx.Connect(t.Context(), connString)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(t.Context())
		}()

		var usedAt *time.Time
		err = conn.QueryRow(t.Context(),
			"SELECT used_at FROM password_reset_tokens WHERE token = $1", token).
			Scan(&usedAt)
		require.NoError(t, err)

		assert.NotNil(t, usedAt)
		assert.WithinDuration(t, time.Now().UTC(), *usedAt, 2*time.Second)
	})

	t.Run("token not found", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createPasswordResetTokensTable(t, connString)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		err = pgRepository.MarkPasswordResetTokenAsUsed(t.Context(), "nonexistent-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reset token not found or already used")
	})

	t.Run("already used token returns error", func(t *testing.T) {
		container := setupPgContainer(t)
		defer func() {
			err := container.Terminate(t.Context())
			require.NoError(t, err)
		}()

		connString, err := container.ConnectionString(t.Context())
		require.NoError(t, err)

		createUserTable(t, connString)
		createFakeUser(t, connString)
		createPasswordResetTokensTable(t, connString)

		token := uuid.NewString()
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		usedAt := time.Now().UTC().Add(-10 * time.Minute)
		createFakePasswordResetToken(t, connString, fakeId, token, expiresAt, &usedAt)

		traceProvider := sdktrace.NewTracerProvider()
		pgRepository := NewPgRepository(&config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: connString,
			},
		}, traceProvider)

		err = pgRepository.MarkPasswordResetTokenAsUsed(t.Context(), token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reset token not found or already used")
	})
}

func createPasswordResetTokensTable(t *testing.T, connString string) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `
		CREATE TABLE password_reset_tokens (
			id varchar PRIMARY KEY DEFAULT gen_random_uuid()::varchar,
			user_id varchar NOT NULL,
			token varchar NOT NULL UNIQUE,
			expires_at timestamp NOT NULL,
			created_at timestamp NOT NULL,
			used_at timestamp
		);
	`

	_, err = conn.Exec(t.Context(), sql)
	require.NoError(t, err)
}

func createFakePasswordResetToken(t *testing.T, connString, userId, token string, expiresAt time.Time, usedAt *time.Time) {
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err)
	defer func() {
		err = conn.Close(t.Context())
		require.NoError(t, err)
	}()

	sql := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, created_at, used_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = conn.Exec(t.Context(), sql,
		userId,
		token,
		expiresAt,
		time.Now().UTC().Add(-30*time.Minute),
		usedAt,
	)
	require.NoError(t, err)
}
