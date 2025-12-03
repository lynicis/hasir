package registry

import (
	"errors"

	"connectrpc.com/connect"
)

var (
	ErrRepositoryAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("repository already exists"))
	ErrRepositoryNotFound      = connect.NewError(connect.CodeNotFound, errors.New("repository not found"))
)
