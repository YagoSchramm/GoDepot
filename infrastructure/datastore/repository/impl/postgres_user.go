package impl

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/repository"
	"github.com/google/uuid"
)

func NewAuthRepository(db *sql.DB) repository.AuthRepository {
	return &authRepository{
		db: db,
	}
}

type authRepository struct {
	db *sql.DB
}

//go:embed _query/auth/get_user_by_email.sql
var getUserByEmailQuery string

//go:embed _query/auth/attempt_register.sql
var attemptRegisterQuery string

//go:embed _query/auth/attempt_login.sql
var attemptLoginQuery string

func (r *authRepository) AttemptRegister(ctx context.Context, user entity.User) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		attemptRegisterQuery,
		user.Email,
		user.Password,
	).Scan(&id)
	if err != nil {
		return nil, derr.JoinError("failed to execute the query", err)
	}

	return &id, nil
}

func (r *authRepository) AttemptLogin(ctx context.Context, credentials entity.UserCredentials) (*entity.User, error) {
	var user entity.User

	row := r.db.QueryRowContext(ctx, attemptLoginQuery, credentials.Email)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		return nil, derr.JoinError("failed to execute the query", err)
	}

	return &user, err
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	row := r.db.QueryRowContext(ctx, getUserByEmailQuery, email)

	var user entity.User
	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, derr.NotFoundError
		}
		return nil, derr.JoinError("failed to get user by email", err)
	}

	return &user, nil
}
