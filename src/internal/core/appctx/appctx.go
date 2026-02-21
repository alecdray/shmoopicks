package appctx

import (
	"context"
	"shmoopicks/src/internal/core/config"
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

func (ctx *Ctx) SetJwt(jwt Claims) {
	ctx.jwt = &jwt
}

func (ctx Ctx) IsAuthenticated() bool {
	return ctx.jwt != nil
}
