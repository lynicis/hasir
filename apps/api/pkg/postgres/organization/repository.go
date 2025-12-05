package organization

import (
	"context"
	"errors"
	"fmt"
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

	"hasir-api/internal/organization"
	"hasir-api/pkg/config"
)

var (
	ErrOrganizationAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("organization already exists"))
	ErrInviteNotFound            = connect.NewError(connect.CodeNotFound, errors.New("invite not found"))
	ErrOrganizationNotFound      = connect.NewError(connect.CodeNotFound, errors.New("organization not found"))
	ErrMemberAlreadyExists       = connect.NewError(connect.CodeAlreadyExists, errors.New("member already exists"))
	ErrMemberNotFound            = connect.NewError(connect.CodeNotFound, errors.New("member not found"))
	ErrFailedAcquireConnection   = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	ErrUniqueViolationCode       = "23505"
)

type OrganizationRepository struct {
	connectionPool *pgxpool.Pool
	tracer         trace.Tracer
}

func NewOrganizationRepository(
	cfg *config.Config,
	traceProvider *sdktrace.TracerProvider,
) *OrganizationRepository {
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
		tracer = traceProvider.Tracer("OrganizationPostgreSQLRepository")
	} else {
		tracer = noop.NewTracerProvider().Tracer("OrganizationPostgreSQLRepository")
	}

	return &OrganizationRepository{
		connectionPool: pgConnectionPool,
		tracer:         tracer,
	}
}

func (r *OrganizationRepository) GetConnectionPool() *pgxpool.Pool {
	return r.connectionPool
}

func (r *OrganizationRepository) GetTracer() trace.Tracer {
	return r.tracer
}

func querySingleRow[T any](ctx context.Context, conn *pgxpool.Conn, span trace.Span, sql string, args []any, notFoundErr error) (*T, error) {
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to execute query"))
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow[T](rows, pgx.RowToStructByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notFoundErr
		}
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect row"))
	}

	return &result, nil
}

func (r *OrganizationRepository) CreateOrganization(ctx context.Context, org *organization.OrganizationDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateOrganization", trace.WithAttributes(attribute.KeyValue{
		Key:   "newOrganization",
		Value: attribute.StringValue(fmt.Sprintf("%+v", org)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `INSERT INTO organizations (id, name, visibility, created_by, created_at)
			VALUES (@Id, @Name, @Visibility, @CreatedBy, @CreatedAt)`
	sqlArgs := pgx.NamedArgs{
		"Id":         org.Id,
		"Name":       org.Name,
		"Visibility": org.Visibility,
		"CreatedBy":  org.CreatedBy,
		"CreatedAt":  time.Now().UTC(),
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrOrganizationAlreadyExists
			}
			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to execute insert organization query"))
	}

	return nil
}

func (r *OrganizationRepository) GetOrganizations(ctx context.Context, page, pageSize int) (*[]organization.OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizations", trace.WithAttributes(
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
	sql := "SELECT * FROM organizations WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	var rows pgx.Rows
	rows, err = connection.Query(ctx, sql, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query organizations"))
	}
	defer rows.Close()

	orgs, err := pgx.CollectRows[organization.OrganizationDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect rows"))
	}

	return &orgs, nil
}

func (r *OrganizationRepository) GetUserOrganizations(ctx context.Context, userId string, page, pageSize int) (*[]organization.OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetUserOrganizations", trace.WithAttributes(
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
		SELECT o.*
		FROM organizations o
		INNER JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND o.deleted_at IS NULL
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := connection.Query(ctx, sql, userId, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to query user organizations"))
	}
	defer rows.Close()

	orgs, err := pgx.CollectRows[organization.OrganizationDTO](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect user organization rows"))
	}

	return &orgs, nil
}

func (r *OrganizationRepository) GetOrganizationsCount(ctx context.Context) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizationsCount")
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return 0, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT COUNT(*) FROM organizations WHERE deleted_at IS NULL"

	var count int
	err = connection.QueryRow(ctx, sql).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count organizations"))
	}

	return count, nil
}

