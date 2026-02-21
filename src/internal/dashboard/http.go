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

func (h *HttpHandler) GetDashboardPage(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
	dashboardPage := DashboardPage()
	dashboardPage.Render(r.Context(), w)
}
