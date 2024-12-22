package rag

import (
	"context"
	"io"
)

func EvaluateByAnthropicDataSets(ctx context.Context, evaluateDataSet io.Reader, ragInstance Service, k int) error {
	return nil
}

type AnthropicEvaluateSet struct {
	Query            string   `json:"query"`
	Answer           string   `json:"answer"`
	GoldenDocUUIDs   []string `json:"golden_doc_uuids"`
	GoldenChunkUUIDS []string `json:"golden_chunk_uuids"`
	GoldenDocuments  []Document
}
