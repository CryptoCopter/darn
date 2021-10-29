package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

const MimeDrop = "DROP"

var ErrMimeDrop = errors.New("MIME must be dropped")

// MimeMap replaces predefined MIME types with others or requires them to be dropped.
//
//   # An example MimeMap could look like this, comment included:
//   text/html        text/plain
//   text/javascript  text/plain
//   text/mp4         DROP
//
type MimeMap map[string]string

// NewMimeMap creates a new MimeMap based on the Reader's data.
func NewMimeMap(file io.Reader) (mm MimeMap, err error) {
	mm = make(MimeMap)

	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		mmLine := scanner.Text()
		if mmLine == "" || strings.HasPrefix(mmLine, "#") {
			continue
		}

		mmFields := strings.Fields(mmLine)
		if l := len(mmFields); l != 2 {
			err = fmt.Errorf("entry in line %d has %d instead of 2 fields", i, l)
			return
		}

		mmKey, mmVal := mmFields[0], mmFields[1]
		if _, exists := mm[mmKey]; exists {
			err = fmt.Errorf("key \"%s\" from line %d was already defined", mmKey, i)
			return
		} else {
			mm[mmKey] = mmVal
		}
	}

	if scannerErr := scanner.Err(); scannerErr != nil {
		err = scannerErr
		return
	}

	return
}

// MustDrop indicates if a MIME type must be dropped.
func (mm MimeMap) MustDrop(mime string) bool {
	v, exists := mm[mime]
	if !exists {
		return false
	}

	if v == MimeDrop {
		return true
	}

	return false
}

// Substitute returns the replaced MIME type and indicates with an error, if
// the input MIME type must be dropped.
func (mm MimeMap) Substitute(mime string) (mimeOut string, err error) {
	v, exists := mm[mime]
	if !exists {
		mimeOut = mime
	}

	if v == MimeDrop {
		err = ErrMimeDrop
		return
	}

	mimeOut = v
	return
}
