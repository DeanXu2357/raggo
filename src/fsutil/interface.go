package fsutil

import "io"

// FileStore provides an interface for file system operations
type FileStore interface {
	// ReadFile reads a file and returns its contents
	ReadFile(path string) ([]byte, error)

	// ReadFileAsStream opens a file and returns a reader
	ReadFileAsStream(path string) (io.ReadCloser, error)

	// MakeDirectory creates a new directory and all necessary parents
	MakeDirectory(path string) error

	// RemoveAll removes a path and any children it contains
	RemoveAll(path string) error

	// GetFileStats returns the total count and size of files in a directory
	GetFileStats(path string) (count int, size int64, err error)
}

// Stat represents statistics about files in a directory
type Stat struct {
	Count int   // Number of files
	Size  int64 // Total size in bytes
}
