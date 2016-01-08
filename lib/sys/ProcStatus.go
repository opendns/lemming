package sys

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

type ProcStatus struct {
	// Task name
	Name string
	// Process ID
	Pid int
	// Parent process ID
	Ppid int
	// User ID
	Uid int
	// Effective user ID
	Euid int
}

// Matches Uid line of /proc/$pid/status.
//   $1: UID
//   $2: EUID
var uidRegexp = regexp.MustCompile(`^Uid:\s+(\d+)\s+(\d+)`)

// Matches task name line of /proc/$pid/status.
//   $1: name
var nameRegexp = regexp.MustCompile(`^Name:\s*(\S+)`)

// Matches parent pid of /proc/$pid/status.
//   $1: ppid
var ppidRegexp = regexp.MustCompile(`^PPid:\s*([0-9]+)`)

// GetProcStatus returns a ProcStatus struct containing information about the
// given pid as documented in /proc/$pid/status.  It requires access to the
// /proc file system, so it is probably Linux-specific.  If the query fails,
// the returned ProcStatus is nil and the error will be set to the cause.
//
// Possible, but not exhaustive, reasons for failure:
//  - the process went away before you called this
//  - your OS doesn't have /proc
//  - permissions for /proc are too restrictive; see procfs docs:
//       https://www.kernel.org/doc/Documentation/filesystems/proc.txt
//       (specifically "hidepid=" in section 4.1)
//  - the Linux kernel /proc/$pid/status format has changed (tested on 3.13.0)
//
func GetProcStatus(pid int) (*ProcStatus, error) {
	status := &ProcStatus{
		Pid:  -1,
		Ppid: -1,
		Uid:  -1,
		Euid: -1,
	}

	procfile := fmt.Sprintf("/proc/%d/status", pid)
	s, err := os.Open(procfile)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	sr := bufio.NewReader(s)
	for {
		line, err := sr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Get uid & euid
		match := uidRegexp.FindStringSubmatch(line)
		if match != nil {
			uid, _ := strconv.ParseInt(match[1], 10, 0)
			euid, _ := strconv.ParseInt(match[2], 10, 0)

			status.Uid = int(uid)
			status.Euid = int(euid)
			continue
		}

		// Get task name
		match = nameRegexp.FindStringSubmatch(line)
		if match != nil {
			status.Name = match[1]
			continue
		}
		match = nameRegexp.FindStringSubmatch(line)
		if match != nil {
			ppid, _ := strconv.ParseInt(match[1], 10, 0)
			status.Ppid = int(ppid)
		}
	}
	return status, nil
}

// GetProcUid returns the UID and effective UID for a given process.  It
// requires access to the /proc file system, so it is probably Linux-specific.
// If the information can't be determined, the underlying error is returned
// and UID/EUID will be set to -1.  See GetProcStatus() for more info.
//
func GetProcUid(pid int) (int, int, error) {
	status, err := GetProcStatus(pid)
	if err != nil {
		return -1, -1, err
	}
	return status.Uid, status.Euid, nil
}
