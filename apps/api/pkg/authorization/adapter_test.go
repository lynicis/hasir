package authorization

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestNewOrgRepositoryAdapter(t *testing.T) {
	repo := NewOrgRepositoryAdapter(nil)

	assert.NotNil(t, repo)
}

func TestOrgRepositoryAdapter_GetMemberRole(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	t.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		wantRole := "owner"

		repo := NewMockOrganizationRepository(mockController)
		repo.EXPECT().GetMemberRoleString(ctx, "org-1", "user-1").Return(wantRole, nil).Times(1)

		adapter := NewOrgRepositoryAdapter(repo)

		role, err := adapter.GetMemberRole(ctx, "org-1", "user-1")

		assert.NoError(t, err)
		assert.Equal(t, wantRole, role)
	})

	t.Run("error", func(t *testing.T) {
		ctx := context.Background()
		wantErr := errors.New("boom")

		repo := NewMockOrganizationRepository(mockController)
		repo.EXPECT().GetMemberRoleString(ctx, "org-1", "user-1").Return("", wantErr).Times(1)

		adapter := NewOrgRepositoryAdapter(repo)

		_, err := adapter.GetMemberRole(ctx, "org-1", "user-1")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
	})
}
