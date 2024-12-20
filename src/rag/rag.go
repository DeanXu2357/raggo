package rag

import (
	"context"
	"io"
)

type Service interface {
	ImportFile(ctx context.Context, name string, file io.Reader) error
	Query(ctx context.Context, str string) ([]Chunks, error)
}

type Chunks struct {
	DocumentName string // the document name of the chunk belongs to
	Content      string
}
