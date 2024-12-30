package translationflow

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
)

const DefaultMaxTokenPerChunk = 1000

type LLMProvider interface {
	TextSplit(ctx context.Context, text string, chunkSize, chunkOverLap int) ([]string, error)
	Reasoning(ctx context.Context, text string) (string, error)
	TokenLength(ctx context.Context, text string) (int, error)
}

// TemplateData holds all the data needed for template execution
type TemplateData struct {
	SourceLang       string
	TargetLang       string
	SourceText       string
	TaggedText       string
	ChunkToTranslate string
	Translation1     string
	TranslationChunk string
	Reflection       string
	ReflectionChunk  string
}

type TranslationFlow struct {
	llmProvider      LLMProvider
	maxTokenPerChunk int
}

func NewTranslationFlow(llmProvider LLMProvider, opts ...Option) *TranslationFlow {
	tf := &TranslationFlow{
		llmProvider:      llmProvider,
		maxTokenPerChunk: DefaultMaxTokenPerChunk,
	}

	for _, opt := range opts {
		opt(tf)
	}

	return tf
}

type Option func(tf *TranslationFlow)

func WithMaxTokenPerChunk(maxTokenPerChunk int) Option {
	return func(tf *TranslationFlow) {
		tf.maxTokenPerChunk = maxTokenPerChunk
	}
}

func (tf *TranslationFlow) Translate(ctx context.Context, text string, sourceLanguage, targetLanguage string) (string, error) {
	tokenLength, err := tf.llmProvider.TokenLength(ctx, text)
	if err != nil {
		return "", fmt.Errorf("failed to get token length: %w", err)
	}

	if tokenLength < tf.maxTokenPerChunk {
		return tf.handleSingleChunkTranslation(ctx, text, sourceLanguage, targetLanguage)
	}
	return tf.handleMultiChunkTranslation(ctx, text, sourceLanguage, targetLanguage, tokenLength)
}

// handleSingleChunkTranslation processes a text that fits within token limits
func (tf *TranslationFlow) handleSingleChunkTranslation(ctx context.Context, text, sourceLanguage, targetLanguage string) (string, error) {
	data := TemplateData{
		SourceLang: sourceLanguage,
		TargetLang: targetLanguage,
		SourceText: text,
	}

	// Step 1: Get initial translation
	translation1, err := tf.getInitialTranslation(ctx, data)
	if err != nil {
		return "", err
	}
	data.Translation1 = translation1

	// Step 2: Get reflection
	reflection, err := tf.getTranslationReflection(ctx, data)
	if err != nil {
		return "", err
	}
	data.Reflection = reflection

	// Step 3: Get improved translation
	return tf.getImprovedTranslation(ctx, data)
}

// handleMultiChunkTranslation processes a text that needs to be split into chunks
func (tf *TranslationFlow) handleMultiChunkTranslation(ctx context.Context, text, sourceLanguage, targetLanguage string, tokenLength int) (string, error) {
	chunkSize := CalculateChunkSize(tokenLength, tf.maxTokenPerChunk)
	chunks, err := tf.llmProvider.TextSplit(ctx, text, chunkSize, chunkSize/10)
	if err != nil {
		return "", fmt.Errorf("failed to split text: %w", err)
	}

	var translatedChunks []string
	for i, chunk := range chunks {
		taggedText := tf.createTaggedText(text, chunk)

		data := TemplateData{
			SourceLang:       sourceLanguage,
			TargetLang:       targetLanguage,
			TaggedText:       taggedText,
			ChunkToTranslate: chunk,
		}

		translatedChunk, err := tf.processChunk(ctx, data, i)
		if err != nil {
			return "", err
		}
		translatedChunks = append(translatedChunks, translatedChunk)
	}

	return strings.Join(translatedChunks, " "), nil
}

// processChunk handles the translation process for a single chunk in multi-chunk mode
func (tf *TranslationFlow) processChunk(ctx context.Context, data TemplateData, chunkIndex int) (string, error) {
	// Step 1: Get initial translation for chunk
	translationChunk, err := tf.getMultiChunkInitialTranslation(ctx, data, chunkIndex)
	if err != nil {
		return "", err
	}
	data.TranslationChunk = translationChunk

	// Step 2: Get reflection for chunk
	reflectionChunk, err := tf.getMultiChunkReflection(ctx, data, chunkIndex)
	if err != nil {
		return "", err
	}
	data.ReflectionChunk = reflectionChunk

	// Step 3: Get improved translation for chunk
	return tf.getMultiChunkImprovedTranslation(ctx, data, chunkIndex)
}

