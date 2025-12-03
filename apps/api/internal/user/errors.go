package user

import (
	"errors"

	"connectrpc.com/connect"
)

var (
	ErrNoRows               = connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	ErrRefreshTokenNotFound = connect.NewError(connect.CodeNotFound, errors.New("refresh token not found"))
	ErrInternalServer       = connect.NewError(connect.CodeInternal, errors.New("something went wrong"))
)
