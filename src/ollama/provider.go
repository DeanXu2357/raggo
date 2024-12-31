package ollama

import (
	"context"

	"github.com/tmc/langchaingo/textsplitter"
)

type OllamaProvider struct {
	ollamaClient *Client
	modelName    string
}

func (o *OllamaProvider) TextSplit(ctx context.Context, text string, chunkSize, chunkOverLap int) ([]string, error) {
	spliter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(chunkOverLap),
		textsplitter.WithLenFunc(
			func(s string) int {
				i, err := o.ollamaClient.CountTokens(ctx, o.modelName, s)
				if err != nil {
					// TODO: log error
					return -1
				}
				return i
			},
		),
	)

	return spliter.SplitText(text)
}

func (o *OllamaProvider) Reasoning(ctx context.Context, system string, prompt string) (string, error) {
	return o.ollamaClient.Generate(ctx, o.modelName, system, prompt, map[string]interface{}{
		"temperature": 0.7,
		"top_p":       0.9,
	})
}

func (o *OllamaProvider) TokenLength(ctx context.Context, text string) (int, error) {
	return o.ollamaClient.CountTokens(ctx, o.modelName, text)
}

func NewOllamaProvider(ollamaClient *Client, modelName string) *OllamaProvider {
	return &OllamaProvider{
		ollamaClient: ollamaClient,
		modelName:    modelName,
	}
}
