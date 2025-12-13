package registry

import (
	"context"
	"errors"
	"fmt"
	"time"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"connectrpc.com/connect"
	"github.com/exaring/otelpgx"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	"hasir-api/internal/registry"
	"hasir-api/pkg/config"
)

var (
	ErrRepositoryAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("repository already exists"))
	ErrRepositoryNotFound      = connect.NewError(connect.CodeNotFound, errors.New("repository not found"))
	ErrFailedAcquireConnection = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
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
		tracer = traceProvider.Tracer("RepositoryPostgreSQLRepository")
	} else {
		tracer = noop.NewTracerProvider().Tracer("RepositoryPostgreSQLRepository")
	}

	return &PgRepository{
		connectionPool: pgConnectionPool,
		tracer:         tracer,
	}
}

func (r *PgRepository) GetConnectionPool() *pgxpool.Pool {
	return r.connectionPool
}

func (r *PgRepository) GetTracer() trace.Tracer {
	return r.tracer
}

func (r *PgRepository) CreateRepository(ctx context.Context, repo *registry.RepositoryDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateRepository", trace.WithAttributes(attribute.KeyValue{
		Key:   "newRepository",
		Value: attribute.StringValue(fmt.Sprintf("%+v", repo)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	now := time.Now().UTC()
	sql := `INSERT INTO repositories (id, name, created_by, organization_id, path, visibility, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err = connection.Exec(ctx, sql,
		repo.Id,
		repo.Name,
		repo.CreatedBy,
		repo.OrganizationId,
		repo.Path,
		repo.Visibility,
		now,
		&now,
	); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == ErrUniqueViolationCode {
			return ErrRepositoryAlreadyExists
		}

		return connect.NewError(
			connect.CodeInternal,
			errors.New("failed to execute insert repository query"),
		)
	}

	return nil
}

func (r *PgRepository) GetRepositoryByName(ctx context.Context, name string) (*registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoryByName", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "name",
			Value: attribute.StringValue(name),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM repositories WHERE name = $1 AND deleted_at IS NULL"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, name)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repository by name"))
	}
	defer rows.Close()

	var repo registry.RepositoryDTO
	repo, err = pgx.CollectOneRow[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositoryNotFound
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &repo, nil
}

func (r *PgRepository) GetRepositories(ctx context.Context, page, pageSize int) (*[]registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositories", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "page",
			Value: attribute.IntValue(page),
		},
		attribute.KeyValue{
			Key:   "pageSize",
			Value: attribute.IntValue(pageSize),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	offset := (page - 1) * pageSize
	sql := "SELECT * FROM repositories WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repositories"))
	}
	defer rows.Close()

	repos, err := pgx.CollectRows[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect rows"))
	}

	return &repos, nil
}

func (r *PgRepository) GetRepositoriesByUser(ctx context.Context, userId string, page, pageSize int) (*[]registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesByUser", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
		attribute.KeyValue{
			Key:   "page",
			Value: attribute.IntValue(page),
		},
		attribute.KeyValue{
			Key:   "pageSize",
			Value: attribute.IntValue(pageSize),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	offset := (page - 1) * pageSize
	sql := `
		SELECT r.*
		FROM repositories r
		INNER JOIN organization_members m ON m.organization_id = r.organization_id
		WHERE m.user_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := connection.Query(ctx, sql, userId, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repositories by user"))
	}
	defer rows.Close()

	repos, err := pgx.CollectRows[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect repositories by user rows"))
	}

	return &repos, nil
}

func (r *PgRepository) GetRepositoriesByOrganizationId(ctx context.Context, organizationId string) (*[]registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesByOrganizationId", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organization_id",
			Value: attribute.StringValue(organizationId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM repositories WHERE organization_id = $1 AND deleted_at IS NULL"

	rows, err := connection.Query(ctx, sql, organizationId)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repositories by organization id"))
	}
	defer rows.Close()

	repos, err := pgx.CollectRows[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect rows"))
	}

	return &repos, nil
}

func (r *PgRepository) GetRepositoriesCount(ctx context.Context) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesCount")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT COUNT(*) FROM repositories WHERE deleted_at IS NULL"

	var count int
	err = connection.QueryRow(ctx, sql).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count repositories"))
	}

	return count, nil
}

