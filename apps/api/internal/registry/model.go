package registry

import (
	"time"

	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"
)

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

var protoVisibilityMap = map[shared.Visibility]Visibility{
	shared.Visibility_VISIBILITY_PUBLIC:  VisibilityPublic,
	shared.Visibility_VISIBILITY_PRIVATE: VisibilityPrivate,
}

type RepositoryDTO struct {
	Id             string     `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	CreatedBy      string     `json:"created_by" db:"created_by"`
	OrganizationId string     `json:"organization_id" db:"organization_id"`
	Path           string     `json:"path" db:"path"`
	Visibility     Visibility `json:"visibility" db:"visibility"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
