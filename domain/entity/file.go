package entity

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	MimeType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}