func (r *PgRepository) GetRepositoriesByUserCount(ctx context.Context, userId string) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesByUserCount", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		SELECT COUNT(*)
		FROM repositories r
		INNER JOIN organization_members m ON m.organization_id = r.organization_id
		WHERE m.user_id = $1 AND r.deleted_at IS NULL`

	var count int
	err = connection.QueryRow(ctx, sql, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count repositories by user"))
	}

	return count, nil
}

func (r *PgRepository) GetRepositoriesByUserAndOrganization(ctx context.Context, userId, organizationId string, page, pageSize int) (*[]registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesByUserAndOrganization", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
		attribute.KeyValue{
			Key:   "page",
			Value: attribute.IntValue(page),
		},
		attribute.KeyValue{
			Key:   "pageSize",
			Value: attribute.IntValue(pageSize),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	offset := (page - 1) * pageSize
	sql := `
		SELECT r.*
		FROM repositories r
		INNER JOIN organization_members m ON m.organization_id = r.organization_id
		WHERE m.user_id = $1 AND r.organization_id = $2 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := connection.Query(ctx, sql, userId, organizationId, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repositories by user and organization"))
	}
	defer rows.Close()

	repos, err := pgx.CollectRows[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect repositories by user and organization rows"))
	}

	return &repos, nil
}

