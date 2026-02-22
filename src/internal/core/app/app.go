package app

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNoValue = errors.New("no value")
)

type App struct {
	config Config
	claims *Claims
}

func NewApp(config Config) App {
	return App{config: config}
}

func (app App) Config() Config {
	return app.config
}

func (app *App) SetClaims(w http.ResponseWriter, claims Claims) error {
	app.claims = &claims
	err := app.claims.Save(app.config.JwtSecret, w)
	if err != nil {
		return fmt.Errorf("failed to save JWT: %w", err)
	}
	return nil
}

func (app *App) HasClaims() bool {
	return app.claims != nil
}

func (app App) GetClaims() (Claims, error) {
	if app.claims == nil {
		return Claims{}, ErrNoValue
	}
	return *app.claims, nil
}

func (app *App) UpdateClaims(w http.ResponseWriter, fn func(Claims) Claims) error {
	if app.claims == nil {
		return ErrNoValue
	}
	return app.SetClaims(w, fn(*app.claims))
}

func (app *App) DeleteClaims(w http.ResponseWriter) {
	app.claims.Delete(w)
	app.claims = nil
}
