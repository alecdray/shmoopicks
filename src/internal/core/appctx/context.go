package appctx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/config"
)

var (
	ErrNoValue = errors.New("no value")
)

type Ctx struct {
	context.Context
	config config.Config
	claims *Claims
}

func NewCtx(ctx context.Context, config config.Config) Ctx {
	return Ctx{Context: ctx, config: config}
}

func (ctx Ctx) Config() config.Config {
	return ctx.config
}

func (ctx *Ctx) SetClaims(w http.ResponseWriter, claims Claims) error {
	ctx.claims = &claims
	err := ctx.claims.Save(ctx.config.JwtSecret, w)
	if err != nil {
		return fmt.Errorf("failed to save JWT: %w", err)
	}
	return nil
}

func (ctx *Ctx) HasClaims() bool {
	return ctx.claims != nil
}

func (ctx Ctx) GetClaims() (Claims, error) {
	if ctx.claims == nil {
		return Claims{}, ErrNoValue
	}
	return *ctx.claims, nil
}

func (ctx *Ctx) UpdateClaims(w http.ResponseWriter, fn func(Claims) Claims) error {
	if ctx.claims == nil {
		return ErrNoValue
	}
	return ctx.SetClaims(w, fn(*ctx.claims))
}

func (ctx *Ctx) DeleteClaims(w http.ResponseWriter) {
	ctx.claims.Delete(w)
	ctx.claims = nil
}
