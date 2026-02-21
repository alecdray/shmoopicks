package auth

import (
	"net/http"
	"shmoopicks/src/internal/core/appctx"
	"shmoopicks/src/internal/core/db"
)

type HttpHandler struct {
	db *db.DB
}

func NewHttpHandler(db *db.DB) *HttpHandler {
	return &HttpHandler{
		db: db,
	}
}

func (h *HttpHandler) GetLoginPage(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
	loginPage := LoginPage()
	loginPage.Render(r.Context(), w)
}
