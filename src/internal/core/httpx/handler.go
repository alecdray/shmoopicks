package httpx

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func HandleErrorResponse(ctx context.Context, w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
	slog.ErrorContext(ctx, "http error", "error", err, "status", status)
}
