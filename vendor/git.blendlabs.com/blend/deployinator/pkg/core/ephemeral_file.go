package core

import (
	"io/ioutil"
	"os"
)

// NewEphemeralFile returns a new ephemeral file.
func NewEphemeralFile(handle *os.File) *EphemeralFile {
	return &EphemeralFile{
		Handle: handle,
	}
}

// NewTempEphemeralFile returns a new temporary ephemeral file.
func NewTempEphemeralFile() (*EphemeralFile, error) {
	tfs, err := ioutil.TempFile("", "dockerContext")
	if err != nil {
		return nil, err
	}
	return &EphemeralFile{
		Handle: tfs,
	}, nil
}

// EphemeralFile is a file that deletes itself on close.
type EphemeralFile struct {
	Handle *os.File
}

// Write writes to the file.
func (ef EphemeralFile) Write(contents []byte) (int, error) {
	return ef.Handle.Write(contents)
}

// Read reads from the file.
func (ef EphemeralFile) Read(p []byte) (int, error) {
	return ef.Handle.Read(p)
}

// WriteAt writes at a given offset.
func (ef EphemeralFile) WriteAt(b []byte, off int64) (int, error) {
	return ef.Handle.WriteAt(b, off)
}

// ReadAt reads at a given offset.
func (ef EphemeralFile) ReadAt(p []byte, off int64) (int, error) {
	return ef.Handle.ReadAt(p, off)
}

// Seek seeks to a given offset within the file.
func (ef EphemeralFile) Seek(offset int64, whence int) (int64, error) {
	return ef.Handle.Seek(offset, whence)
}

// Stat stats the file.
func (ef EphemeralFile) Stat() (os.FileInfo, error) {
	return ef.Handle.Stat()
}

// Close deletes the file and closes it..
func (ef EphemeralFile) Close() error {
	err := ef.Handle.Close()
	if err != nil {
		return err
	}

	return os.Remove(ef.Handle.Name())
}
