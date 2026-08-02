package authorization

import "context"

type OrganizationRepository interface {
	GetMemberRoleString(ctx context.Context, organizationId, userId string) (string, error)
}

type OrganizationRepositoryAdapter struct {
	repo OrganizationRepository
}

func NewOrgRepositoryAdapter(repo OrganizationRepository) *OrganizationRepositoryAdapter {
	return &OrganizationRepositoryAdapter{
		repo: repo,
	}
}

func (a *OrganizationRepositoryAdapter) GetMemberRole(ctx context.Context, organizationId, userId string) (string, error) {
	return a.repo.GetMemberRoleString(ctx, organizationId, userId)
}
