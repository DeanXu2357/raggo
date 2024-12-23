/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ollama/ollama/api"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

// evaluateCmd represents the evaluate command
var evaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: Evaluate,
}

func init() {
	rootCmd.AddCommand(evaluateCmd)
	evaluateCmd.Flags().StringP("input", "i", "", "Input JSON file path")
	evaluateCmd.MarkFlagRequired("input")
	evaluateCmd.Flags().StringP("evaluate", "e", "", "Evaluation JSON file path")
	evaluateCmd.MarkFlagRequired("evaluate")
	evaluateCmd.Flags().IntP("k", "k", 5, "Number of results to retrieve (default: 5)")
	evaluateCmd.Flags().BoolP("contextual", "c", false, "Add contextual information before storing in database")
	evaluateCmd.Flags().StringP("model", "m", "llama3.2:3b", "LLM model to use for generating context")
}

func Evaluate(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	inputPath, _ := cmd.Flags().GetString("input")
	evaluatePath, _ := cmd.Flags().GetString("evaluate")
	k, _ := cmd.Flags().GetInt("k")
	useContextual, _ := cmd.Flags().GetBool("contextual")
	model, _ := cmd.Flags().GetString("model")

	fmt.Printf("Starting evaluation with:\n")
	fmt.Printf("- Input file: %s\n", inputPath)
	fmt.Printf("- Evaluation file: %s\n", evaluatePath)
	fmt.Printf("- k: %d\n", k)
	fmt.Printf("- Using contextual information: %v\n", useContextual)
	fmt.Printf("- LLM model: %s\n", model)

	// Generate class name with timestamp
	className := fmt.Sprintf("CodeChunk_%d", time.Now().Unix())

	// weaviate connection
	cfg := weaviate.Config{
		Host:   "localhost:8088",
		Scheme: "http",
	}
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		fmt.Printf("Failed to create Weaviate client: %v\n", err)
		return
	}

	// load json file
	jsonFile, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Printf("Failed to read input file: %v\n", err)
		return
	}

	var codebaseRaws []CodeBaseRaw
	err = json.Unmarshal(jsonFile, &codebaseRaws)
	if err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		return
	}

	// Check if CodeChunk class exists
	if err := createWeaviateCollection(ctx, client, className); err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		// Then delete the class
		err = client.Schema().ClassDeleter().WithClassName(className).Do(ctx)
		if err != nil {
			fmt.Printf("Failed to cleanup Weaviate data: %v\n", err)
			return
		}
	}()

	// Calculate total chunks
	var totalChunks int
	for _, codebase := range codebaseRaws {
		totalChunks += len(codebase.Chunks)
	}

	// Create progress bar for importing
	importBar := progressbar.NewOptions(totalChunks,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetDescription("[cyan]Importing chunks[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	// Import data to Weaviate
	batcher := client.Batch().ObjectsBatcher()
	objs := make([]*models.Object, 0, totalChunks)
	for _, codebase := range codebaseRaws {
		for _, chunk := range codebase.Chunks {
			importBar.Add(1)
			properties := map[string]interface{}{
				"docId":        codebase.ID,
				"docUUID":      codebase.UUIDHash,
				"chunkId":      chunk.ID,
				"docIndex":     chunk.Index,
				"chunkContent": chunk.Content,
			}

			content := chunk.Content
			if useContextual {
				// Add contextual information
				context := situateContext(ctx, codebase.Content, chunk.Content, model)
				if context != "" {
					content = fmt.Sprintf("%s\n%s", context, content)
				}
			}

			// Get embedding for the chunk content
			embedding, err := getEmbedding(content)
			if err != nil {
				fmt.Printf("\nFailed to get embedding for chunk %s: %v\n", chunk.ID, err)
				continue
			}

			obj := &models.Object{
				Class:      className,
				Properties: properties,
				Vector:     embedding,
			}

			objs = append(objs, obj)
		}
	}

	resp, err := batcher.WithObjects(objs...).Do(ctx)
	if err != nil {
		fmt.Printf("Failed to import data: %v\n", err)
		return
	}

	fmt.Printf("Successfully imported %d code chunks to %s\n", len(resp), className)

	// Open evaluation file
	evalFile, err := os.Open(evaluatePath)
	if err != nil {
		fmt.Printf("Failed to open evaluation file: %v\n", err)
		return
	}
	defer evalFile.Close()

	// Count total evaluations
	evalScanner := bufio.NewScanner(evalFile)
	var totalEvals int
	for evalScanner.Scan() {
		totalEvals++
	}
	evalFile.Seek(0, 0) // Reset file pointer to beginning

	// Create progress bar for evaluation
	evalBar := progressbar.NewOptions(totalEvals,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetDescription("[cyan]Evaluating queries[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	// Read evaluation file line by line
	scanner := bufio.NewScanner(evalFile)
	const maxCapacity = 4 * 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	var totalScore float64
	var processedEvals int

	for scanner.Scan() {
		evalBar.Add(1)
		var evalRaw EvaluateRaw
		err := json.Unmarshal(scanner.Bytes(), &evalRaw)
		if err != nil {
			fmt.Printf("Failed to parse evaluation line: %v\n", err)
			continue
		}

		// Call RetrievalFunction with query
		retrievedChunks, err := RetrievalFunction(ctx, client, evalRaw.Query, k, className)
		if err != nil {
			fmt.Printf("Failed to retrieve chunks for query: %v\n", err)
			continue
		}

		// Calculate score for this evaluation
		var matchCount int
		for _, retrievedChunk := range retrievedChunks {
			for _, goldenChunk := range evalRaw.GoldenChunkUUIDs {
				goldenUUID := goldenChunk.DocUUID
				goldenIndex := goldenChunk.Index

				// Compare UUID and index directly
				if retrievedChunk.DocUUID == goldenUUID && retrievedChunk.Index == goldenIndex {
					matchCount++
					break
				}
			}
		}

		score := float64(matchCount) / float64(len(evalRaw.GoldenChunkUUIDs))
		totalScore += score
		processedEvals++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading evaluation file: %v\n", err)
		return
	}

	if processedEvals > 0 {
		averageScore := (totalScore / float64(processedEvals)) * 100
		fmt.Printf("\nEvaluation Results (k=%d):\n", k)
		fmt.Printf("Total evaluations: %d\n", processedEvals)
		fmt.Printf("Average score: %.2f%%\n", averageScore)
	} else {
		fmt.Println("No evaluations were processed")
	}

	fmt.Printf("Successfully deleted class %s\n", className)
}

func createWeaviateCollection(ctx context.Context, client *weaviate.Client, className string) error {
	schema, err := client.Schema().Getter().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schema: %v\n", err)
	}

	classExists := false
	for _, class := range schema.Classes {
		if class.Class == className {
			classExists = true
			break
		}
	}

	if !classExists {
		// Create Weaviate schema if not exists
		classObj := &models.Class{
			Class:      className,
			Vectorizer: "none", // We'll use custom vectors
			Properties: []*models.Property{
				{
					Name:     "docId",
					DataType: []string{"string"},
				},
				{
					Name:     "docUUID",
					DataType: []string{"string"},
				},
				{
					Name:     "chunkId",
					DataType: []string{"string"},
				},
				{
					Name:     "docIndex",
					DataType: []string{"number"},
				},
				{
					Name:     "chunkContent",
					DataType: []string{"text"},
				},
			},
		}

		// Create schema
		err = client.Schema().ClassCreator().WithClass(classObj).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to create schema: %v\n", err)
		}
		fmt.Printf("Created schema: %s\n", className)
	}

	return nil
}

type CodeBaseRaw struct {
	ID       string           `json:"doc_id"`
	UUIDHash string           `json:"original_uuid"`
	Content  string           `json:"content"`
	Chunks   []CodeBaseChunks `json:"chunks"`
}

type CodeBaseChunks struct {
	ID      string `json:"chunk_id"`
	Index   int64  `json:"original_index"`
	Content string `json:"content"`
}

type EvaluateRaw struct {
	Query            string     `json:"query"`
	GoldenChunkUUIDs []ChunkRef `json:"golden_chunk_uuids"`
}

type ChunkRef struct {
	DocUUID string
	Index   int64
}

func (e *ChunkRef) UnmarshalJSON(data []byte) error {
	var temp []interface{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) != 2 {
		return fmt.Errorf("ChunkRef must have exactly 2 elements")
	}

	uuidStr, ok := temp[0].(string)
	if !ok {
		return fmt.Errorf("first element must be a string")
	}

	index, ok := temp[1].(float64)
	if !ok {
		return fmt.Errorf("second element must be a number")
	}

	e.DocUUID = uuidStr
	e.Index = int64(index)

	return nil
}

var ollamaClient *api.Client

func init() {
	baseURL, err := url.Parse("http://localhost:11434")
	if err != nil {
		fmt.Printf("Failed to parse Ollama URL: %v\n", err)
		os.Exit(1)
	}
	ollamaClient = api.NewClient(baseURL, http.DefaultClient)
}

func getEmbedding(text string) ([]float32, error) {
	// Create embeddings request
	req := api.EmbeddingRequest{
		Model:  "nomic-embed-text",
		Prompt: text,
	}

	// Call Ollama API using the client
	resp, err := ollamaClient.Embeddings(context.Background(), &req)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %v", err)
	}

	// Convert float64 to float32
	embedding32 := make([]float32, len(resp.Embedding))
	for i, v := range resp.Embedding {
		embedding32[i] = float32(v)
	}

	return embedding32, nil
}

