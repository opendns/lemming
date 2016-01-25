package main

import (
	"flag"
	"fmt"
	"github.com/opendns/lemming/lib/log"
	"os"
	"os/signal"
	"syscall"
)

const AppVersion = "1.0.0"

func main() {
	//log.SetDebug(true)
	signal.Ignore(syscall.SIGWINCH)
	signal.Ignore(syscall.SIGHUP)

	// Build help menu and print to screen
	version := flag.Bool("v", false, "print current killtracer version.")
	flag.Parse()
	if *version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	// Start Application
	go WatchDebugSettings()

	log.Info("Watching trace pipe for kill signals")
	WatchTracePipe()

	// Should be unreachable
	log.Error("Unexpected exit from WatchTracePipe()")
}
