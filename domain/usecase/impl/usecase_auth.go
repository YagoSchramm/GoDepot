package impl

import (
	"context"
	"errors"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/domain/rules"
	"github.com/YagoSchramm/GoDepot/domain/usecase"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/repository"
	"github.com/YagoSchramm/GoDepot/infrastructure/security/hash"
	"github.com/YagoSchramm/GoDepot/infrastructure/security/jwt"
	"github.com/google/uuid"
)

func NewAuthRepository(repository repository.AuthRepository, secret string) usecase.AuthUseCase {
	return authUseCase{
		repository: repository,
		secret:     secret,
	}
}

type authUseCase struct {
	repository repository.AuthRepository
	secret     string
}

func (u authUseCase) AttemptLogin(ctx context.Context, credentials entity.UserCredentials) (string, error) {
	err := rules.ValidateLogin(credentials)
	if err != nil {
		return "", nil
	}
	existedUser, err := u.repository.GetUserByEmail(ctx, credentials.Email)
	if err != nil && !errors.Is(err, derr.NotFoundError) {
		return "", derr.JoinError("failed to get user by email", err)
	}

	if existedUser == nil {
		return "", derr.NotFoundError
	}

	user, err := u.repository.AttemptLogin(ctx, credentials)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", derr.NewNotFoundError("user not found")
	}

	valid := hash.CheckHash([]byte(credentials.Password), []byte(user.Password))
	if !valid {
		return "", derr.InvalidCredentials
	}

	token, err := jwt.GenerateToken(existedUser.ID, credentials.Email, []byte(u.secret))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u authUseCase) AttemptRegister(ctx context.Context, user entity.User) (string, error) {
	err := rules.ValidateRegister(user)
	if err != nil {
		return "", err
	}

	existedUser, err := u.repository.GetUserByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, derr.NotFoundError) {
		return "", derr.JoinError("failed to get user by email", err)
	}

	if existedUser != nil {
		return "", derr.UserAlreadyExists
	}

	hashedPassword, err := hash.Hash(user.Password)
	if err != nil {
		return "", derr.JoinError("failed to hash the password", err)
	}

	user.Password = hashedPassword

	id, err := u.repository.AttemptRegister(ctx, user)
	if err != nil {
		return "", derr.JoinError("failed to attempt register the user", err)
	}

	token, err := jwt.GenerateToken(*id, user.Email, []byte(u.secret))
	if err != nil {
		return "", derr.JoinError("failed to generate the token", err)

	}

	return token, err
}

func (u authUseCase) ValidateSession(ctx context.Context, userID uuid.UUID, email string) error {
	if userID == uuid.Nil || email == "" {
		return derr.UnauthorizedError
	}

	user, err := u.repository.GetUserByEmail(ctx, email)
	if err != nil {
		var clientErr derr.ClientError
		if errors.As(err, &clientErr) && clientErr.Code == derr.NotFoundError.Code {
			return derr.UnauthorizedError
		}
		return err
	}
	if user.ID != userID {
		return derr.UnauthorizedError
	}

	return nil
}