func RetrievalFunction(ctx context.Context, client *weaviate.Client, query string, k int, className string) ([]RetrievalChunk, error) {
	// Get embedding for query
	queryEmbedding, err := getEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %v", err)
	}

	// Perform semantic search using vector
	fields := []graphql.Field{
		{Name: "docUUID"},
		{Name: "chunkId"},
		{Name: "docIndex"},
		{Name: "chunkContent"},
	}

	result, err := client.GraphQL().Get().
		WithClassName(className).
		WithFields(fields...).
		WithNearVector(client.GraphQL().NearVectorArgBuilder().
			WithVector(queryEmbedding)).
		WithLimit(k).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %v", err)
	}

	// Extract results
	chunks := make([]RetrievalChunk, 0)
	if result.Data == nil {
		return chunks, nil
	}

	// Get the results from the GraphQL response
	if getCodeChunk, ok := result.Data["Get"].(map[string]interface{})[className].([]interface{}); ok {
		for _, item := range getCodeChunk {
			if chunk, ok := item.(map[string]interface{}); ok {
				// Parse docIndex
				docIndex, _ := chunk["docIndex"].(float64)

				retrievalChunk := RetrievalChunk{
					DocUUID: chunk["docUUID"].(string),
					Index:   int64(docIndex),
					ChunkID: chunk["chunkId"].(string),
					Content: chunk["chunkContent"].(string),
				}
				chunks = append(chunks, retrievalChunk)
			}
		}
	}

	return chunks, nil
}

