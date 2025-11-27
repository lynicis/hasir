package organization

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	organizationv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/organization/v1"
)

type Service interface {
	CreateOrganization(ctx context.Context, req *organizationv1.CreateOrganizationRequest, createdBy string) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) CreateOrganization(ctx context.Context, req *organizationv1.CreateOrganizationRequest, createdBy string) error {
	existingOrg, err := s.repository.GetOrganizationByName(ctx, req.GetName())
	if err != nil && !errors.Is(err, ErrOrganizationNotFound) {
		return err
	}

	if existingOrg != nil {
		return connect.NewError(connect.CodeAlreadyExists, errors.New("organization already exists"))
	}

	visibility := protoVisibilityMap[req.GetVisibility()]
	org := &OrganizationDTO{
		Id:         uuid.NewString(),
		Name:       req.GetName(),
		Visibility: visibility,
		CreatedBy:  createdBy,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repository.CreateOrganization(ctx, org); err != nil {
		return err
	}

	return nil
}
