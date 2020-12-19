package common

import (
	"os"
	"os/signal"
	"syscall"
)

func OnSigQuit(handler func()) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGTSTP, syscall.SIGQUIT)

	go func() {
		<-c
		handler()
		os.Exit(0)
	}()
}
