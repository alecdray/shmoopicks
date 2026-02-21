package appctx

import (
	"context"
	"shmoopicks/src/internal/core/config"
)

type AppCtx struct {
	context.Context
	config config.Config
}

func NewAppCtx(ctx context.Context, config config.Config) AppCtx {
	return AppCtx{ctx, config}
}

func (ctx AppCtx) Config() config.Config {
	return ctx.config
}
