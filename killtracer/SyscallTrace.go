package main

import (
	"fmt"
)

// SyscallTrace holds information about a traced system call.
//
type SyscallTrace struct {
	// Source process ID
	SPid int64
	// Source task name
	SName string
	// Source signalling process user ID
	SUid int
	// Signalling process effective user ID
	SEuid int
	// Target process ID
	TPid int64
	// Signal number
	Signal int64
	// System call exit value
	ExitValue int64
}

// System call return values are 64-bit so technically they can be any number.
// However, for kill(2) they should be only 0 or -1.
const UnlikelyExitValue = 0x123456789012345

// NewSyscallTrace returns a SyscallTrace initialized to unlikely values.
//
func NewSyscallTrace() *SyscallTrace {
	return &SyscallTrace{
		SPid:      -1,
		SUid:      -1,
		SEuid:     -1,
		TPid:      -1,
		Signal:    -1,
		ExitValue: UnlikelyExitValue,
	}
}

// String returns a human-readable representation of a SyscallTrace.
//
func (s *SyscallTrace) String() (str string) {
	str = fmt.Sprintf("signal[%d] exit[0x%X] target[%d] source[%s-%d] ", s.Signal, s.ExitValue, s.TPid, s.SName, s.SPid)
	if s.SUid >= 0 || s.SEuid > 0 {
		str = fmt.Sprintf("%ssourceUid[%d] sourceEuid[%d]", str, s.SUid, s.SEuid)
	} else {
		str = fmt.Sprintf("%ssourceUid[??] sourceEuid[??]", str)
	}
	return
}
