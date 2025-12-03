package organization

import (
	"context"
	"time"
)

type Repository interface {
	CreateOrganization(ctx context.Context, org *OrganizationDTO) error
	GetOrganizations(ctx context.Context, page, pageSize int) (*[]OrganizationDTO, error)
	GetOrganizationsCount(ctx context.Context) (int, error)
	GetUserOrganizations(ctx context.Context, userId string, page, pageSize int) (*[]OrganizationDTO, error)
	GetUserOrganizationsCount(ctx context.Context, userId string) (int, error)
	GetOrganizationByName(ctx context.Context, name string) (*OrganizationDTO, error)
	GetOrganizationById(ctx context.Context, id string) (*OrganizationDTO, error)
	UpdateOrganization(ctx context.Context, org *OrganizationDTO) error
	DeleteOrganization(ctx context.Context, id string) error
	CreateInvites(ctx context.Context, invites []*OrganizationInviteDTO) error
	GetInviteByToken(ctx context.Context, token string) (*OrganizationInviteDTO, error)
	UpdateInviteStatus(ctx context.Context, id string, status InviteStatus, acceptedAt *time.Time) error
	AddMember(ctx context.Context, member *OrganizationMemberDTO) error
	GetMembers(ctx context.Context, organizationId string) ([]*OrganizationMemberDTO, []string, []string, error)
	GetMemberRole(ctx context.Context, organizationId, userId string) (MemberRole, error)
	GetMemberRoleString(ctx context.Context, organizationId, userId string) (string, error)
	UpdateMemberRole(ctx context.Context, organizationId, userId string, role MemberRole) error
	DeleteMember(ctx context.Context, organizationId, userId string) error
}
