package module

import (
	"go.uber.org/zap"
)

func (ctx *ComposeContext) logCompositionInfo(message string, method *string, url *string, name *string) {
	ctx.webComposer.logger.Info(
		message,
		zap.String("request.url", ctx.httpRequest.URL.String()),
		zap.String("request.method", ctx.httpRequest.Method),
		zap.String("component.url", *url),
		zap.String("component.method", *method),
		zap.String("component.name", *name),
	)
}

func (ctx *ComposeContext) logCompositionError(message string, method *string, url *string, name *string, err error) {
	ctx.webComposer.logger.Error(
		message,
		zap.String("request.url", ctx.httpRequest.URL.String()),
		zap.String("request.method", ctx.httpRequest.Method),
		zap.String("component.url", *url),
		zap.String("component.method", *method),
		zap.String("component.name", *name),
		zap.Error(err),
	)
}
