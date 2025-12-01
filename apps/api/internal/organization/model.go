package organization

import (
	"hasir-api/pkg/proto"
	"time"
)

type OrganizationDTO struct {
	Id         string           `json:"id" db:"id"`
	Name       string           `json:"name" db:"name"`
	Visibility proto.Visibility `json:"visibility" db:"visibility"`
	CreatedBy  string           `json:"created_by" db:"created_by"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
	DeletedAt  *time.Time       `json:"deleted_at,omitempty" db:"deleted_at"`
}

type InviteStatus string

const (
	InviteStatusPending   InviteStatus = "pending"
	InviteStatusAccepted  InviteStatus = "accepted"
	InviteStatusExpired   InviteStatus = "expired"
	InviteStatusCancelled InviteStatus = "cancelled"
)

type OrganizationInviteDTO struct {
	Id             string       `json:"id" db:"id"`
	OrganizationId string       `json:"organization_id" db:"organization_id"`
	Email          string       `json:"email" db:"email"`
	Token          string       `json:"token" db:"token"`
	InvitedBy      string       `json:"invited_by" db:"invited_by"`
	Status         InviteStatus `json:"status" db:"status"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	ExpiresAt      time.Time    `json:"expires_at" db:"expires_at"`
	AcceptedAt     *time.Time   `json:"accepted_at,omitempty" db:"accepted_at"`
}

type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
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

type EmailJobDTO struct {
	Id               string         `json:"id" db:"id"`
	InviteId         string         `json:"invite_id" db:"invite_id"`
	OrganizationId   string         `json:"organization_id" db:"organization_id"`
	Email            string         `json:"email" db:"email"`
	OrganizationName string         `json:"organization_name" db:"organization_name"`
	InviteToken      string         `json:"invite_token" db:"invite_token"`
	Status           EmailJobStatus `json:"status" db:"status"`
	Attempts         int            `json:"attempts" db:"attempts"`
	MaxAttempts      int            `json:"max_attempts" db:"max_attempts"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	ProcessedAt      *time.Time     `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
	ErrorMessage     *string        `json:"error_message,omitempty" db:"error_message"`
}
