package module

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/domain/usecase"
	"github.com/YagoSchramm/GoDepot/domain/usecase/dto"
	"github.com/YagoSchramm/GoDepot/infrastructure/router"
	"github.com/YagoSchramm/GoDepot/infrastructure/security/jwt"
	"github.com/gorilla/mux"
)

func NewFileModule(fileUseCase usecase.FileUseCase) router.Module {
	return &fileModule{
		fileUseCase: fileUseCase,
		name:        "Files",
		path:        "/files",
	}
}

type fileModule struct {
	fileUseCase usecase.FileUseCase
	name        string
	path        string
}

type syncFolderRequest struct {
	Path string `json:"path"`
}

func (m fileModule) Middlewares() []mux.MiddlewareFunc {
	return nil
}

func (m fileModule) Name() string {
	return m.name
}

func (m fileModule) Path() string {
	return m.path
}

func (m fileModule) Routes() []router.RouteDefinition {
	return []router.RouteDefinition{
		{
			Path:        "",
			Description: "List indexed files for the authenticated user",
			Handler:     m.listFiles,
			HttpMethods: []string{http.MethodGet},
			Public:      false,
		},
		{
			Path:        "/content",
			Description: "Serve an indexed file, optionally transformed",
			Handler:     m.getFile,
			HttpMethods: []string{http.MethodGet},
			Public:      false,
		},
		{
			Path:        "/sync-folder",
			Description: "Set the folder that should be synchronized for the authenticated user",
			Handler:     m.setSyncFolder,
			HttpMethods: []string{http.MethodPost},
			Public:      false,
		},
	}
}

func (m fileModule) getFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("user_claims").(*jwt.Claims)
	if !ok {
		router.HandleError(w, derr.UnauthorizedError)
		return
	}

	request, err := parseFileContentRequest(r)
	if err != nil {
		router.HandleError(w, err)
		return
	}

	result, err := m.fileUseCase.GetFile(ctx, claims.UserID, request.Name, entity.Options{
		Width:   request.Width,
		Height:  request.Height,
		Format:  request.Format,
		Quality: request.Quality,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to get file", "error", err)
		router.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=300")
	if _, err := w.Write(result.Data); err != nil {
		slog.ErrorContext(ctx, "failed to write file response", "error", err)
	}
}

func (m fileModule) listFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("user_claims").(*jwt.Claims)
	if !ok {
		router.HandleError(w, derr.UnauthorizedError)
		return
	}

	files, err := m.fileUseCase.ListFiles(ctx, claims.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list files", "error", err)
		router.HandleError(w, err)
		return
	}

	if err := router.Write(w, files); err != nil {
		slog.ErrorContext(ctx, "failed to write response", "error", err)
	}
}

func parseFileContentRequest(r *http.Request) (dto.FileContentRequest, error) {
	query := r.URL.Query()

	width, err := parseOptionalPositiveInt(query.Get("w"), "w")
	if err != nil {
		return dto.FileContentRequest{}, err
	}
	height, err := parseOptionalPositiveInt(query.Get("h"), "h")
	if err != nil {
		return dto.FileContentRequest{}, err
	}
	quality, err := parseOptionalPositiveInt(query.Get("quality"), "quality")
	if err != nil {
		return dto.FileContentRequest{}, err
	}

	name := strings.TrimSpace(query.Get("name"))
	if name == "" {
		return dto.FileContentRequest{}, derr.NewBadRequestError("file name is required")
	}

	return dto.FileContentRequest{
		Name:    name,
		Width:   width,
		Height:  height,
		Format:  strings.ToLower(strings.TrimSpace(query.Get("format"))),
		Quality: quality,
	}, nil
}

func parseOptionalPositiveInt(value string, field string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	number, err := strconv.Atoi(value)
	if err != nil || number < 0 {
		return 0, derr.NewBadRequestError(field + " must be a positive integer")
	}
	if field == "quality" && number > 100 {
		return 0, derr.NewBadRequestError("quality must be between 0 and 100")
	}
	return number, nil
}

func (m fileModule) setSyncFolder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("user_claims").(*jwt.Claims)
	if !ok {
		router.HandleError(w, derr.UnauthorizedError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read request body", "error", err)
		router.HandleError(w, err)
		return
	}

	var request syncFolderRequest
	if err := json.Unmarshal(body, &request); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal request body", "error", err)
		router.HandleError(w, derr.BadRequestError)
		return
	}

	if err := m.fileUseCase.SetSyncRoot(ctx, claims.UserID, request.Path); err != nil {
		slog.ErrorContext(ctx, "failed to set sync folder", "error", err)
		router.HandleError(w, derr.NewBadRequestError(err.Error()))
		return
	}

	if err := router.Write(w, router.NewSuccessfulResponse()); err != nil {
		slog.ErrorContext(ctx, "failed to write response", "error", err)
	}
}