func (r *OrganizationRepository) GetUserOrganizationsCount(ctx context.Context, userId string) (int, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetUserOrganizationsCount", trace.WithAttributes(
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
		FROM organizations o
		INNER JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND o.deleted_at IS NULL`

	var count int
	err = connection.QueryRow(ctx, sql, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, connect.NewError(connect.CodeInternal, errors.New("failed to count user organizations"))
	}

	return count, nil
}

func (r *OrganizationRepository) GetOrganizationByName(ctx context.Context, name string) (*organization.OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizationByName", trace.WithAttributes(
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

	sql := "SELECT * FROM organizations WHERE name = $1 AND deleted_at IS NULL"
	return querySingleRow[organization.OrganizationDTO](ctx, connection, span, sql, []any{name}, ErrOrganizationNotFound)
}

func (r *OrganizationRepository) GetOrganizationById(ctx context.Context, id string) (*organization.OrganizationDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetOrganizationById", trace.WithAttributes(
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

	sql := "SELECT * FROM organizations WHERE id = $1 AND deleted_at IS NULL"
	return querySingleRow[organization.OrganizationDTO](ctx, connection, span, sql, []any{id}, ErrOrganizationNotFound)
}

func (r *OrganizationRepository) UpdateOrganization(ctx context.Context, org *organization.OrganizationDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateOrganization", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(org.Id),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `UPDATE organizations
			SET name = @Name,
				visibility = @Visibility
			WHERE id = @Id AND deleted_at IS NULL`
	sqlArgs := pgx.NamedArgs{
		"Id":         org.Id,
		"Name":       org.Name,
		"Visibility": org.Visibility,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrOrganizationAlreadyExists
			}
			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to update organization"))
	}

	if result.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

func (r *OrganizationRepository) DeleteOrganization(ctx context.Context, id string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteOrganization", trace.WithAttributes(
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
	sql := `UPDATE organizations SET deleted_at = @DeletedAt WHERE id = @Id AND deleted_at IS NULL`
	sqlArgs := pgx.NamedArgs{
		"Id":        id,
		"DeletedAt": now,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to delete organization"))
	}

	if result.RowsAffected() == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

func (r *OrganizationRepository) CreateInvites(ctx context.Context, invites []*organization.OrganizationInviteDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "CreateInvites", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "inviteCount",
			Value: attribute.IntValue(len(invites)),
		},
	))
	defer span.End()

	if len(invites) == 0 {
		return nil
	}

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	orgID := invites[0].OrganizationId

	emails := make([]string, 0, len(invites))
	for _, invite := range invites {
		emails = append(emails, invite.Email)
	}

	existingEmails := make(map[string]struct{}, len(emails))
	rows, err := connection.Query(
		ctx,
		`SELECT email FROM organization_invites WHERE organization_id = $1 AND status = 'pending' AND email = ANY($2)`,
		orgID,
		emails,
	)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to check existing pending invites"))
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			span.RecordError(err)
			return connect.NewError(connect.CodeInternal, errors.New("failed to scan existing pending invites"))
		}
		existingEmails[email] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to iterate existing pending invites"))
	}

	filteredInvites := make([]*organization.OrganizationInviteDTO, 0, len(invites))
	for _, invite := range invites {
		if _, found := existingEmails[invite.Email]; found {
			continue
		}

		filteredInvites = append(filteredInvites, invite)
	}

	invites = filteredInvites

	if len(invites) == 0 {
		return nil
	}

	tx, err := connection.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to begin transaction"))
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	inviteSQL := `INSERT INTO organization_invites (id, organization_id, email, token, invited_by, role, status, created_at, expires_at)
			VALUES (@Id, @OrganizationId, @Email, @Token, @InvitedBy, @Role, @Status, @CreatedAt, @ExpiresAt)`

	inviteBatch := &pgx.Batch{}
	for _, invite := range invites {
		sqlArgs := pgx.NamedArgs{
			"Id":             invite.Id,
			"OrganizationId": invite.OrganizationId,
			"Email":          invite.Email,
			"Token":          invite.Token,
			"InvitedBy":      invite.InvitedBy,
			"Role":           invite.Role,
			"Status":         invite.Status,
			"CreatedAt":      invite.CreatedAt,
			"ExpiresAt":      invite.ExpiresAt,
		}
		inviteBatch.Queue(inviteSQL, sqlArgs)
	}

	inviteResults := tx.SendBatch(ctx, inviteBatch)

	for i := 0; i < len(invites); i++ {
		_, err := inviteResults.Exec()
		if err != nil {
			span.RecordError(err)
			_ = inviteResults.Close()
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == ErrUniqueViolationCode {
					return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("invite already exists for email: %s", invites[i].Email))
				}
			}
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create invite %d: %w", i, err))
		}
	}

	if err := inviteResults.Close(); err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to close invite batch results: %w", err))
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to commit transaction"))
	}
	committed = true

	return nil
}

func (r *OrganizationRepository) GetInviteByToken(ctx context.Context, token string) (*organization.OrganizationInviteDTO, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetInviteByToken", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "token",
			Value: attribute.StringValue(token),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := "SELECT * FROM organization_invites WHERE token = $1"
	return querySingleRow[organization.OrganizationInviteDTO](ctx, connection, span, sql, []any{token}, ErrInviteNotFound)
}

func (r *OrganizationRepository) UpdateInviteStatus(ctx context.Context, id string, status organization.InviteStatus, acceptedAt *time.Time) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateInviteStatus", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "status",
			Value: attribute.StringValue(string(status)),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `UPDATE organization_invites SET status = @Status, accepted_at = @AcceptedAt WHERE id = @Id`
	sqlArgs := pgx.NamedArgs{
		"Id":         id,
		"Status":     status,
		"AcceptedAt": acceptedAt,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to update invite status"))
	}

	if result.RowsAffected() == 0 {
		return ErrInviteNotFound
	}

	return nil
}

func (r *OrganizationRepository) AddMember(ctx context.Context, member *organization.OrganizationMemberDTO) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "AddMember", trace.WithAttributes(attribute.KeyValue{
		Key:   "member",
		Value: attribute.StringValue(fmt.Sprintf("%+v", member)),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `INSERT INTO organization_members (id, organization_id, user_id, role, joined_at)
			VALUES (@Id, @OrganizationId, @UserId, @Role, @JoinedAt)`
	sqlArgs := pgx.NamedArgs{
		"Id":             member.Id,
		"OrganizationId": member.OrganizationId,
		"UserId":         member.UserId,
		"Role":           member.Role,
		"JoinedAt":       member.JoinedAt,
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == ErrUniqueViolationCode {
				return ErrMemberAlreadyExists
			}
			return err
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to add member"))
	}

	return nil
}

type memberRow struct {
	organization.OrganizationMemberDTO
	Username string `db:"username"`
	Email    string `db:"email"`
}

func (r *OrganizationRepository) GetMembers(ctx context.Context, organizationId string) ([]*organization.OrganizationMemberDTO, []string, []string, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetMembers", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, nil, nil, ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `SELECT id, organization_id, user_id, role, joined_at, username, email
			FROM organization_members_view
			WHERE organization_id = $1
			ORDER BY joined_at ASC`

	rows, err := connection.Query(ctx, sql, organizationId)
	if err != nil {
		span.RecordError(err)
		return nil, nil, nil, connect.NewError(connect.CodeInternal, errors.New("failed to query members"))
	}
	defer rows.Close()

	memberRows, err := pgx.CollectRows[memberRow](rows, pgx.RowToStructByName)
	if err != nil {
		span.RecordError(err)
		return nil, nil, nil, connect.NewError(connect.CodeInternal, errors.New("failed to collect member rows"))
	}

	members := make([]*organization.OrganizationMemberDTO, len(memberRows))
	usernames := make([]string, len(memberRows))
	emails := make([]string, len(memberRows))

	for i, row := range memberRows {
		members[i] = &row.OrganizationMemberDTO
		usernames[i] = row.Username
		emails[i] = row.Email
	}

	return members, usernames, emails, nil
}

func (r *OrganizationRepository) GetMemberRole(ctx context.Context, organizationId, userId string) (organization.MemberRole, error) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "GetMemberRole", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return "", ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2`

	var role organization.MemberRole
	err = connection.QueryRow(ctx, sql, organizationId, userId).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrMemberNotFound
		}
		span.RecordError(err)
		return "", connect.NewError(connect.CodeInternal, errors.New("failed to query member role"))
	}

	return role, nil
}

