package module

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/domain/usecase"
	"github.com/YagoSchramm/GoDepot/infrastructure/router"
	"github.com/YagoSchramm/GoDepot/infrastructure/security/jwt"
	"github.com/gorilla/mux"
)

func NewAuthModule(authUseCase usecase.AuthUseCase, secret string) router.Module {
	return &authModule{
		authUseCase: authUseCase,
		name:        "Auth",
		path:        "/auth",
		secret:      secret,
	}
}

type authModule struct {
	authUseCase usecase.AuthUseCase
	name        string
	path        string
	secret      string
}

func (m authModule) Middlewares() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{m.sessionMiddleware()}
}

func (m authModule) Name() string {
	return m.name
}

func (m authModule) Path() string {
	return m.path
}

func (m authModule) Routes() []router.RouteDefinition {
	return []router.RouteDefinition{
		{
			Path:        "/login",
			Description: "Attempt to login",
			Handler:     m.login,
			HttpMethods: []string{http.MethodPost},
			Public:      true,
		},
		{
			Path:        "/register",
			Description: "Attempt to register",
			Handler:     m.register,
			HttpMethods: []string{http.MethodPost},
			Public:      true,
		},
	}
}

func (m authModule) sessionMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				authHeader := r.Header.Get("Authorization")
				scheme, token, ok := strings.Cut(authHeader, " ")
				if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
					w.Header().Set("WWW-Authenticate", "Bearer")
					router.HandleError(w, derr.UnauthorizedError)
					return
				}

				claims, err := jwt.ValidateToken(token, []byte(m.secret))
				if err != nil {
					router.HandleError(w, derr.UnauthorizedError)
					return
				}

				err = m.authUseCase.ValidateSession(ctx, claims.UserID, claims.Email)
				if err != nil {
					router.HandleError(w, err)
					return
				}
				ctx = context.WithValue(ctx, "user_claims", claims)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
	}
}

func (m authModule) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read request body", "error", err) // needs logger
		router.HandleError(w, err)
		return
	}

	var credentials entity.UserCredentials
	err = json.Unmarshal(body, &credentials)
	if err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal request body", "error", err) // needs logger
		router.HandleError(w, err)
		return
	}

	token, err := m.authUseCase.AttemptLogin(ctx, credentials)
	if err != nil {
		slog.ErrorContext(ctx, "failed to attempt login", "error", err) // needs logger
		router.HandleError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	err = router.Write(w, token)
	if err != nil {
		slog.ErrorContext(ctx, "failed to write response", "error", err) // needs logger
	}
}

func (m authModule) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read request body", "error", err) // needs logger
		router.HandleError(w, err)
		return
	}

	var user entity.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal request body", "error", err) // needs logger
		return
	}

	token, err := m.authUseCase.AttemptRegister(ctx, user)
	if err != nil {
		slog.ErrorContext(ctx, "failed to attempt register", "error", err) // needs logger
		router.HandleError(w, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	err = router.Write(w, token)
	if err != nil {
		slog.ErrorContext(ctx, "failed to write response", "error", err) // needs logger
	}
}
