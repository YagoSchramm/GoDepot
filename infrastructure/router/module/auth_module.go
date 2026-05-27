package module

import (
	"github.com/YagoSchramm/GoDepot/domain/usecase"
	"github.com/YagoSchramm/GoDepot/infrastructure/router"
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

func (a *authModule) Middlewares() []mux.MiddlewareFunc {
	panic("unimplemented")
}

func (a *authModule) Name() string {
	panic("unimplemented")
}

func (a *authModule) Path() string {
	panic("unimplemented")
}

func (a *authModule) Routes() []router.RouteDefinition {
	panic("unimplemented")
}