type RetrievalChunk struct {
	DocUUID string
	Index   int64
	ChunkID string
	Content string
}

const (
	DOCUMENT_CONTEXT_PROMPT = `
<document>
{doc_content}
</document>
`
	CHUNK_CONTEXT_PROMPT = `
Here is the chunk we want to situate within the whole document
<chunk>
{chunk_content}
</chunk>

Please give a short succinct context to situate this chunk within the overall document for the purposes of improving search retrieval of the chunk.
Answer only with the succinct context and nothing else.
`
)

func situateContext(ctx context.Context, doc, chunk string, generator string) string {
	if generator == "" {
		generator = "llama3.2:3b"
	}

	// Replace template placeholders
	docPrompt := strings.Replace(DOCUMENT_CONTEXT_PROMPT, "{doc_content}", doc, 1)
	chunkPrompt := strings.Replace(CHUNK_CONTEXT_PROMPT, "{chunk_content}", chunk, 1)

	// Combine prompts
	prompt := docPrompt + "\n" + chunkPrompt

	// Create generate request
	req := api.GenerateRequest{
		Model:  generator,
		Prompt: prompt,
	}

	var response string
	// Call Ollama API
	err := ollamaClient.Generate(ctx, &req, func(resp api.GenerateResponse) error {
		response += resp.Response
		return nil
	})

	if err != nil {
		fmt.Printf("Failed to generate context: %v\n", err)
		return ""
	}

	return response
}
