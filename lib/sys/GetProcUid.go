package sys

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// Matches Uid line of /proc/$pid/status.
//   $1: UID
//   $2: EUID
var uidRegexp = regexp.MustCompile(`^Uid:\s+(\d+)\s+(\d+)`)

// GetProcUid returns the UID and effective UID for a given process.  It
// requires access to the /proc file system, so it is probably Linux-specific.
// If the information can't be determined, the underlying error is returned
// and UID/EUID will be set to -1.
//
// Possible, but not exhaustive, reasons for failure:
//  - the process went away before you called this
//  - your OS doesn't have /proc
//  - permissions for /proc are too restrictive; see procfs docs:
//       https://www.kernel.org/doc/Documentation/filesystems/proc.txt
//       (specifically "hidepid=" in section 4.1)
//  - the Linux kernel /proc/$pid/status format has changed (tested on 3.13.0)
//
func GetProcUid(pid int) (int, int, error) {
	procfile := fmt.Sprintf("/proc/%d/status", pid)
	s, err := os.Open(procfile)
	if err != nil {
		return -1, -1, err
	}
	defer s.Close()
	sr := bufio.NewReader(s)
	var uid, euid int64
	for {
		line, err := sr.ReadString('\n')
		if err != nil {
			return -1, -1, err
		}
		matches := uidRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		uid, err = strconv.ParseInt(matches[1], 10, 0)
		if err != nil {
			return -1, -1, err
		}
		euid, err = strconv.ParseInt(matches[2], 10, 0)
		if err != nil {
			return -1, -1, err
		}
		break // success
	}
	return int(uid), int(euid), nil
}
