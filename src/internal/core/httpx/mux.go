package httpx

import (
	"net/http"
	"shmoopicks/src/internal/core/app"
	"shmoopicks/src/internal/core/contextx"
)

type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

func WrapHandler(handler http.Handler, middlewares ...Middleware) HandlerFunc {
	return ApplyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}, middlewares...)
}

type mux struct {
	mux        *http.ServeMux
	app        app.App
	middleware []Middleware
}

func NewMux(app app.App, middlewares ...Middleware) *mux {
	return &mux{
		mux:        http.NewServeMux(),
		app:        app,
		middleware: middlewares,
	}
}

func (wm *mux) wrapHandler(handler Handler) http.HandlerFunc {
	handlerFunc := ApplyMiddleware(handler.ServeHTTP, wm.middleware...)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := contextx.NewContextX(r.Context()).
			WithApp(wm.app)

		handlerFunc(w, r.WithContext(ctx))
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
	wm.Handle(pattern, mux)
}
