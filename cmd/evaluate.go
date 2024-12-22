/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

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
}

func Evaluate(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	inputPath, _ := cmd.Flags().GetString("input")
	evaluatePath, _ := cmd.Flags().GetString("evaluate")

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
	schema, err := client.Schema().Getter().Do(ctx)
	if err != nil {
		fmt.Printf("Failed to get schema: %v\n", err)
		return
	}

	classExists := false
	for _, class := range schema.Classes {
		if class.Class == "CodeChunk" {
			classExists = true
			break
		}
	}

	if !classExists {
		// Create Weaviate schema if not exists
		classObj := &models.Class{
			Class: "CodeChunk",
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
			fmt.Printf("Failed to create schema: %v\n", err)
			return
		}
		fmt.Println("Created CodeChunk schema")
	}

	// Import data to Weaviate
	batcher := client.Batch().ObjectsBatcher()
	var totalChunks int

	objs := make([]*models.Object, 0)
	for _, codebase := range codebaseRaws {
		for _, chunk := range codebase.Chunks {
			properties := map[string]interface{}{
				"docId":        codebase.ID,
				"docUUID":      codebase.UUIDHash,
				"chunkId":      chunk.ID,
				"docIndex":     chunk.Index,
				"chunkContent": chunk.Content,
			}

			obj := &models.Object{
				Class:      "CodeChunk",
				Properties: properties,
			}

			objs = append(objs, obj)
			totalChunks++
		}
	}

	resp, err := batcher.WithObjects(objs...).Do(ctx)
	if err != nil {
		fmt.Printf("Failed to import data: %v\n", err)
		return
	}

	fmt.Printf("Successfully imported %d code chunks to Weaviate\n", len(resp))

	// Open evaluation file
	evalFile, err := os.Open(evaluatePath)
	if err != nil {
		fmt.Printf("Failed to open evaluation file: %v\n", err)
		return
	}
	defer evalFile.Close()

	// Read evaluation file line by line
	scanner := bufio.NewScanner(evalFile)
	const maxCapacity = 4 * 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	var totalScore float64
	var totalEvals int

	for scanner.Scan() {
		var evalRaw EvaluateRaw
		err := json.Unmarshal(scanner.Bytes(), &evalRaw)
		if err != nil {
			fmt.Printf("Failed to parse evaluation line: %v\n", err)
			continue
		}

		// Call RetrievalFunction with query
		retrievedChunks, err := RetrievalFunction(ctx, client, evalRaw.Query, 5)
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
		totalEvals++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading evaluation file: %v\n", err)
		return
	}

	if totalEvals > 0 {
		averageScore := (totalScore / float64(totalEvals)) * 100
		fmt.Printf("Evaluation Results:\n")
		fmt.Printf("Total evaluations: %d\n", totalEvals)
		fmt.Printf("Average score: %.2f%%\n", averageScore)
	} else {
		fmt.Println("No evaluations were processed")
	}

	// Cleanup: Delete all data from Weaviate
	err = client.Schema().ClassDeleter().WithClassName("CodeChunk").Do(ctx)
	if err != nil {
		fmt.Printf("Failed to cleanup Weaviate data: %v\n", err)
		return
	}
	fmt.Println("Successfully cleaned up Weaviate data")
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

func RetrievalFunction(ctx context.Context, client *weaviate.Client, query string, k int) ([]RetrievalChunk, error) {
	// Perform semantic search using nearText operator
	fields := []graphql.Field{
		{Name: "docUUID"},
		{Name: "chunkId"},
		{Name: "docIndex"},
		{Name: "chunkContent"},
	}

	result, err := client.GraphQL().Get().
		WithClassName("CodeChunk").
		WithFields(fields...).
		WithNearText(client.GraphQL().NearTextArgBuilder().
			WithConcepts([]string{query})).
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
	if getCodeChunk, ok := result.Data["Get"].(map[string]interface{})["CodeChunk"].([]interface{}); ok {
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
