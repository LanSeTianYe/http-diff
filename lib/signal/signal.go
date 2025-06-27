package signal

import (
	"os"
	"os/signal"
	"syscall"
)

func GetShutdownChannel() chan os.Signal {
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	return signalCh
}
