package apphttp

import (
	"net/http"
	"shmoopicks/src/internal/core/appctx"
	"shmoopicks/src/internal/core/config"
)

type Handler interface {
	ServeHTTP(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request)
}

type HandlerFunc func(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request)

func (f HandlerFunc) ServeHTTP(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

func WrapHandler(handler http.Handler, middlewares ...Middleware) HandlerFunc {
	return ApplyMiddleware(func(ctx appctx.Ctx, w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}, middlewares...)
}

type mux struct {
	mux        *http.ServeMux
	config     config.Config
	middleware []Middleware
}

func NewMux(config config.Config, middlewares ...Middleware) *mux {
	return &mux{
		mux:        http.NewServeMux(),
		config:     config,
		middleware: middlewares,
	}
}

func (wm *mux) wrapHandler(handler Handler) http.HandlerFunc {
	handlerFunc := ApplyMiddleware(handler.ServeHTTP, wm.middleware...)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appctx.NewCtx(r.Context(), wm.config)
		handlerFunc(ctx, w, r)
	}
}

func (wm *mux) HandleFunc(pattern string, handler Handler, middlewares ...Middleware) {
	handler = ApplyMiddleware(handler.ServeHTTP, middlewares...)
	wm.mux.HandleFunc(pattern, wm.wrapHandler(handler))
}

func (wm *mux) Handle(pattern string, handler Handler, middlewares ...Middleware) {
	handler = ApplyMiddleware(handler.ServeHTTP, middlewares...)
	wm.mux.Handle(pattern, wm.wrapHandler(handler))
}

func (wm *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wm.mux.ServeHTTP(w, r)
}

func (wm *mux) Use(pattern string, mux *mux) {
	wm.mux.Handle(pattern, mux)
}
