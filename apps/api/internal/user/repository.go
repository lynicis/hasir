package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"hasir-api/pkg/config"
)

type Repository interface {
	CreateUser(ctx context.Context, user *UserDTO) error
	GetUserByEmail(ctx context.Context, email string) (*UserDTO, error)
	GetUserById(ctx context.Context, id string) (*UserDTO, error)
	CreateRefreshToken(ctx context.Context, id, token string, expiresAt time.Time) error
	UpdateUserById(ctx context.Context, id string, user *UserDTO) error
	DeleteUser(ctx context.Context, userId string) error
}

var (
	ErrFailedAcquireConnection = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	ErrIdentifierAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("email already exists"))
	ErrNoRows                  = connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	ErrInternalServer          = connect.NewError(connect.CodeInternal, errors.New("something went wrong"))
	ErrUniqueViolationCode     = "23505"
)

type PgRepository struct {
	connectionPool *pgxpool.Pool
	tracer         trace.Tracer
}

func NewPgRepository(
	cfg *config.Config,
	traceProvider *sdktrace.TracerProvider,
) *PgRepository {
	credential := cfg.PostgresConfig.GetPostgresDsn()
	pgConfig, err := pgxpool.ParseConfig(credential)
	if err != nil {
		zap.L().Fatal("failed to parse database config", zap.Error(err))
	}

	if traceProvider != nil {
		pgConfig.ConnConfig.Tracer = otelpgx.NewTracer(
			otelpgx.WithTracerProvider(traceProvider),
			otelpgx.WithDisableConnectionDetailsInAttributes(),
		)
	}

	var pgConnectionPool *pgxpool.Pool
	pgConnectionPool, err = pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		zap.L().Fatal("failed to connect database", zap.Error(err))
	}

	if traceProvider != nil {
		if err := otelpgx.RecordStats(pgConnectionPool); err != nil {
			zap.L().Fatal("unable to record database stats", zap.Error(err))
		}
	}

	var connection *pgxpool.Conn
	connection, err = pgConnectionPool.Acquire(context.Background())
	if err != nil {
		zap.L().Fatal("failed to acquire connection", zap.Error(err))
	}
	defer connection.Release()

	err = connection.Ping(context.Background())
	if err != nil {
		zap.L().Fatal("failed to ping database", zap.Error(err))
	}

	var tracer trace.Tracer
	if traceProvider != nil {
		tracer = traceProvider.Tracer("UserPostgreSQLRepository")
	} else {
		tracer = noop.NewTracerProvider().Tracer("UserPostgreSQLRepository")
	}

	return &PgRepository{
		connectionPool: pgConnectionPool,
		tracer:         tracer,
	}
}

func (r *PgRepository) CreateUser(ctx context.Context, user *UserDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateUser", trace.WithAttributes(attribute.KeyValue{
		Key:   "newUser",
		Value: attribute.StringValue(fmt.Sprintf("%+v", user)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "insert into users (id, username, email, password, created_at) values (@Id, @Username, @Email, @Password, @CreatedAt)"
	sqlArgs := pgx.NamedArgs{
		"Id":        user.Id,
		"Username":  user.Username,
		"Email":     user.Email,
		"Password":  user.Password,
		"CreatedAt": time.Now().UTC(),
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == ErrUniqueViolationCode {
			return ErrIdentifierAlreadyExists
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to execute insert user query"))
	}

	return nil
}

func (r *PgRepository) GetUserByEmail(ctx context.Context, email string) (*UserDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetUserByEmail", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "email",
			Value: attribute.StringValue(email),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "select * from users where email = $1"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, email)
	if err != nil {
		span.RecordError(err)
		return nil, ErrInternalServer
	}
	defer rows.Close()

	var user UserDTO
	user, err = pgx.CollectOneRow[UserDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &user, nil
}

func (r *PgRepository) GetUserById(ctx context.Context, id string) (*UserDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetUserById", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "select * from users where id = $1"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, id)
	if err != nil {
		span.RecordError(err)
		return nil, ErrInternalServer
	}
	defer rows.Close()

	var user UserDTO
	user, err = pgx.CollectOneRow[UserDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &user, nil
}

func (r *PgRepository) CreateRefreshToken(ctx context.Context, id, token string, expiresAt time.Time) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateRefreshToken", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "token",
			Value: attribute.StringValue(token),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "insert into refresh_tokens (user_id, token, created_at, expires_at) values (@UserId, @Token, @CreatedAt, @ExpiresAt)"
	sqlArgs := pgx.NamedArgs{
		"UserId":    id,
		"Token":     token,
		"CreatedAt": time.Now().UTC(),
		"ExpiresAt": expiresAt,
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)
		return ErrInternalServer
	}

	return nil
}

func (r *PgRepository) UpdateUserById(ctx context.Context, id string, user *UserDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateUserById", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "updatedUser",
			Value: attribute.StringValue(fmt.Sprintf("%+v", user)),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	var setParts []string
	sqlArgs := pgx.NamedArgs{"Id": id}

	if user.Username != "" {
		setParts = append(setParts, "username = @Username")
		sqlArgs["Username"] = user.Username
	}
	if user.Email != "" {
		setParts = append(setParts, "email = @Email")
		sqlArgs["Email"] = user.Email
	}
	if user.Password != "" {
		setParts = append(setParts, "password = @Password")
		sqlArgs["Password"] = user.Password
	}

	if len(setParts) == 0 {

		sql := "select id from users where id = @Id"
		var existingId string
		err = connection.QueryRow(ctx, sql, sqlArgs).Scan(&existingId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNoRows
			}
			return connect.NewError(connect.CodeInternal, errors.New("failed to verify user existence"))
		}
		return nil
	}

	sql := fmt.Sprintf("update users set %s where id = @Id", strings.Join(setParts, ", "))

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrIdentifierAlreadyExists
			}

			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to execute update user query"))
	}

	if result.RowsAffected() == 0 {
		return ErrNoRows
	}

	return nil
}

func (r *PgRepository) DeleteUser(ctx context.Context, userId string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteAccount", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(userId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "update users set deleted_at = @DeletedAt where id = @UserId"
	sqlArgs := pgx.NamedArgs{
		"UserId":    userId,
		"DeletedAt": time.Now().UTC(),
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)
		return ErrInternalServer
	}

	return nil
}
