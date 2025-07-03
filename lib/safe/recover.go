package safe

import (
	"context"

	"http-diff/lib/logger"

	"go.uber.org/zap"
)

func Recovery(panicWriter func(message any)) {
	if err := recover(); err != nil {
		if panicWriter != nil {
			panicWriter(err)
		}
	}
}

func RecoveryWithLogger(f func(), ctx context.Context, tag any) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(ctx, "Recovered from panic", zap.Any("error", err), zap.Any("tag", tag))
		}
	}()

	f()
}

func RecoveryWithLoggerAndCallback(f func(), ctx context.Context, tag any, callback func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(ctx, "Recovered from panic", zap.Any("error", err), zap.Any("tag", tag))
			if callback != nil {
				defer callback()
			}
		}
	}()

	f()
}
