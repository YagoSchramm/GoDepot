package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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
		index:    idx,
		watcher:  syncWatcher,
		registry: registry,
		cache:    fileCache,
	}
}

type fileUseCase struct {
	index    index.FileIndex
	watcher  *watcher.Watcher
	registry *processor.Registry
	cache    cache.Cache
}

const uploadRoot = "files"

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

func (u fileUseCase) UploadFile(ctx context.Context, userID uuid.UUID, request dto.UploadFileRequest) (entity.File, error) {
	if userID == uuid.Nil {
		return entity.File{}, derr.UnauthorizedError
	}
	if request.Reader == nil {
		return entity.File{}, derr.NewBadRequestError("file is required")
	}

	name, err := safeUploadName(request.Name)
	if err != nil {
		return entity.File{}, err
	}

	userDir := filepath.Join(uploadRoot, userID.String())
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return entity.File{}, derr.JoinError("failed to create upload folder", err)
	}

	dstPath := filepath.Join(userDir, name)
	dst, err := os.Create(dstPath)
	if err != nil {
		return entity.File{}, derr.JoinError("failed to create uploaded file", err)
	}
	defer dst.Close()

	contentType, size, err := copyAndDetect(dst, request.Reader, request.ContentType, name)
	if err != nil {
		return entity.File{}, err
	}

	info, err := dst.Stat()
	if err != nil {
		return entity.File{}, derr.JoinError("failed to stat uploaded file", err)
	}

	file := entity.File{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       name,
		Path:       dstPath,
		MimeType:   contentType,
		Size:       size,
		ModifiedAt: info.ModTime(),
	}
	if file.ModifiedAt.IsZero() {
		file.ModifiedAt = time.Now()
	}

	u.index.Add(file)
	u.cache.Invalidate(fileListCacheKey(userID))
	return file, nil
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

func safeUploadName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = path.Base(strings.ReplaceAll(name, "\\", "/"))
	if name == "" || name == "." || name == "/" {
		return "", derr.NewBadRequestError("file name is required")
	}
	return name, nil
}

func copyAndDetect(dst io.Writer, src io.Reader, contentType string, name string) (string, int64, error) {
	var header bytes.Buffer
	headerSize, err := io.Copy(&header, io.LimitReader(src, 512))
	if err != nil {
		return "", 0, derr.JoinError("failed to read uploaded file", err)
	}

	if _, err := dst.Write(header.Bytes()); err != nil {
		return "", 0, derr.JoinError("failed to write uploaded file", err)
	}
	written, err := io.Copy(dst, src)
	if err != nil {
		return "", 0, derr.JoinError("failed to write uploaded file", err)
	}

	detected := strings.TrimSpace(contentType)
	if detected == "" || detected == "application/octet-stream" {
		detected = http.DetectContentType(header.Bytes())
	}
	if detected == "application/octet-stream" {
		if byExt := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); byExt != "" {
			detected = byExt
		}
	}

	return detected, headerSize + written, nil
}
