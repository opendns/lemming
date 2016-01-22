package main

import (
	"github.com/opendns/lemming/lib/log"
	"github.com/opendns/lemming/lib/sys"
	"regexp"
	"strconv"
)

const TracePipe = "/sys/kernel/debug/tracing/trace_pipe"

/* The output of trace_pipe is a never-ending stream.  It looks like this:
 *
 *            who-21651 [001] .... 1299466.655190: sys_kill(pid: 45db, sig: 0)
 *            who-21651 [001] .... 1299466.655197: sys_kill -> 0x0
 */

// Entry regexp:
//   $1 -> signalling process task name
//   $2 -> signalling process id (base 10)
//   $3 -> target process id (base 16)
//   $4 -> signal number (base 16)
var entryRegexp = regexp.MustCompile(`\s*(\w+)-(\d+).*?: sys_kill\(pid: ([0-9a-f]+), sig: ([0-9a-f]+)`)

// Exit regex:
//   $1 -> signalling process task name
//   $2 -> signalling process id (base 16)
//   $3 -> kill(2) system call return value (base 16)
var exitRegexp = regexp.MustCompile(`\s*(\w+)-(\d+).*?: sys_kill -> 0x([0-9a-f]+)`)

// WatchTracePipe will read system call info from the kernel trace stream and
// report interesting signals sent by kill(2).
//
func WatchTracePipe() {
	reader := NewPipeReader(TracePipe)
	if err := reader.Open(); err != nil {
		log.Error("Could not open `%s': %v", TracePipe, err) // panics
	}
	defer reader.Close()

	for {
		entryline, err := reader.ReadLine()
		if err != nil {
			log.Warning("Cannot continue without trace pipe")
			break
		}

		match := entryRegexp.FindStringSubmatch(entryline)
		if match != nil {
			name := match[1]
			spid, _ := strconv.ParseInt(match[2], 10, 0)
			tpid, _ := strconv.ParseInt(match[3], 16, 0)
			signal, _ := strconv.ParseInt(match[4], 16, 0)

			trace := NewSyscallTrace()
			trace.SName = name
			trace.SPid = spid
			trace.TPid = tpid
			trace.Signal = signal

			// If the source process was something fleeting like kill(3), it may not
			// still be in /proc.  Try to get its info ASAP.
			if trace.Signal != 0 {
				status, err := sys.GetProcStatus(int(trace.SPid))
				if err == nil && status != nil {
					trace.SUid = status.Uid
					trace.SEuid = status.Euid
				} else {
					log.Debug("Couldn't get calling process UID: %v", err)
				}
			}

			// Read the next line, which should be the exit of the same system call
			exitline, err := reader.ReadLine()
			if err != nil {
				log.Warning("Cannot continue without trace pipe")
				break
			}

			// Eval second line and make sure it's for the same process
			match = exitRegexp.FindStringSubmatch(exitline)
			if match == nil {
				log.Warning("Did not see expected sys_kill_exit after sys_kill_entry: malformed exit line:")
				log.Warning("ENTRY: %s", entryline)
				log.Warning(" EXIT: %s", exitline)
				continue
			}
			name = match[1]
			spid, _ = strconv.ParseInt(match[2], 10, 0)
			exitValue, _ := strconv.ParseInt(match[3], 16, 0)
			if name != trace.SName || spid != trace.SPid {
				log.Warning("Did not see expected sys_kill_exit after sys_kill_entry: mismatched process info:")
				log.Warning("ENTRY: %s", entryline)
				log.Warning(" EXIT: %s", exitline)
				continue
			}
			trace.ExitValue = exitValue

			if trace.Signal == 0 {
				// Calling kill with signal 0 only asks the kernel whether a process
				// is alive -- the target process does not receive a signal.
				// These are common and normally not worth logging
				log.Debug("0-signal ('process ping') detected: %s", trace.String())
			} else {
				// Saw a real signal
				log.Info("Signal detected: %s", trace.String())
			}
		}
	}
}
