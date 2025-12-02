package organization

import "context"

type OrgRepositoryAdapter struct {
	repo interface {
		GetMemberRoleString(ctx context.Context, organizationId, userId string) (string, error)
	}
}

func NewOrgRepositoryAdapter(repo interface {
	GetMemberRoleString(ctx context.Context, organizationId, userId string) (string, error)
}) *OrgRepositoryAdapter {
	return &OrgRepositoryAdapter{repo: repo}
}

func (a *OrgRepositoryAdapter) GetMemberRole(ctx context.Context, organizationId, userId string) (string, error) {
	return a.repo.GetMemberRoleString(ctx, organizationId, userId)
}