func (r *PgRepository) GetRepositoriesByUserAndOrganizationCount(ctx context.Context, userId, organizationId string) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoriesByUserAndOrganizationCount", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		SELECT COUNT(*)
		FROM repositories r
		INNER JOIN organization_members m ON m.organization_id = r.organization_id
		WHERE m.user_id = $1 AND r.organization_id = $2 AND r.deleted_at IS NULL`

	var count int
	err = connection.QueryRow(ctx, sql, userId, organizationId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count repositories by user and organization"))
	}

	return count, nil
}

func (r *PgRepository) GetRepositoryById(ctx context.Context, id string) (*registry.RepositoryDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetRepositoryById", trace.WithAttributes(
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

	sql := "SELECT * FROM repositories WHERE id = $1 AND deleted_at IS NULL"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, id)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query repository by id"))
	}
	defer rows.Close()

	var repo registry.RepositoryDTO
	repo, err = pgx.CollectOneRow[registry.RepositoryDTO](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRepositoryNotFound
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &repo, nil
}

func (r *PgRepository) UpdateRepository(ctx context.Context, repo *registry.RepositoryDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateRepository", trace.WithAttributes(attribute.KeyValue{
		Key:   "repository",
		Value: attribute.StringValue(fmt.Sprintf("%+v", repo)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	now := time.Now().UTC()
	sql := `UPDATE repositories
			SET name = $1, updated_at = $2
			WHERE id = $3 AND deleted_at IS NULL`

	result, err := connection.Exec(ctx, sql, repo.Name, &now, repo.Id)
	if err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrRepositoryAlreadyExists
			}
			return connect.NewError(connect.CodeInternal, err)
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to execute update repository query"))
	}

	if result.RowsAffected() == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}

func (r *PgRepository) DeleteRepository(ctx context.Context, id string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteRepository", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	now := time.Now().UTC()
	sql := `UPDATE repositories
			SET deleted_at = $1
			WHERE id = $2 AND deleted_at IS NULL`

	result, err := connection.Exec(ctx, sql, &now, id)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(
			connect.CodeInternal,
			errors.New("failed to execute delete repository query"),
		)
	}

	if result.RowsAffected() == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}

func (r *PgRepository) DeleteRepositoriesByOrganizationId(ctx context.Context, organizationId string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteRepositoriesByOrganizationId", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	now := time.Now().UTC()
	sql := `UPDATE repositories
			SET deleted_at = $1
			WHERE organization_id = $2 AND deleted_at IS NULL`

	_, err = connection.Exec(ctx, sql, &now, organizationId)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(
			connect.CodeInternal,
			errors.New("failed to execute batch delete repositories query"),
		)
	}

	return nil
}

func (r *PgRepository) UpdateSdkPreferences(
	ctx context.Context,
	repositoryId string,
	preferences []registry.SdkPreferencesDTO,
) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateSdkPreferences", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "repositoryId",
			Value: attribute.StringValue(repositoryId),
		},
		attribute.KeyValue{
			Key:   "preferences",
			Value: attribute.StringValue(fmt.Sprintf("%+v", preferences)),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	tx, err := connection.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to begin transaction"))
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	now := time.Now().UTC()
	upsertSQL := `
		INSERT INTO sdk_preferences (id, repository_id, sdk, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (repository_id, sdk)
		DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`

	for _, pref := range preferences {
		_, err = tx.Exec(ctx, upsertSQL,
			pref.Id,
			repositoryId,
			string(pref.Sdk),
			pref.Status,
			now,
			&now,
		)
		if err != nil {
			span.RecordError(err)
			return connect.NewError(
				connect.CodeInternal,
				errors.New("failed to upsert sdk preference"),
			)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to commit transaction"))
	}

	return nil
}

func (r *PgRepository) GetSdkPreferences(
	ctx context.Context,
	repositoryId string,
) ([]registry.SdkPreferencesDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetSdkPreferences", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "repositoryId",
			Value: attribute.StringValue(repositoryId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM sdk_preferences WHERE repository_id = $1"

	rows, err := connection.Query(ctx, sql, repositoryId)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("failed to query sdk preferences"),
		)
	}
	defer rows.Close()

	preferences, err := pgx.CollectRows[registry.SdkPreferencesDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("failed to collect sdk preferences rows"),
		)
	}

	return preferences, nil
}

func (r *PgRepository) GetSdkPreferencesByRepositoryIds(
	ctx context.Context,
	repositoryIds []string,
) (map[string][]registry.SdkPreferencesDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetSdkPreferencesByRepositoryIds", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "repositoryCount",
			Value: attribute.IntValue(len(repositoryIds)),
		},
	))
	defer span.End()

	if len(repositoryIds) == 0 {
		return make(map[string][]registry.SdkPreferencesDTO), nil
	}

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM sdk_preferences WHERE repository_id = ANY($1)"

	rows, err := connection.Query(ctx, sql, repositoryIds)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("failed to query sdk preferences"),
		)
	}
	defer rows.Close()

	preferences, err := pgx.CollectRows[registry.SdkPreferencesDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("failed to collect sdk preferences rows"),
		)
	}

	preferencesMap := make(map[string][]registry.SdkPreferencesDTO)
	for _, pref := range preferences {
		preferencesMap[pref.RepositoryId] = append(preferencesMap[pref.RepositoryId], pref)
	}

	return preferencesMap, nil
}

func (r *PgRepository) GetCommits(ctx context.Context, repoPath string) (*registryv1.GetCommitsResponse, error) {
	var span trace.Span
	_, span = r.tracer.Start(ctx, "GetCommits", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "repoPath",
			Value: attribute.StringValue(repoPath),
		},
	))
	defer span.End()

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to open git repository"))
	}

	ref, err := repo.Head()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get repository HEAD"))
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get commit log"))
	}
	defer commitIter.Close()

	var commits []*registryv1.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		commit := &registryv1.Commit{
			Id:      c.Hash.String(),
			Message: c.Message,
			User: &registryv1.Commit_User{
				Id:       c.Author.Email,
				Username: c.Author.Name,
			},
			CommitedAt: timestamppb.New(c.Author.When),
		}
		commits = append(commits, commit)
		return nil
	})
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to iterate commits"))
	}

	return &registryv1.GetCommitsResponse{
		Commits: commits,
	}, nil
}

func (r *PgRepository) GetFileTree(ctx context.Context, repoPath string, subPath *string) (*registryv1.GetFileTreeResponse, error) {
	var span trace.Span
	_, span = r.tracer.Start(ctx, "GetFileTree", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "repoPath",
			Value: attribute.StringValue(repoPath),
		},
	))
	defer span.End()

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to open git repository"))
	}

	ref, err := repo.Head()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get repository HEAD"))
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get commit object"))
	}

	tree, err := commit.Tree()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get tree object"))
	}

	var targetTree *object.Tree
	targetPath := ""
	if subPath != nil && *subPath != "" {
		targetPath = *subPath
		targetTree, err = tree.Tree(targetPath)
		if err != nil {
			span.RecordError(err)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("path not found in repository"))
		}
	} else {
		targetTree = tree
	}

	var nodes []*registryv1.FileTreeNode
	for _, entry := range targetTree.Entries {
		nodePath := entry.Name
		if targetPath != "" {
			nodePath = targetPath + "/" + entry.Name
		}

		var nodeType registryv1.NodeType
		var children []*registryv1.FileTreeNode

		if entry.Mode.IsFile() {
			nodeType = registryv1.NodeType_NODE_TYPE_FILE
		} else {
			nodeType = registryv1.NodeType_NODE_TYPE_DIRECTORY

			subTree, err := tree.Tree(nodePath)
			if err == nil {
				children = buildFileTreeNodes(tree, subTree, nodePath)
			}
		}

		node := &registryv1.FileTreeNode{
			Name:     entry.Name,
			Path:     nodePath,
			Type:     nodeType,
			Children: children,
		}
		nodes = append(nodes, node)
	}

	return &registryv1.GetFileTreeResponse{
		Nodes: nodes,
	}, nil
}

func buildFileTreeNodes(rootTree *object.Tree, tree *object.Tree, basePath string) []*registryv1.FileTreeNode {
	var nodes []*registryv1.FileTreeNode

	for _, entry := range tree.Entries {
		nodePath := basePath + "/" + entry.Name

		var nodeType registryv1.NodeType
		var children []*registryv1.FileTreeNode

		if entry.Mode.IsFile() {
			nodeType = registryv1.NodeType_NODE_TYPE_FILE
		} else {
			nodeType = registryv1.NodeType_NODE_TYPE_DIRECTORY

			subTree, err := rootTree.Tree(nodePath)
			if err == nil {
				children = buildFileTreeNodes(rootTree, subTree, nodePath)
			}
		}

		node := &registryv1.FileTreeNode{
			Name:     entry.Name,
			Path:     nodePath,
			Type:     nodeType,
			Children: children,
		}
		nodes = append(nodes, node)
	}

	return nodes
}

func (r *PgRepository) GetFilePreview(ctx context.Context, repoPath, filePath string) (*registryv1.GetFilePreviewResponse, error) {
	var span trace.Span
	_, span = r.tracer.Start(ctx, "GetFilePreview", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "repoPath",
			Value: attribute.StringValue(repoPath),
		},
		attribute.KeyValue{
			Key:   "filePath",
			Value: attribute.StringValue(filePath),
		},
	))
	defer span.End()

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to open git repository"))
	}

	ref, err := repo.Head()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get repository HEAD"))
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get commit object"))
	}

	tree, err := commit.Tree()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get tree object"))
	}

	file, err := tree.File(filePath)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("file not found in repository"))
	}

	content, err := file.Contents()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to read file content"))
	}

	mimeType := detectMimeType(filePath, []byte(content))
	size := file.Size

	return &registryv1.GetFilePreviewResponse{
		Content:  content,
		MimeType: mimeType,
		Size:     size,
	}, nil
}

var mimeTypeMap = map[string]string{
	".txt":        "text/plain",
	".md":         "text/markdown",
	".json":       "application/json",
	".yaml":       "application/yaml",
	".yml":        "application/yaml",
	".xml":        "application/xml",
	".html":       "text/html",
	".htm":        "text/html",
	".css":        "text/css",
	".js":         "application/javascript",
	".ts":         "application/typescript",
	".go":         "text/x-go",
	".py":         "text/x-python",
	".java":       "text/x-java",
	".c":          "text/x-c",
	".cpp":        "text/x-c++",
	".cc":         "text/x-c++",
	".cxx":        "text/x-c++",
	".h":          "text/x-c-header",
	".hpp":        "text/x-c-header",
	".rs":         "text/x-rust",
	".sh":         "application/x-sh",
	".sql":        "application/sql",
	".proto":      "text/x-protobuf",
	".dockerfile": "text/x-dockerfile",
	".Dockerfile": "text/x-dockerfile",
}

func detectMimeType(filePath string, content []byte) string {
	ext := ""
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			ext = filePath[i:]
			break
		}
		if filePath[i] == '/' {
			break
		}
	}

	if mimeType, ok := mimeTypeMap[ext]; ok {
		return mimeType
	}

	if len(content) > 0 && content[0] == 0 {
		return "application/octet-stream"
	}
	return "text/plain"
}
