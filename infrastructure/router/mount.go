package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Mount(root *mux.Router, protectedMiddlewares []mux.MiddlewareFunc, modules ...Module) {
	for _, module := range modules {
		subrouter := root.PathPrefix(module.Path()).Subrouter()

		for _, route := range module.Routes() {
			handler := http.Handler(route.Handler)

			if !route.Public {
				for i := len(protectedMiddlewares) - 1; i >= 0; i-- {
					handler = protectedMiddlewares[i](handler)
				}
			}

			subrouter.Handle(route.Path, handler).Methods(route.HttpMethods...)
		}
	}
}
