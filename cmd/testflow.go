package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"raggo/src/core/translationflow"
	"raggo/src/infrastructure/integrations/ollama"
	"raggo/src/storage/minioctrl"
	"raggo/src/storage/postgres/chunkctrl"
	"raggo/src/storage/postgres/resourcectrl"
)

var (
	resourceID string
	inputText  string
	sourceLang string
	targetLang string
	country    string
	model      string
)

// testflowCmd represents the testflow command
var testflowCmd = &cobra.Command{
	Use:   "testflow",
	Short: "Test translation flow with resource ID or direct text input",
	Long: `Test the translation flow functionality by providing either a resource ID or direct text input.
When using a resource ID, this command will fetch the resource and its chunks from the database and MinIO.
When using direct text input, the text will be processed directly through the translation flow.`,
	Run: func(cmd *cobra.Command, args []string) {
		if resourceID == "" && inputText == "" {
			fmt.Println("Error: either resource ID or input text is required")
			return
		}
		if resourceID != "" && inputText != "" {
			fmt.Println("Error: cannot provide both resource ID and input text")
			return
		}
		if sourceLang == "" {
			fmt.Println("Error: source language is required")
			return
		}
		if targetLang == "" {
			fmt.Println("Error: target language is required")
			return
		}
		if country == "" {
			fmt.Println("Error: country is required")
			return
		}
		if model == "" {
			fmt.Println("Error: model is required")
			return
		}

		ctx := context.Background()

		// Create ollama client and provider
		httpClient := &http.Client{}
		ollamaClient := ollama.NewClient("http://ollama:11434/api", httpClient)
		provider := ollama.NewOllamaProvider(ollamaClient, model)

		// Create translation flow
		flow := translationflow.NewTranslationFlow(provider)

		// Handle direct text input
		if inputText != "" {
			fmt.Println("Processing direct text input:")
			fmt.Println("-------------------")
			translatedContent, err := flow.Translate(
				ctx,
				inputText,
				sourceLang,
				targetLang,
				country,
			)
			if err != nil {
				fmt.Printf("Error translating text: %v\n", err)
				return
			}
			fmt.Println("Translation result:")
			fmt.Println("-------------------")
			fmt.Println(translatedContent)
			fmt.Println("-------------------")
			return
		}

		// Handle resource ID input
		id, err := strconv.ParseInt(resourceID, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing resource ID: %v\n", err)
			return
		}

		// Initialize PostgreSQL connection
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			viper.GetString("postgres.host"),
			viper.GetString("postgres.user"),
			viper.GetString("postgres.password"),
			viper.GetString("postgres.db"),
			viper.GetString("postgres.port"))
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Initialize services
		resourceService, err := resourcectrl.NewResourceService(db)
		if err != nil {
			log.Fatalf("Failed to create resource service: %v", err)
		}

		chunkService, err := chunkctrl.NewChunkService(db)
		if err != nil {
			log.Fatalf("Failed to create chunk service: %v", err)
		}

		minioService, err := minioctrl.NewMinioService(
			viper.GetString("minio.endpoint"),
			viper.GetString("minio.access_key"),
			viper.GetString("minio.secret_key"),
			false,
		)
		if err != nil {
			log.Fatalf("Failed to create minio service: %v", err)
		}

		// Get resource
		resource, err := resourceService.GetByID(ctx, id)
		if err != nil {
			fmt.Printf("Error getting resource: %v\n", err)
			return
		}
		if resource == nil {
			fmt.Printf("Resource not found: %s\n", resourceID)
			return
		}

		// Get chunks
		chunks, err := chunkService.GetByResourceID(ctx, resource.ID)
		if err != nil {
			fmt.Printf("Error getting chunks: %v\n", err)
			return
		}

		// Process each chunk
		fmt.Println("Processing chunks:")
		for i, chunk := range chunks {
			fmt.Printf("\nProcessing chunk %d/%d (ID: %s)\n", i+1, len(chunks), chunk.ChunkID)

			// Get chunk content from MinioURL
			bucket, objectName := minioService.GetBucketAndObjectFromURL(chunk.MinioURL)
			chunkContent, err := minioService.GetObject(ctx, bucket, objectName)
			if err != nil {
				fmt.Printf("Error getting chunk content: %v\n", err)
				continue
			}

			// Translate chunk
			translatedContent, err := flow.Translate(
				ctx,
				string(chunkContent),
				sourceLang,
				targetLang,
				country,
			)
			if err != nil {
				fmt.Printf("Error translating chunk: %v\n", err)
				continue
			}

			fmt.Printf("Translation result for chunk %s:\n", chunk.ChunkID)
			fmt.Println("-------------------")
			fmt.Println(translatedContent)
			fmt.Println("-------------------")
		}
	},
}

func init() {
	rootCmd.AddCommand(testflowCmd)

	// Add flags
	testflowCmd.Flags().StringVarP(&resourceID, "resource", "r", "", "Resource ID")
	testflowCmd.Flags().StringVarP(&inputText, "text", "x", "", "Direct text input")
	testflowCmd.Flags().StringVarP(&sourceLang, "source", "s", "", "Source language (required)")
	testflowCmd.Flags().StringVarP(&targetLang, "target", "t", "", "Target language (required)")
	testflowCmd.Flags().StringVarP(&country, "country", "c", "", "Target country (required)")
	testflowCmd.Flags().StringVarP(&model, "model", "m", "", "Model to use for translation (required)")

	// Mark flags as required
	testflowCmd.MarkFlagRequired("source")
	testflowCmd.MarkFlagRequired("target")
	testflowCmd.MarkFlagRequired("country")
	testflowCmd.MarkFlagRequired("model")
}
