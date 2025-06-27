package signal

import (
	"os"
	"os/signal"
	"syscall"
)

func ReceiveShutdownSignal() chan os.Signal {
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	return signalCh
}
