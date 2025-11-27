package organization

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

type OrganizationDTO struct {
	Id         string     `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	Visibility Visibility `json:"visibility" db:"visibility"`
	CreatedBy  string     `json:"created_by" db:"created_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
