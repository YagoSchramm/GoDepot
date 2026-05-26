package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
)

type SucessfulResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewSuccessfulResponse() SucessfulResponse {
	return SucessfulResponse{
		Code:    "SUCCESS",
		Message: "success",
	}
}

func Write(w http.ResponseWriter, v any) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return derr.JoinError("failed to marshal response body", err)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	if err != nil {
		return derr.JoinError("failed to write response body", err)
	}

	return nil
}

func HandleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var clientErr derr.ClientError
	if errors.As(err, &clientErr) {
		status := http.StatusBadRequest
		switch clientErr.Code {
		case derr.NotFoundError.Code:
			status = http.StatusNotFound
		case derr.UnauthorizedError.Code:
			status = http.StatusUnauthorized
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(clientErr)
		return
	}

	var repositoryErr derr.RepositoryError
	if errors.As(err, &repositoryErr) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(derr.InternalServerError)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(derr.InternalServerError)
}
