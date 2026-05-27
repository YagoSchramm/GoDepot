package repository

import (
	"context"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/google/uuid"
)

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	// Store user register information
	AttemptRegister(ctx context.Context, user entity.User) (*uuid.UUID, error)
	// Store user login information
	AttemptLogin(ctx context.Context, credentials entity.UserCredentials) (*entity.User, error)
}
