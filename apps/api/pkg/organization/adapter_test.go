package organization

import (
	"context"
	"errors"
	"testing"
)

type fakeOrgRepo struct {
	role string
	err  error
}

func (f *fakeOrgRepo) GetMemberRoleString(_ context.Context, _ string, _ string) (string, error) {
	return f.role, f.err
}

func TestOrgRepositoryAdapter_GetMemberRole_ForwardsCall(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	wantRole := "owner"
	repo := &fakeOrgRepo{role: wantRole}

	adapter := NewOrgRepositoryAdapter(repo)

	role, err := adapter.GetMemberRole(ctx, "org-1", "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if role != wantRole {
		t.Fatalf("expected role %q, got %q", wantRole, role)
	}
}

func TestOrgRepositoryAdapter_GetMemberRole_PropagatesError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	wantErr := errors.New("boom")
	repo := &fakeOrgRepo{err: wantErr}

	adapter := NewOrgRepositoryAdapter(repo)

	_, err := adapter.GetMemberRole(ctx, "org-1", "user-1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
