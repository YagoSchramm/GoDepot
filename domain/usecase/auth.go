package usecase

import (
	"context"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/google/uuid"
)

type AuthUseCase interface {
	AttemptRegister(ctx context.Context, user entity.User) (string, error)
	AttemptLogin(ctx context.Context, credentials entity.UserCredentials) (string, error)
	ValidateSession(ctx context.Context, userID uuid.UUID, email string) error
}
