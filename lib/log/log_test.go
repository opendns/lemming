package log_test

import (
	"github.com/opendns/lemming/lib/log"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"
)

func TestStdout(t *testing.T) {
	log.Init()
	if log.Writer() != os.Stdout {
		t.Error("Default Init() did not use os.Stdout")
	}
}

func TestStderr(t *testing.T) {
	log.InitWithStderr()
	if log.Writer() != os.Stderr {
		t.Error("InitWithStderr() did not use os.Stderr")
	}
}

func TestDebug(t *testing.T) {
	log.SetDebug(false)
	testSomeLogMethod(t, log.Debug, "DEBUG", false)
	log.SetDebug(true)
	testSomeLogMethod(t, log.Debug, "DEBUG", true)
	log.SetDebug(false)
	testSomeLogMethod(t, log.Debug, "DEBUG", false)
}

func TestInfo(t *testing.T) {
	testSomeLogMethod(t, log.Info, "INFO", true)
}

func TestWarning(t *testing.T) {
	testSomeLogMethod(t, log.Warning, "WARNING", true)
}

func TestError(t *testing.T) {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0777)
	if err != nil {
		t.Skip("Cannot open /dev/null")
	}
	defer devnull.Close()

	// Must reinitialize the logger since it points to a stale pipe
	log.InitWithWriter(devnull)

	boom := "BOOM"
	defer func() {
		if r := recover(); r == nil {
			t.Error("log.Error() did not panic, but it should have")
		} else if !strings.Contains(r.(string), boom) {
			t.Errorf("Got unexpected panic from log.Error(): %s", r.(string))
		}
	}()
	log.Error(boom)
}

// Signature of all the logging functions
type LogMethod func(string, ...interface{})

// Helper function for the logging tests; sets up an io.Pipe to validate output
func testSomeLogMethod(t *testing.T, fn LogMethod, level string, expectOutput bool) {
	r, w := io.Pipe()
	defer r.Close()
	log.InitWithWriter(w)

	// Generate log message
	rs := randomString()
	go func(fn LogMethod, rs string, w io.WriteCloser) {
		fn(rs)
		w.Close()
	}(fn, rs, w)

	// Check we got the message
	var output []byte = make([]byte, 1024)
	_, readErr := r.Read(output)
	if readErr != nil && readErr != io.EOF {
		t.Fatalf("Cannot read log output from io.Pipe: %v", readErr)
	}
	if readErr == io.EOF {
		if expectOutput {
			// This is what we wanted
			t.Fatalf("Got EOF when output was expected")
		} else {
			return
		}
	}
	t.Logf("Log output: <<<%s>>>", string(output))
	if !strings.Contains(string(output), rs) {
		t.Error("Log output did not have message")
	}
	if !strings.Contains(string(output), level) {
		t.Error("Log output did not have expected level")
	}
}

// This must not have a % in it or else it will disrupt formatting
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ/:_@$()!* ")

// randomString returns a random 20-character string consisting of things you
// might see in a log message.
func randomString() string {
	r := make([]rune, 20)
	for i := range r {
		r[i] = letters[rand.Intn(len(letters))]
	}
	return string(r)
}
