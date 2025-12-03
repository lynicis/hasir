package authorization

import (
	"context"
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"go.uber.org/mock/gomock"
)

func TestIsUserOwner(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		setupMock     func(m *MockMemberRoleChecker)
		wantErr       bool
		wantErrCode   connect.Code
		wantErrSubstr string
	}{
		{
			name: "returns nil when user is owner",
			setupMock: func(m *MockMemberRoleChecker) {
				m.EXPECT().
					GetMemberRole(ctx, "org-1", "user-1").
					Return(MemberRoleOwner, nil)
			},
			wantErr: false,
		},
		{
			name: "returns permission denied when user is not a member",
			setupMock: func(m *MockMemberRoleChecker) {
				m.EXPECT().
					GetMemberRole(ctx, "org-1", "user-1").
					Return("", ErrMemberNotFound)
			},
			wantErr:       true,
			wantErrCode:   connect.CodePermissionDenied,
			wantErrSubstr: "not a member",
		},
		{
			name: "propagates unexpected error from checker",
			setupMock: func(m *MockMemberRoleChecker) {
				m.EXPECT().
					GetMemberRole(ctx, "org-1", "user-1").
					Return("", errors.New("db down"))
			},
			wantErr:       true,
			wantErrCode:   0, // non-connect error
			wantErrSubstr: "db down",
		},
		{
			name: "returns permission denied when user is not owner",
			setupMock: func(m *MockMemberRoleChecker) {
				m.EXPECT().
					GetMemberRole(ctx, "org-1", "user-1").
					Return(MemberRoleAuthor, nil)
			},
			wantErr:       true,
			wantErrCode:   connect.CodePermissionDenied,
			wantErrSubstr: "only organization owners",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockChecker := NewMockMemberRoleChecker(ctrl)
			tt.setupMock(mockChecker)

			err := IsUserOwner(ctx, mockChecker, "org-1", "user-1")

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if tt.wantErrCode != 0 {
				var connectErr *connect.Error
				if !errors.As(err, &connectErr) {
					t.Fatalf("expected connect.Error, got %T", err)
				}

				if connectErr.Code() != tt.wantErrCode {
					t.Fatalf("expected code %v, got %v", tt.wantErrCode, connectErr.Code())
				}

				if tt.wantErrSubstr != "" && !strings.Contains(connectErr.Message(), tt.wantErrSubstr) {
					t.Fatalf("expected error message to contain %q, got %q", tt.wantErrSubstr, connectErr.Message())
				}
			} else if tt.wantErrSubstr != "" && !strings.Contains(err.Error(), tt.wantErrSubstr) {
				t.Fatalf("expected error message to contain %q, got %q", tt.wantErrSubstr, err.Error())
			}
		})
	}
}
