package main

import (
	//"bufio"
	"fmt"
	"github.com/opendns/lemming/lib/log"
	"io/ioutil"
	//"os"
	"strconv"
	"strings"
	"time"
)

const TraceKillEnterEnable string = "/sys/kernel/debug/tracing/events/syscalls/sys_enter_kill/enable"
const TraceKillExitEnable string = "/sys/kernel/debug/tracing/events/syscalls/sys_exit_kill/enable"

// Delay in the KillTraceEnabler's check of whether the values are still on
const TraceKillWatchPollTime = 10 * time.Second

// getKernelToggle reads the value of a kernel toggle and returns it as an int.
// If error is non-nil then something went wrong trying to read it and the
// returned value is meaningless.
//
func getKernelToggle(file string) (int, error) {
	valStr, err := ioutil.ReadFile(file)
	if err != nil {
		log.Warning("Could not read contents of `%s': %v", file, err)
		return -1, err
	}
	// Get old value
	val, err := strconv.Atoi(strings.TrimSpace(string(valStr)))
	if err != nil {
		log.Warning("Could not parse contents of `%s' into an int: %v", file, err)
		return -1, err
	}
	return val, nil
}

// setKernelToggle writes the given value to the supplied file.  If err is non-
// nil then something went wrong and it was probably not set.
//
func setKernelToggle(file string, val int) error {
	log.Info("Writing `%d` to `%s'", val, file)
	if err := ioutil.WriteFile(file, []byte(fmt.Sprintf("%d", val)), 0644); err != nil {
		log.Warning("Could not write `%d' to `%s': %v", val, file, err)
		return err
	}
	return nil
}

// enableKernelToggle writes the given integer to the given file.  If the
// supplied *int is non-nil, it will read the old value and save it there.
// Helper function for KillTraceEnabler.Enable.
//
func enableKernelToggle(file string, old *int) error {
	if old != nil {
		val, err := getKernelToggle(file)
		if err != nil {
			return err
		}
		log.Info("Old value of `%s' was %d", file, val)
		*old = val
	}
	return setKernelToggle(file, 1)
}

type KillTraceEnabler struct {
	oldEnterValue int
	oldExitValue  int
	done          chan int
}

// NewKillTraceEnabler initializes and returns a KillTraceEnabler struct.
//
func NewKillTraceEnabler() *KillTraceEnabler {
	var e KillTraceEnabler
	e.done = make(chan int)
	return &e
}

// Enables kernel syscall tracing for the kill system call, both entry and
// exit.  Stores the old values.
//
func (e *KillTraceEnabler) Enable() {
	if enableKernelToggle(TraceKillEnterEnable, &e.oldEnterValue) != nil {
		log.Error("Could not enable syscall_kill entry tracing; are you root?") // panics
	}
	if enableKernelToggle(TraceKillExitEnable, &e.oldExitValue) != nil {
		log.Error("Could not enable syscall_kill exit tracing; are you root?") // panics
	}
}

// Restores kernel syscall tracing to the previous values.
//
func (e *KillTraceEnabler) Disable() {
	if setKernelToggle(TraceKillEnterEnable, e.oldEnterValue) != nil {
		log.Warning("Failed to restore syscall_kill entry tracing to old value (%d)", e.oldEnterValue)
	}
	if setKernelToggle(TraceKillExitEnable, e.oldExitValue) != nil {
		log.Warning("Failed to restore syscall_kill exit tracing to old value (%d)", e.oldExitValue)
	}
}

// Watches the values of the syscall_kill entry and exit trace flags.  If they
// are disabled, will re-enable them and log a warning.  Does not exit until
// the EndWatch() method is called.
//
func (e *KillTraceEnabler) Watch() {
	for {
		select {
		case <-e.done:
			// Somebody called EndWatch()
			return
		default:
			// carry on
		}

		enterTrace, err := getKernelToggle(TraceKillEnterEnable)
		if err != nil {
			log.Warning("Watcher can't read `%s': %v", TraceKillEnterEnable, err)
			continue
		}
		exitTrace, err := getKernelToggle(TraceKillExitEnable)
		if err != nil {
			log.Warning("Watcher can't read `%s': %v", TraceKillExitEnable, err)
			continue
		}
		if enterTrace == 0 || exitTrace == 0 {
			log.Warning("Something disabled syscall_kill tracing! (entry = %d, exit = %d)  Re-enabling", enterTrace, exitTrace)
			if enableKernelToggle(TraceKillEnterEnable, nil) != nil {
				log.Error("Could not re-enable syscall_kill entry tracing!") // panics
			}
			if enableKernelToggle(TraceKillExitEnable, nil) != nil {
				log.Error("Could not re-enable syscall_kill exit tracing!") // panics
			}
		}
		log.Debug("Watcher: syscall_kill enter/exit tracing still enabled; sleeping %d seconds", TraceKillWatchPollTime/time.Second)
		time.Sleep(TraceKillWatchPollTime)
	}

}

// Signals the Watch() goroutine to exit.
//
func (e *KillTraceEnabler) EndWatch() {
	e.done <- 1
}
