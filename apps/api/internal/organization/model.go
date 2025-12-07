package organization

import (
	"time"

	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"

	"hasir-api/pkg/proto"
)

type OrganizationDTO struct {
	Id         string           `db:"id"`
	Name       string           `db:"name"`
	Visibility proto.Visibility `db:"visibility"`
	CreatedBy  string           `db:"created_by"`
	CreatedAt  time.Time        `db:"created_at"`
	DeletedAt  *time.Time       `db:"deleted_at"`
}

type InviteStatus string

const (
	InviteStatusPending   InviteStatus = "pending"
	InviteStatusAccepted  InviteStatus = "accepted"
	InviteStatusExpired   InviteStatus = "expired"
	InviteStatusCancelled InviteStatus = "cancelled"
)

type OrganizationInviteDTO struct {
	Id             string       `db:"id"`
	OrganizationId string       `db:"organization_id"`
	Email          string       `db:"email"`
	Token          string       `db:"token"`
	InvitedBy      string       `db:"invited_by"`
	Role           MemberRole   `db:"role"`
	Status         InviteStatus `db:"status"`
	CreatedAt      time.Time    `db:"created_at"`
	ExpiresAt      time.Time    `db:"expires_at"`
	AcceptedAt     *time.Time   `db:"accepted_at"`
}

type MemberRole string

const (
	MemberRoleReader MemberRole = "reader"
	MemberRoleAuthor MemberRole = "author"
	MemberRoleOwner  MemberRole = "owner"
)

type OrganizationMemberDTO struct {
	Id             string     `json:"id" db:"id"`
	OrganizationId string     `json:"organization_id" db:"organization_id"`
	UserId         string     `json:"user_id" db:"user_id"`
	Role           MemberRole `json:"role" db:"role"`
	JoinedAt       time.Time  `json:"joined_at" db:"joined_at"`
}

type EmailJobStatus string

const (
	EmailJobStatusPending    EmailJobStatus = "pending"
	EmailJobStatusProcessing EmailJobStatus = "processing"
	EmailJobStatusCompleted  EmailJobStatus = "completed"
	EmailJobStatusFailed     EmailJobStatus = "failed"
)

var SharedRoleToMemberRoleMap = map[shared.Role]MemberRole{
	shared.Role_ROLE_OWNER:  MemberRoleOwner,
	shared.Role_ROLE_AUTHOR: MemberRoleAuthor,
	shared.Role_ROLE_READER: MemberRoleReader,
}

var MemberRoleToSharedRoleMap = map[MemberRole]shared.Role{
	MemberRoleOwner:  shared.Role_ROLE_OWNER,
	MemberRoleAuthor: shared.Role_ROLE_AUTHOR,
	MemberRoleReader: shared.Role_ROLE_READER,
}

type EmailJobDTO struct {
	Id               string         `db:"id"`
	InviteId         string         `db:"invite_id"`
	OrganizationId   string         `db:"organization_id"`
	Email            string         `db:"email"`
	OrganizationName string         `db:"organization_name"`
	InviteToken      string         `db:"invite_token"`
	Status           EmailJobStatus `db:"status"`
	Attempts         int            `db:"attempts"`
	MaxAttempts      int            `db:"max_attempts"`
	CreatedAt        time.Time      `db:"created_at"`
	ProcessedAt      *time.Time     `db:"processed_at"`
	CompletedAt      *time.Time     `db:"completed_at"`
	ErrorMessage     *string        `db:"error_message"`
}
