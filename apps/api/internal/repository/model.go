package repository

import "time"

type RepositoryDTO struct {
	Id        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	OwnerId   string    `json:"owner_id" db:"owner_id"`
	Path      string    `json:"path" db:"path"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