func (r *OrganizationRepository) GetMemberRoleString(ctx context.Context, organizationId, userId string) (string, error) {
	role, err := r.GetMemberRole(ctx, organizationId, userId)
	return string(role), err
}

func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, organizationId, userId string, role organization.MemberRole) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "UpdateMemberRole", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
		attribute.KeyValue{
			Key:   "role",
			Value: attribute.StringValue(string(role)),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `UPDATE organization_members SET role = @Role WHERE organization_id = @OrganizationId AND user_id = @UserId`
	sqlArgs := pgx.NamedArgs{
		"OrganizationId": organizationId,
		"UserId":         userId,
		"Role":           role,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to update member role"))
	}

	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (r *OrganizationRepository) DeleteMember(ctx context.Context, organizationId, userId string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "DeleteMember", trace.WithAttributes(
		attribute.KeyValue{
			Key:   "organizationId",
			Value: attribute.StringValue(organizationId),
		},
		attribute.KeyValue{
			Key:   "userId",
			Value: attribute.StringValue(userId),
		},
	))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `DELETE FROM organization_members WHERE organization_id = @OrganizationId AND user_id = @UserId`
	sqlArgs := pgx.NamedArgs{
		"OrganizationId": organizationId,
		"UserId":         userId,
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(connect.CodeInternal, errors.New("failed to delete member"))
	}

	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}
