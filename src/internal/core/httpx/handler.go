package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"shmoopicks/src/internal/core/appctx"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func HandleErrorResponse(ctx appctx.Ctx, w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
	slog.ErrorContext(ctx, "http error", "error", err, "status", status)
}
