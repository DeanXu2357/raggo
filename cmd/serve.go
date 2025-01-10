/*
Copyright © 2024 Dean
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	weaviateClient "github.com/weaviate/weaviate-go-client/v4/weaviate"

	httpHdlr "raggo/handler/http"
	"raggo/src/core/knowledgebase"
	"raggo/src/infrastructure/integrations/ollama"
	jobctrl "raggo/src/infrastructure/job"
	"raggo/src/storage/minioctrl"
	"raggo/src/storage/postgres/chunkctrl"
	pgKnowledgeBase "raggo/src/storage/postgres/knowledgebasectrl"
	"raggo/src/storage/postgres/resourcectrl"
	"raggo/src/storage/weaviate"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a rag server",
	Long:  `The serve command starts an HTTP server that provides the rag service.`,
	Run:   RunServer,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	settingDefaultConfig()
}

func RunServer(cmd *cobra.Command, args []string) {
	// Initialize PostgreSQL connection
	host := viper.GetString("postgres.host")
	user := viper.GetString("postgres.user")
	password := viper.GetString("postgres.password")
	dbname := viper.GetString("postgres.db")
	port := viper.GetString("postgres.port")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	logger := watermill.NewStdLogger(false, false)

	// Initialize AMQP publisher
	publisher, err := amqp.NewPublisher(
		amqp.NewDurableQueueConfig(viper.GetString("amqp.url")),
		logger,
	)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Initialize job repository and service
	jobRepo := jobctrl.NewPostgresJobRepository(db)
	jobService := jobctrl.NewJobService(publisher, jobRepo, logger, nil)

	// Initialize services
	resourceService, err := resourcectrl.NewResourceService(db)
	if err != nil {
		log.Fatalf("Failed to create resource service: %v", err)
	}

	chunkService, err := chunkctrl.NewChunkService(db)
	if err != nil {
		log.Fatalf("Failed to create resource service: %v", err)
	}

	// Initialize MinIO service
	minioService, err := minioctrl.NewMinioService(
		viper.GetString("minio.endpoint"),
		viper.GetString("minio.access_key"),
		viper.GetString("minio.secret_key"),
		false,
	)
	if err != nil {
		log.Fatalf("Failed to create MinIO service: %v", err)
	}

	// Initialize handlers
	pdfHandler, err := httpHdlr.NewPDFHandler(
		minioService,
		viper.GetString("minio.pdf_bucket"),
		viper.GetString("minio.domain"),
		resourceService,
	)
	if err != nil {
		log.Fatalf("Failed to initialize PDF handler: %v", err)
	}

	// Setup gin router
	r := gin.Default()

	// Initialize conversion handler
	conversionHandler, err := httpHdlr.NewConversionHandler(
		minioService,
		viper.GetString("minio.pdf_bucket"),
		viper.GetString("minio.chunk_bucket"),
		viper.GetString("minio.domain"),
		viper.GetString("unstructured.url"),
		resourceService,
		chunkService,
	)
	if err != nil {
		log.Fatalf("Failed to initialize conversion handler: %v", err)
	}

	// Initialize translation handler
	translationHandler, err := httpHdlr.NewTranslationHandler(jobService)
	if err != nil {
		log.Fatalf("Failed to create MinIO service: %v", err)
	}

	// Initialize Ollama client
	oc := ollama.NewClient(viper.GetString("ollama.url"), &http.Client{
		Timeout: 30 * time.Second,
	})

	// Initialize Weaviate SDK
	wc := weaviateClient.New(weaviateClient.Config{
		Host:   viper.GetString("weaviate.url"),
		Scheme: "http",
	})
	wsdk := weaviate.NewSDK(wc)

	// Initialize knowledge base service and handler
	knowledgeBaseRepo := pgKnowledgeBase.NewRepository(db)
	knowledgeBaseService, err := knowledgebase.NewService(
		knowledgeBaseRepo,
		wsdk,
		oc,
		minioService,
		resourceService,
		chunkService,
	)
	if err != nil {
		log.Fatalf("Failed to create knowledge base service: %v", err)
	}
	knowledgeBaseHandler, err := httpHdlr.NewKnowledgeBaseHandler(knowledgeBaseService)
	if err != nil {
		log.Fatalf("Failed to initialize knowledge base handler: %v", err)
	}

	// Register routes
	r.GET("/pdfs", pdfHandler.List)
	r.POST("/pdfs", pdfHandler.Upload)
	r.POST("/conversion", conversionHandler.Convert)
	r.POST("/translation", translationHandler.Translate)

	// Knowledge base routes
	r.GET("/api/v1/knowledge-bases", knowledgeBaseHandler.ListKnowledgeBases)
	r.GET("/api/v1/knowledge-bases/:id/resources", knowledgeBaseHandler.ListKnowledgeBaseResources)
	r.POST("/api/v1/knowledge-bases/:id/query", knowledgeBaseHandler.QueryKnowledgeBase)
	r.POST("/api/v1/knowledge-bases/:id/resources", knowledgeBaseHandler.AddResourceToKnowledgeBase)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + viper.GetString("server.port"),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Parse shutdown timeout
	timeout, err := time.ParseDuration(viper.GetString("server.shutdown_timeout"))
	if err != nil {
		log.Printf("Invalid shutdown timeout: %v, using default 5s", err)
		timeout = 5 * time.Second
	}

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get underlying *sql.DB: %v", err)
	} else {
		// Close database connection
		sqlDB.Close()
	}

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
