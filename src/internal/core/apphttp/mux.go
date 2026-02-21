package apphttp

import (
	"net/http"
	"shmoopicks/src/internal/core/appctx"
	"shmoopicks/src/internal/core/config"
)

type AppHandler interface {
	ServeHTTP(ctx appctx.AppCtx, w http.ResponseWriter, r *http.Request)
}

type AppHandlerFunc func(ctx appctx.AppCtx, w http.ResponseWriter, r *http.Request)

func (f AppHandlerFunc) ServeHTTP(ctx appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

func WrapHandler(handler http.Handler, middlewares ...Middleware) AppHandlerFunc {
	return ApplyMiddleware(func(ctx appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}, middlewares...)
}

type wrappedMux struct {
	mux        *http.ServeMux
	config     config.Config
	middleware []Middleware
}

func NewWrappedMux(config config.Config, middlewares ...Middleware) *wrappedMux {
	return &wrappedMux{
		mux:        http.NewServeMux(),
		config:     config,
		middleware: middlewares,
	}
}

func (wm *wrappedMux) wrapHandler(handler AppHandler) http.HandlerFunc {
	handlerFunc := ApplyMiddleware(handler.ServeHTTP, wm.middleware...)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appctx.NewAppCtx(r.Context(), wm.config)
		handlerFunc(ctx, w, r)
	}
}

func (wm *wrappedMux) HandleFunc(pattern string, handler AppHandler, middlewares ...Middleware) {
	handler = ApplyMiddleware(handler.ServeHTTP, middlewares...)
	wm.mux.HandleFunc(pattern, wm.wrapHandler(handler))
}

func (wm *wrappedMux) Handle(pattern string, handler AppHandler, middlewares ...Middleware) {
	handler = ApplyMiddleware(handler.ServeHTTP, middlewares...)
	wm.mux.Handle(pattern, wm.wrapHandler(handler))
}

func (wm *wrappedMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wm.mux.ServeHTTP(w, r)
}

func (wm *wrappedMux) Use(pattern string, mux *wrappedMux) {
	wm.mux.Handle(pattern, mux)
}
