package main

import (
	"bufio"
	"fmt"
	"github.com/opendns/lemming/lib/log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const TracePipe = "/sys/kernel/debug/tracing/trace_pipe"

/* The output of trace_pipe is a never-ending stream.  It looks like this:
 *
 *            who-21651 [001] .... 1299466.655190: sys_kill(pid: 45db, sig: 0)
 *            who-21651 [001] .... 1299466.655197: sys_kill -> 0x0
 */

// Entry regexp:
//   $1 -> signalling process binary (base 10)
//   $2 -> signalling process id (base 16)
//   $3 -> target process id (base 16)
//   $4 -> signal system call return value (base 16)
var entryRegexp = regexp.MustCompile(`\s*(\w+)-(\d+).*?: sys_kill\(pid: ([0-9a-f]+), sig: ([0-9a-f]+)`)

// Exit regex:
//   $1 -> signalling binary (base 10)
//   $2 -> signalling process id (base 16)
//   $3 -> signal return value (base 16, 0 = success as per the manpage for (2)kill
var exitRegexp = regexp.MustCompile(`\s*(\w+)-(\d+).*?: sys_kill -> 0x([0-9a-f]+)`)

func WatchTracePipe() {
	tracePipeReader, err := os.Open(TracePipe)
	if err != nil {
		log.Error("Could not open `%s': %v", TracePipe, err) // panics
	}
	tp := bufio.NewReader(tracePipeReader)
	for {
		line, err := tp.ReadString('\n')
		if err != nil {
			// Attempt to reopen once -- during testing I saw at
			// least one EINTR, so this might be necessary (erbarret)
			log.Warning("Got error reading trace_pipe: %v", err)
			tracePipeReader.Close()
			tracePipeReader, err := os.Open(TracePipe)
			if err != nil {
				log.Error("Could not reopen `%s': %v", TracePipe, err) // panics
			}
			tp = bufio.NewReader(tracePipeReader)
		}
		log.Debug("Got trace_pipe line: %s", strings.TrimSpace(line))
		matches := entryRegexp.FindStringSubmatch(line)
		if matches != nil {
			log.Debug("Looks like a syscall_kill entry line")
			parseEntryMatch(matches)
		}
		matches = exitRegexp.FindStringSubmatch(line)
		if matches != nil {
			log.Debug("Looks like a syscall_kill exit line")
			parseExitMatch(matches)
		}
	}
}

// hexToDec returns the given hex (base-16) string as decimal.  If the conversion fails,
// returns the original string.
func hexToDec(v string) string {
	value, err := strconv.ParseInt(v, 16, 0)
	if err == nil {
		return fmt.Sprintf("%d", value)
	} else {
		return v
	}
}

func parseEntryMatch(matches []string) {
	log.Info("Saw kill from %s (pid %s), signal target pid %s, signal value %s", matches[1], matches[2], hexToDec(matches[3]), hexToDec(matches[4]))
}

func parseExitMatch(matches []string) {
	log.Info("Signal by %s (pid %s) had return value %s", matches[1], matches[2], hexToDec(matches[3]))
}
