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

	"hasir-api/internal/user"
	"hasir-api/pkg/config"
)

var (
	ErrFailedAcquireConnection = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	ErrIdentifierAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("email already exists"))
	ErrNoRows                  = connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	ErrRefreshTokenNotFound    = connect.NewError(connect.CodeNotFound, errors.New("refresh token not found"))
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

func (r *PgRepository) CreateUser(ctx context.Context, user *user.UserDTO) error {
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

func (r *PgRepository) GetUserByEmail(ctx context.Context, email string) (*user.UserDTO, error) {
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

	var userDTO user.UserDTO
	userDTO, err = pgx.CollectOneRow[user.UserDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &userDTO, nil
}

func (r *PgRepository) GetUserById(ctx context.Context, id string) (*user.UserDTO, error) {
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

	var userDTO user.UserDTO
	userDTO, err = pgx.CollectOneRow[user.UserDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &userDTO, nil
}

func (r *PgRepository) CreateRefreshToken(ctx context.Context, id, token string, expiresAt time.Time) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateRefreshToken", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "jti",
			Value: attribute.StringValue(token),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "insert into refresh_tokens (user_id, jti, created_at, expires_at) values (@UserId, @Jti, @CreatedAt, @ExpiresAt)"
	sqlArgs := pgx.NamedArgs{
		"UserId":    id,
		"Jti":       token,
		"CreatedAt": time.Now().UTC(),
		"ExpiresAt": expiresAt,
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)
		return ErrInternalServer
	}

	return nil
}

func (r *PgRepository) GetRefreshTokenByTokenId(ctx context.Context, token string) (*user.RefreshTokensDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRefreshTokenByTokenId", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "jti",
			Value: attribute.StringValue(token),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "select user_id as id, jti, expires_at, created_at from refresh_tokens where jti = $1"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, token)
	if err != nil {
		span.RecordError(err)
		return nil, ErrInternalServer
	}
	defer rows.Close()

	var refreshToken user.RefreshTokensDTO
	refreshToken, err = pgx.CollectOneRow[user.RefreshTokensDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &refreshToken, nil
}

func (r *PgRepository) DeleteRefreshToken(ctx context.Context, userId, token string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteRefreshToken", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(userId),
		},
		attribute.KeyValue{
			Key:   "jti",
			Value: attribute.StringValue(token),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "delete from refresh_tokens where user_id = @UserId and jti = @Jti"
	sqlArgs := pgx.NamedArgs{
		"UserId": userId,
		"Jti":    token,
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)
		return ErrInternalServer
	}

	return nil
}

func (r *PgRepository) UpdateUserById(ctx context.Context, id string, user *user.UserDTO) error {
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

func (r *PgRepository) CreateApiKey(ctx context.Context, userId, apiKey string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateApiKey")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		INSERT INTO api_keys (id, user_id, name, key, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
	`

	_, err = connection.Exec(
		ctx,
		sql,
		userId,
		"default", // Default name for the API key
		apiKey,
		time.Now().UTC(),
	)

	if err != nil {
		span.RecordError(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == ErrUniqueViolationCode {
			return connect.NewError(connect.CodeAlreadyExists, errors.New("api key already exists"))
		}
		return ErrInternalServer
	}

	return nil
}

func (r *PgRepository) GetApiKeys(ctx context.Context, userId string, page, pageSize int) (*[]user.ApiKeyDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetApiKeys")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	offset := (page - 1) * pageSize

	sql := `
		SELECT id, user_id, name, key, created_at
		FROM api_keys
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := connection.Query(ctx, sql, userId, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, ErrInternalServer
	}
	defer rows.Close()

	var apiKeys []user.ApiKeyDTO
	for rows.Next() {
		var key user.ApiKeyDTO
		err = rows.Scan(
			&key.Id,
			&key.UserId,
			&key.Name,
			&key.Key,
			&key.CreatedAt,
		)
		if err != nil {
			span.RecordError(err)
			return nil, ErrInternalServer
		}
		apiKeys = append(apiKeys, key)
	}

	if err = rows.Err(); err != nil {
		span.RecordError(err)
		return nil, ErrInternalServer
	}

	return &apiKeys, nil
}

func (r *PgRepository) GetApiKeysCount(ctx context.Context, userId string) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetApiKeysCount")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		SELECT COUNT(*)
		FROM api_keys
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var count int
	err = connection.QueryRow(ctx, sql, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, ErrInternalServer
	}

	return count, nil
}

func (r *PgRepository) RevokeApiKey(ctx context.Context, userId, keyId string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "RevokeApiKey")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		UPDATE api_keys 
		SET deleted_at = $1 
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	result, err := connection.Exec(
		ctx,
		sql,
		time.Now().UTC(),
		keyId,
		userId,
	)

	if err != nil {
		span.RecordError(err)
		return ErrInternalServer
	}

	if result.RowsAffected() == 0 {
		return connect.NewError(connect.CodeNotFound, errors.New("api key not found or already revoked"))
	}

	return nil
}

func (r *PgRepository) CreateSshKey(ctx context.Context, userId, publicKey string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateSshKey")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		INSERT INTO ssh_keys (id, user_id, name, public_key, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
	`

	_, err = connection.Exec(
		ctx,
		sql,
		userId,
		"default", // Default name for the SSH key
		publicKey,
		time.Now().UTC(),
	)

	if err != nil {
		span.RecordError(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == ErrUniqueViolationCode {
			return connect.NewError(connect.CodeAlreadyExists, errors.New("ssh key already exists"))
		}
		return ErrInternalServer
	}

	return nil
}

func (r *PgRepository) GetSshKeys(ctx context.Context, userId string, page, pageSize int) (*[]user.SshKeyDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetSshKeys")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	offset := (page - 1) * pageSize

	sql := `
		SELECT id, user_id, name, public_key, created_at
		FROM ssh_keys
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := connection.Query(ctx, sql, userId, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, ErrInternalServer
	}
	defer rows.Close()

	var sshKeys []user.SshKeyDTO
	for rows.Next() {
		var key user.SshKeyDTO
		err = rows.Scan(
			&key.Id,
			&key.UserId,
			&key.Name,
			&key.PublicKey,
			&key.CreatedAt,
		)
		if err != nil {
			span.RecordError(err)
			return nil, ErrInternalServer
		}
		sshKeys = append(sshKeys, key)
	}

	if err = rows.Err(); err != nil {
		span.RecordError(err)
		return nil, ErrInternalServer
	}

	return &sshKeys, nil
}

func (r *PgRepository) GetSshKeysCount(ctx context.Context, userId string) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetSshKeysCount")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		SELECT COUNT(*)
		FROM ssh_keys
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var count int
	err = connection.QueryRow(ctx, sql, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, ErrInternalServer
	}

	return count, nil
}

func (r *PgRepository) RevokeSshKey(ctx context.Context, userId, keyId string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "RevokeSshKey")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		UPDATE ssh_keys 
		SET deleted_at = $1 
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	result, err := connection.Exec(
		ctx,
		sql,
		time.Now().UTC(),
		keyId,
		userId,
	)

	if err != nil {
		span.RecordError(err)
		return ErrInternalServer
	}

	if result.RowsAffected() == 0 {
		return connect.NewError(connect.CodeNotFound, errors.New("ssh key not found or already revoked"))
	}

	return nil
}
