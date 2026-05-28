package entity

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Name       string
	Path       string
	MimeType   string
	Size       int64
	ModifiedAt time.Time
}
