package main

import (
	"fmt"
	"github.com/opendns/lemming/lib/log"
	"io/ioutil"
	"time"
)

const TraceKillEnterEnable string = "/sys/kernel/debug/tracing/events/syscalls/sys_enter_kill/enable"
const TraceKillExitEnable string = "/sys/kernel/debug/tracing/events/syscalls/sys_exit_kill/enable"

// How frequently we should write the kernel debug settings
const WatchTime = 10 * time.Second

// Enables kernel tracing for the kill(2) system call, both entry and exit.
// This function never exits -- it will continue to do so every 10 seconds
// until the program is killed.
//
func WatchDebugSettings() {
	log.Info("Enabling kernel sys_kill tracing")
	for {
		err := setKernelToggle(TraceKillEnterEnable, 1)
		if err != nil {
			log.Warning("Watcher can't set `%s': %v", TraceKillEnterEnable, err)
			continue
		}
		err = setKernelToggle(TraceKillExitEnable, 1)
		if err != nil {
			log.Warning("Watcher can't set `%s': %v", TraceKillExitEnable, err)
			continue
		}
		log.Debug("DebugSettingsWatcher: sleeping %d seconds", WatchTime/time.Second)
		time.Sleep(WatchTime)
	}
}

// setKernelToggle writes the given value to the supplied file.  If err is non-
// nil then something went wrong and it was probably not set.
//
func setKernelToggle(file string, val int) error {
	log.Debug("Writing `%d` to `%s'", val, file)
	if err := ioutil.WriteFile(file, []byte(fmt.Sprintf("%d\n", val)), 0644); err != nil {
		log.Warning("Could not write `%d' to `%s': %v", val, file, err)
		return err
	}
	return nil
}
