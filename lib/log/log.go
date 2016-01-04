// The lemming/log package is your standard logger.  It is designed to be
// extremely simple to use.  It requires nothing more than an import, unless
// you want to be fancy and log to stderr or your own creation.
//
// Example:
//
//     import "github.com/opendns/lemming/lib/log"
//
//     func main() {
//             log.Info("Starting")
//             if err := doSomething(); err != nil {
//                     log.Warning("Something failed: %v", err)
//             }
//             if err := doSomethingReallyImportant(); err != nil {
//			log.Error("We're totally broken: %v", err)
//             }
//             log.Info("Exiting")
//      }
//
package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

var logger struct {
	Initialized bool
	Writer      io.Writer
}

// Writer returns the writer currently selected by the logger output.
func Writer() io.Writer {
	return logger.Writer
}

// Init initializes the global logger.  Output will be sent to stdout.  This is
// the default and will be called automatically, so normally there is no need
// to call it explicitly.
func Init() {
	InitWithWriter(os.Stdout)
}

// InitWithStderr initializes the global logger.  Output will be sent to
// stderr.
func InitWithStderr() {
	InitWithWriter(os.Stderr)
}

// InitWithWriter initializes the global logger.  Output will be sent to the
// given io.Writer.
func InitWithWriter(w io.Writer) {
	logger.Writer = w
	logger.Initialized = true
}

// Debug logs the given message at DEBUG level.  This is for getting verbose
// detail from the program during development or debugging, whose content would
// not be interesting to an end-user.
func Debug(format string, v ...interface{}) {
	doLog("DEBUG", format, v...)
}

// Info logs the given message at INFO level.  This is for normal program status.
func Info(format string, v ...interface{}) {
	doLog("INFO", format, v...)
}

// Warning logs the given message at WARNING level.  This is for unusual
// occurrences that should be brought to the user's attention.
func Warning(format string, v ...interface{}) {
	doLog("WARNING", format, v...)
}

// Error logs the given message at ERROR level, then panics.  This is used for
// breaking issues (bad program state, unavailable resources, etc.).
func Error(format string, v ...interface{}) {
	msg := formatLog("ERROR", format, v...)
	doRawLog(msg)
	panic(msg) // TODO: Make panic-on-error configurable
}

// Format a log message into a single string.
func formatLog(level string, format string, v ...interface{}) string {
	var msg string
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	} else {
		// Don't run the message through fmt.Sprintf if no args were
		// supplied.  This avoids the %!(MISSING) spam that fmt adds if
		// it sees an unmatched formatting string (e.g. a %s without a
		// matching arg).  This makes it easy to log.Info(someVariable)
		// without worrying about junk in the output should it contain
		// a % sign.
		msg = format
	}
	timestr := time.Now().UTC().Format("2006-01-02 15:04:05.000000")
	return fmt.Sprintf("%s [%s]: %s\n", timestr, level, msg)
}

// Log the given pre-formatted string.
func doRawLog(msg string) {
	if !logger.Initialized {
		Init() // Stdout by default
	}
	if _, err := logger.Writer.Write([]byte(msg)); err != nil {
		panic("Cannot write to log output")
	}
}

// Format and log a message.
func doLog(level string, format string, v ...interface{}) {
	msg := formatLog(level, format, v...)
	doRawLog(msg)
}
