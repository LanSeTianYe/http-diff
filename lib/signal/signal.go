package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"http-diff/lib/logger"

	"go.uber.org/zap"
)

// BlockWaitSignal 阻塞等待系统信号
func BlockWaitSignal(ctx context.Context) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	sig := <-signalCh
	logger.Info(ctx, "receive shutdown signal", zap.String("signal", sig.String()))
}
