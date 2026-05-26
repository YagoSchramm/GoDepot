package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Module interface {
	Name() string
	Path() string
	Routes() []RouteDefinition
	Middlewares() []mux.MiddlewareFunc
}

type RouteDefinition struct {
	Path        string
	Description string
	Handler     http.HandlerFunc
	HttpMethods []string
	Public      bool
	Middlewares []mux.MiddlewareFunc
}
