package log_test

import (
	"flag"
	"github.com/opendns/lemming/lib/log"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	os.Exit(m.Run())
}

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
	testSomeLogMethod(t, log.Debug, "DEBUG")
}

func TestInfo(t *testing.T) {
	testSomeLogMethod(t, log.Info, "INFO")
}

func TestWarning(t *testing.T) {
	testSomeLogMethod(t, log.Warning, "WARNING")
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
func testSomeLogMethod(t *testing.T, fn LogMethod, level string) {
	r, w := io.Pipe()
	defer w.Close()
	defer r.Close()
	log.InitWithWriter(w)

	// Generate log message
	rs := randomString()
	go fn(rs)

	// Check we got the message
	var output []byte = make([]byte, 1024)
	if _, err := r.Read(output); err != nil {
		t.Fatalf("Cannot read log output from io.Pipe: %v", err)
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
