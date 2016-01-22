package main

import (
	"bufio"
	"github.com/opendns/lemming/lib/log"
	"io"
	"os"
	"strings"
)

// A PipeReader is a simple class which tries to read lines one at a time from
// the supplied file.  It will try to re-open the file ONCE on read errors,
// then give up.
//
type PipeReader struct {
	file string
	r    io.ReadCloser
	rb   *bufio.Reader
}

// NewPipeReader returns a new, unopened PipeReader.
//
func NewPipeReader(file string) *PipeReader {
	return &PipeReader{
		file: file,
	}
}

// Open opens the supplied file on the PipeReader.
//
func (p *PipeReader) Open() (err error) {
	r, err := os.Open(p.file)
	if err == nil {
		p.r = r
		p.rb = bufio.NewReader(p.r)
	}
	return
}

// Close closes the io.Reader opened on the PipeReader.
//
func (p *PipeReader) Close() (err error) {
	if p.r != nil {
		err = p.r.Close()
		p.r = nil
		p.rb = nil
	}
	return
}

// ReadLine reads one line from the PipeReader's opened file.  If it fails,
// it will close and re-open the file once, and try to read a line again.  If
// the second attempt fails then it gives up and the error is returned.
//
func (p *PipeReader) ReadLine() (line string, err error) {
	line, err = p.rb.ReadString('\n')
	if err != nil {
		log.Warning("Error reading `%s'; will try to reopen: %v", p.file, err)
		p.Close()
		err = p.Open()
		if err != nil {
			log.Warning("Could not re-open `%s': %v", p.file, err)
			return
		}
		line, err = p.rb.ReadString('\n')
		if err != nil {
			log.Warning("Error reading `%s' after re-open: %v", p.file, err)
		}
	}
	line = strings.TrimRight(line, "\n")
	log.Debug("Got trace_pipe line: %s", line)
	return
}
