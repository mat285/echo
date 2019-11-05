package core

import (
	"bufio"
	"io"

	exception "github.com/blend/go-sdk/exception"
)

// LineHandler is a function handler for a scanned line
type LineHandler func(line []byte) error

// ScanLines scans lines from a reader, calling the handler for each line
func ScanLines(rd io.Reader, handler LineHandler) error {
	reader := bufio.NewReader(rd)
	var line []byte
	var lineErr error
	for lineErr != io.EOF {
		line, lineErr = reader.ReadBytes('\n')
		if lineErr != nil && lineErr != io.EOF {
			return exception.New(lineErr)
		}
		// ReadBytes always includes the delimiter so this only happens when
		// there is no newline at EOF
		if len(line) == 0 {
			break
		}
		if err := handler(line[:len(line)-1]); err != nil {
			return exception.New(err)
		}
	}
	return nil
}
