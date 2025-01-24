package fsutil

import (
	"io"
	"os"
)

// LocalFileStore implements FileStore using the local filesystem
type LocalFileStore struct {
	// No fields needed as we're using the standard library directly
}

// NewLocalFileStore creates a new LocalFileStore
func NewLocalFileStore() FileStore {
	return &LocalFileStore{}
}

func (fs *LocalFileStore) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (fs *LocalFileStore) ReadFileAsStream(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (fs *LocalFileStore) MakeDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func (fs *LocalFileStore) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (fs *LocalFileStore) GetFileStats(path string) (count int, size int64, err error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, 0, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				return 0, 0, err
			}
			count++
			size += info.Size()
		}
	}

	return count, size, nil
}
