package rag

import (
	"context"
	"io"
)

type Service interface {
	ImportFile(ctx context.Context, name string, file io.Reader) error
	Query(ctx context.Context, str string, k int) ([]Chunks, error)
}

type Document struct {
	ID      int64
	Name    string
	Content string
	Chunks  []Chunks
}

type Chunks struct {
	ID           int64
	Index        int64  // the index of the chunk in the document
	DocumentName string // the document name of the chunk belongs to
	Content      string
}
