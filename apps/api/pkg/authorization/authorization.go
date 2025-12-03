package authorization

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

const (
	MemberRoleReader = "reader"
	MemberRoleAuthor = "author"
	MemberRoleOwner  = "owner"
)

var (
	ErrMemberNotFound = connect.NewError(connect.CodeNotFound, errors.New("member not found"))
)

type MemberRoleChecker interface {
	GetMemberRole(ctx context.Context, organizationId, userId string) (string, error)
}

func IsUserOwner(ctx context.Context, checker MemberRoleChecker, organizationId, userId string) error {
	role, err := checker.GetMemberRole(ctx, organizationId, userId)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return connect.NewError(connect.CodePermissionDenied, errors.New("you are not a member of this organization"))
		}
		return err
	}

	if role != MemberRoleOwner {
		return connect.NewError(connect.CodePermissionDenied, errors.New("only organization owners can perform this operation"))
	}

	return nil
}
