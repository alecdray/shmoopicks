package appctx

import (
	"context"
	"shmoopicks/src/internal/core/config"
)

type AppCtx struct {
	context.Context
	config config.Config
	jwt    *Claims
}

func NewAppCtx(ctx context.Context, config config.Config) AppCtx {
	return AppCtx{Context: ctx, config: config}
}

func (ctx AppCtx) Config() config.Config {
	return ctx.config
}

func (ctx *AppCtx) SetJwt(jwt Claims) {
	ctx.jwt = &jwt
}

func (ctx AppCtx) IsAuthenticated() bool {
	return ctx.jwt != nil
}
