package service

import (
	"errors"
	"log"
	"os"
	"strings"

	usecaseimpl "github.com/YagoSchramm/GoDepot/domain/usecase/impl"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/index/impl"
	repoimpl "github.com/YagoSchramm/GoDepot/infrastructure/datastore/repository/impl"
	"github.com/YagoSchramm/GoDepot/infrastructure/foundation/db"
	"github.com/YagoSchramm/GoDepot/infrastructure/foundation/watcher"
	approuter "github.com/YagoSchramm/GoDepot/infrastructure/router"
	modules "github.com/YagoSchramm/GoDepot/infrastructure/router/module"
	"github.com/gorilla/mux"
)

func Build() (*mux.Router, func(), error) {
	content, err := os.ReadFile(".env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, func() {}, err
	}

	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			key, value, found := strings.Cut(line, "=")
			if !found {
				continue
			}

			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			value = strings.Trim(value, `"'`)

			if key == "" {
				continue
			}

			if os.Getenv(key) == "" {
				_ = os.Setenv(key, value)
			}
		}
	}

	dsn := os.Getenv("DATABASE_URL")
	secret := os.Getenv("JWT_SECRET")

	if dsn == "" {
		return nil, func() {}, errors.New("DATABASE_URL is not set")
	}

	dbConn, err := db.NewPostgresConnection(dsn)
	if err != nil {
		return nil, func() {}, err
	}

	idx := impl.NewFileIndex()

	w, err := watcher.NewWatcher("./files", idx)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Stop()
	if err := w.Start(); err != nil {
		log.Fatal(err)
	}

	authRepository := repoimpl.NewAuthRepository(dbConn)
	cleanup := func() {
		_ = dbConn.Close()
	}

	authUseCase := usecaseimpl.NewAuthRepository(authRepository, secret)

	authModule := modules.NewAuthModule(authUseCase, secret)

	router := mux.NewRouter()
	approuter.Mount(
		router,
		authModule.Middlewares(),
		authModule,
	)

	return router, cleanup, nil
}
