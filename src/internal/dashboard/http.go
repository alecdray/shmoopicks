package dashboard

import (
	"net/http"
	"shmoopicks/src/internal/core/appctx"
)

type HttpHandler struct {
	// Define fields here
}

func NewHttpHandler() *HttpHandler {
	return &HttpHandler{}
}

func (h *HttpHandler) HandleGetDashboard(ctx appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	dashboardComponent := Dashboard()
	dashboardComponent.Render(r.Context(), w)
}
