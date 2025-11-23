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

	"apps/api/pkg/config"
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

		user := &UserDTO{
			Id:        uuid.NewString(),
			Username:  "test",
			Email:     "test@mail.com",
			Password:  "Asdfg123456_",
			CreatedAt: time.Now().UTC(),
		}

		err = pgRepository.CreateUser(t.Context(), user)

		assert.NoError(t, err)
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
