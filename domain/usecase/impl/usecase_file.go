package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/domain/usecase"
	"github.com/YagoSchramm/GoDepot/domain/usecase/dto"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/cache"
	"github.com/YagoSchramm/GoDepot/infrastructure/datastore/index"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/processor"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/watcher"
	"github.com/google/uuid"
)

func NewFileUseCase(idx index.FileIndex, syncWatcher *watcher.Watcher, registry *processor.Registry, fileCache cache.Cache) usecase.FileUseCase {
	return fileUseCase{
		index:   idx,
		watcher: syncWatcher,
		registry: registry,
		cache:   fileCache,
	}
}

type fileUseCase struct {
	index    index.FileIndex
	watcher  *watcher.Watcher
	registry *processor.Registry
	cache    cache.Cache
}

func (u fileUseCase) SetSyncRoot(ctx context.Context, userID uuid.UUID, path string) error {
	return u.watcher.SetRoot(userID, path)
}

func (u fileUseCase) ListFiles(ctx context.Context, userID uuid.UUID) ([]entity.File, error) {
	cacheKey := fileListCacheKey(userID)
	if data, ok := u.cache.Get(cacheKey); ok {
		var files []entity.File
		if err := json.Unmarshal(data, &files); err == nil {
			return files, nil
		}
		u.cache.Invalidate(cacheKey)
	}

	files := u.index.ListByUserID(userID.String())
	data, err := json.Marshal(files)
	if err != nil {
		return nil, err
	}
	u.cache.Set(cacheKey, data)
	return files, nil
}

func (u fileUseCase) GetFile(ctx context.Context, userID uuid.UUID, name string, opts entity.Options) (entity.Result, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return entity.Result{}, derr.NewBadRequestError("file name is required")
	}

	file, err := u.index.Get(userID.String(), name)
	if err != nil {
		return entity.Result{}, err
	}

	cacheKey := fileContentCacheKey(userID, file, opts)
	if data, ok := u.cache.Get(cacheKey); ok {
		var cached dto.CachedFileResult
		if err := json.Unmarshal(data, &cached); err == nil {
			return entity.Result{
				Data:        cached.Data,
				ContentType: cached.ContentType,
			}, nil
		}
		u.cache.Invalidate(cacheKey)
	}

	result, err := u.processFile(file, opts)
	if err != nil {
		return entity.Result{}, err
	}

	cached := dto.CachedFileResult{
		Data:        result.Data,
		ContentType: result.ContentType,
	}
	data, err := json.Marshal(cached)
	if err != nil {
		return entity.Result{}, err
	}
	u.cache.Set(cacheKey, data)
	return result, nil
}

func (u fileUseCase) processFile(file entity.File, opts entity.Options) (entity.Result, error) {
	if !hasProcessingOptions(opts) {
		data, err := os.ReadFile(file.Path)
		if err != nil {
			return entity.Result{}, derr.JoinError("failed to read file", err)
		}
		return entity.Result{
			Data:        data,
			ContentType: file.MimeType,
		}, nil
	}

	p := u.registry.Resolve(file.MimeType)
	if p == nil {
		return entity.Result{}, derr.NewClientError("UNSUPPORTED_FILE_TYPE", "unsupported file type")
	}
	return p.Process(file, opts)
}

func fileListCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("files:%s:list", userID.String())
}

func fileContentCacheKey(userID uuid.UUID, file entity.File, opts entity.Options) string {
	return fmt.Sprintf(
		"files:%s:content:%s:%d:%d:%s:%d:%d",
		userID.String(),
		file.Name,
		opts.Width,
		opts.Height,
		strings.ToLower(opts.Format),
		opts.Quality,
		file.ModifiedAt.UnixNano(),
	)
}

func hasProcessingOptions(opts entity.Options) bool {
	return opts.Width > 0 || opts.Height > 0 || opts.Format != "" || opts.Quality > 0
}
