package main

import (
	"github.com/opendns/lemming/lib/log"
	"os/signal"
	"syscall"
)

func main() {
	//log.SetDebug(true)
	signal.Ignore(syscall.SIGWINCH)
	signal.Ignore(syscall.SIGHUP)

	go WatchDebugSettings()

	log.Info("Watching trace pipe for kill signals")
	WatchTracePipe()

	// Should be unreachable
	log.Error("Unexpected exit from WatchTracePipe()")
}