// Template execution helpers
func (tf *TranslationFlow) executeTemplates(systemTmpl, promptTmpl string, data TemplateData) (string, error) {
	var systemBuf, promptBuf bytes.Buffer

	sysT := template.Must(template.New("system").Parse(systemTmpl))
	if err := sysT.Execute(&systemBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute system template: %w", err)
	}

	prmptT := template.Must(template.New("prompt").Parse(promptTmpl))
	if err := prmptT.Execute(&promptBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return promptBuf.String(), nil
}

// Single chunk translation helpers
func (tf *TranslationFlow) getInitialTranslation(ctx context.Context, data TemplateData) (string, error) {
	prompt, err := tf.executeTemplates(
		OneChunkInitialTranslationSystemMessageTmpl,
		OneChunkInitialTranslationPromptTmpl,
		data,
	)
	if err != nil {
		return "", fmt.Errorf("failed to prepare initial translation templates: %w", err)
	}

	translation, err := tf.llmProvider.Reasoning(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get initial translation: %w", err)
	}
	return translation, nil
}

func (tf *TranslationFlow) getTranslationReflection(ctx context.Context, data TemplateData) (string, error) {
	prompt, err := tf.executeTemplates(
		OneChunkReflectOnTranslationSystemMessageTmpl,
		OneChunkReflectOnTranslationPromptTmpl,
		data,
	)
	if err != nil {
		return "", fmt.Errorf("failed to prepare reflection templates: %w", err)
	}

	reflection, err := tf.llmProvider.Reasoning(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get translation reflection: %w", err)
	}
	return reflection, nil
}

func (tf *TranslationFlow) getImprovedTranslation(ctx context.Context, data TemplateData) (string, error) {
	prompt, err := tf.executeTemplates(
		OneChunkImprovementTranslationSystemMessageTmpl,
		OneChunkImprovementTranslationPromptTmpl,
		data,
	)
	if err != nil {
		return "", fmt.Errorf("failed to prepare improvement templates: %w", err)
	}

	improvedTranslation, err := tf.llmProvider.Reasoning(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get improved translation: %w", err)
	}
	return improvedTranslation, nil
}

// Multi-chunk translation helpers
func (tf *TranslationFlow) createTaggedText(fullText, currentChunk string) string {
	return strings.Replace(
		fullText,
		currentChunk,
		"<TRANSLATE_THIS>"+currentChunk+"</TRANSLATE_THIS>",
		1,
	)
}

func (tf *TranslationFlow) getMultiChunkInitialTranslation(ctx context.Context, data TemplateData, chunkIndex int) (string, error) {
	prompt, err := tf.executeTemplates(
		MultiChunkTranslationSystemMessageTmpl,
		MultiChunkTranslationPromptTmpl,
		data,
	)
	if err != nil {
		return "", fmt.Errorf("failed to prepare initial translation templates for chunk %d: %w", chunkIndex, err)
	}

	translation, err := tf.llmProvider.Reasoning(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get translation for chunk %d: %w", chunkIndex, err)
	}
	return translation, nil
}

func (tf *TranslationFlow) getMultiChunkReflection(ctx context.Context, data TemplateData, chunkIndex int) (string, error) {
	prompt, err := tf.executeTemplates(
		MultiChunkReflectionSystemMessageTmpl,
		MultiChunkReflectionPromptTmpl,
		data,
	)
	if err != nil {
		return "", fmt.Errorf("failed to prepare reflection templates for chunk %d: %w", chunkIndex, err)
	}

	reflection, err := tf.llmProvider.Reasoning(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get reflection for chunk %d: %w", chunkIndex, err)
	}
	return reflection, nil
}

func (tf *TranslationFlow) getMultiChunkImprovedTranslation(ctx context.Context, data TemplateData, chunkIndex int) (string, error) {
	prompt, err := tf.executeTemplates(
		MultiChunkImprovementSystemMessageTmpl,
		MultiChunkImprovementPromptTmpl,
		data,
	)
	if err != nil {
		return "", fmt.Errorf("failed to prepare improvement templates for chunk %d: %w", chunkIndex, err)
	}

	improvedTranslation, err := tf.llmProvider.Reasoning(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get final translation for chunk %d: %w", chunkIndex, err)
	}
	return improvedTranslation, nil
}

func (tf *TranslationFlow) CreateTaggedTextForTest(fullText, currentChunk string) string {
	return tf.createTaggedText(fullText, currentChunk)
}

func (tf *TranslationFlow) ExecuteTemplatesForTest(systemTmpl, promptTmpl string, data TemplateData) (string, error) {
	return tf.executeTemplates(systemTmpl, promptTmpl, data)
}

// CalculateChunkSize calculates the size of each chunk based on the total token count and limit.
//
// Parameters:
//
//	tokenCount: The total number of tokens
//	tokenLimit: The maximum number of tokens allowed per chunk
//
// Returns:
//
//	The calculated chunk size
//
// Description:
//
//	This function calculates the chunk size based on the given token count and token limit.
//	If the token count is less than or equal to the limit, it returns the token count as the chunk size.
//	Otherwise, it calculates the number of chunks needed to accommodate all tokens within the limit.
//	The chunk size is determined by dividing the total token count by the number of chunks.
//	If there are remaining tokens after division, they are distributed evenly across the chunks.
func CalculateChunkSize(tokenCount, tokenLimit int) int {
	if tokenCount <= tokenLimit {
		return tokenCount
	}

	// Calculate required number of chunks
	// Using (tokenCount + tokenLimit - 1) for ceiling division
	numChunks := (tokenCount + tokenLimit - 1) / tokenLimit

	// Calculate base chunk size
	chunkSize := tokenCount / numChunks

	// Handle remaining tokens
	remainingTokens := tokenCount % tokenLimit
	if remainingTokens > 0 {
		chunkSize += remainingTokens / numChunks
	}

	return chunkSize
}
