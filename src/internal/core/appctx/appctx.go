package appctx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/config"
)

var (
	ErrNoJwt = errors.New("no jwt")
)

type Ctx struct {
	context.Context
	config config.Config
	jwt    *Claims
}

func NewCtx(ctx context.Context, config config.Config) Ctx {
	return Ctx{Context: ctx, config: config}
}

func (ctx Ctx) Config() config.Config {
	return ctx.config
}

func (ctx *Ctx) SetJwt(w http.ResponseWriter, jwt Claims) error {
	ctx.jwt = &jwt
	err := ctx.jwt.Save(ctx.config.JwtSecret, w)
	if err != nil {
		return fmt.Errorf("failed to save JWT: %w", err)
	}
	return nil
}

func (ctx *Ctx) HasJwt() bool {
	return ctx.jwt != nil
}

func (ctx Ctx) GetJwt() (Claims, error) {
	if ctx.jwt == nil {
		return Claims{}, ErrNoJwt
	}
	return *ctx.jwt, nil
}

func (ctx *Ctx) UpdateJwt(w http.ResponseWriter, fn func(Claims) Claims) error {
	if ctx.jwt == nil {
		return ErrNoJwt
	}
	return ctx.SetJwt(w, fn(*ctx.jwt))
}

func (ctx *Ctx) DeleteJwt(w http.ResponseWriter) {
	ctx.jwt.Delete(w)
	ctx.jwt = nil
}

func (ctx Ctx) IsAuthenticated() bool {
	return ctx.jwt != nil
}
