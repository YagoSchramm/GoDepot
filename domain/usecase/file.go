package usecase

import (
	"context"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/google/uuid"
)

type FileUseCase interface {
	SetSyncRoot(ctx context.Context, userID uuid.UUID, path string) error
	ListFiles(ctx context.Context, userID uuid.UUID) ([]entity.File, error)
	GetFile(ctx context.Context, userID uuid.UUID, name string, opts entity.Options) (entity.Result, error)
}
