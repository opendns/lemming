package sys

import (
	"os"
	"testing"
)

// checkProcFilesystemExists automatically skips the test if /proc doesn't exist, or
// fatals it if it's not accessible at all.
func checkProcFilesystemExists(t *testing.T) {
	procdir, err := os.Open("/proc")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("No /proc file system available for testing")
		} else {
			t.Fatal("Unexpected error reading /proc: %v")
		}
	}
	procdir.Close()
}

// Test that we can read our own UID/EUID from /proc
func TestGetProcUidOnOurself(t *testing.T) {
	checkProcFilesystemExists(t)

	ourUid := os.Getuid()
	ourEuid := os.Geteuid()
	tuid, teuid, err := GetProcUid(os.Getpid())
	if err != nil {
		t.Error("Trouble getting UID from /proc: %v", err)
	}
	if tuid != ourUid {
		t.Error("Returned UID (%d) does not match getuid(2) (%d)", tuid, ourUid)
	}
	if teuid != ourEuid {
		//
		t.Error("Returned EUID (%d) does not match getuid(2) (%d)")
	}
}

// Run test on pid 1 -- in all sane cases this will exist and be owned by root
func TestGetProcUidOnRoot(t *testing.T) {
	checkProcFilesystemExists(t)

	tuid, teuid, err := GetProcUid(1)
	if err != nil {
		if os.IsPermission(err) {
			// This is likely to be the case in a Jenkins container
			t.Skip("No permission to read /proc/1/status")
		}
		t.Error("Unexpected error reading /proc/1/status: %v", err)
	}
	if tuid != 0 {
		t.Error("Expected 0 for root process EUID, got %d", tuid)
	}
	if teuid != 0 {
		t.Error("Expected 0 for root process EUID, got %d", teuid)
	}
}

// Run test on an improbable process ID
func TestGetProcUidOnInvalid(t *testing.T) {
	checkProcFilesystemExists(t)

	tuid, teuid, err := GetProcUid(-2)
	if err == nil {
		t.Error("Expected failure to read /proc/-2, got unexpected success")
	}
	if tuid != -1 {
		t.Error("Expected -1 for UID of failed lookup, got %d", tuid)
	}
	if teuid != -1 {
		t.Error("Expected -1 for EUID of failed lookup, got %d", teuid)
	}
}
