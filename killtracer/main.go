package main

import (
	"github.com/opendns/lemming/lib/log"
	"os/signal"
	"syscall"
)

func main() {
	signal.Ignore(syscall.SIGWINCH)
	e := NewKillTraceEnabler()
	log.Info("Enabling syscall_kill tracing")
	e.Enable()
	defer e.Disable()
	log.Info("Starting syscall_kill trace watcher thread")
	go e.Watch()

	log.Info("Watching trace pipe for kill signals")
	WatchTracePipe()

	log.Info("Restoring old syscall_kill tracing values")
	e.EndWatch()
}
