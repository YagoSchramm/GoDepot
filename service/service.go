package service

import (
	"errors"
	"os"
	"strings"
	"time"

	usecaseimpl "github.com/YagoSchramm/GoDepot/domain/usecase/impl"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/cache"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/db"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/index/impl"
	repoimpl "github.com/YagoSchramm/GoDepot/infrastructure/datastore/repository/impl"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/processor"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/processor/impl/document"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/processor/impl/image"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/processor/impl/raw"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/processor/impl/video"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/watcher"
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
	fileCache := cache.NewMemoryCache(5 * time.Minute)

	w, err := watcher.NewWatcher(idx, fileCache)
	if err != nil {
		return nil, func() {}, err
	}

	registry := processor.NewRegistry()

	registry.Register(image.NewImageProcessor())
	registry.Register(video.NewVideoProcessor())
	registry.Register(document.NewDocumentProcessor())
	registry.Register(raw.NewRawProcessor())

	w.Start()

	authRepository := repoimpl.NewAuthRepository(dbConn)
	cleanup := func() {
		w.Stop()
		_ = dbConn.Close()
	}

	authUseCase := usecaseimpl.NewAuthRepository(authRepository, secret)
	fileUseCase := usecaseimpl.NewFileUseCase(idx, w, registry, fileCache)

	authModule := modules.NewAuthModule(authUseCase, secret)
	fileModule := modules.NewFileModule(fileUseCase)

	router := mux.NewRouter()
	approuter.Mount(
		router,
		authModule.Middlewares(),
		authModule,
		fileModule,
	)

	return router, cleanup, nil
}
