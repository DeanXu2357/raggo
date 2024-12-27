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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	httpHdlr "raggo/handler/http"
	"raggo/src/chunkctrl"
	"raggo/src/resourcectrl"
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

	log.Printf("Environment variables:")
	log.Printf("POSTGRES_HOST: %s", os.Getenv("POSTGRES_HOST"))
	log.Printf("POSTGRES_PORT: %s", os.Getenv("POSTGRES_PORT"))
	log.Printf("POSTGRES_USER: %s", os.Getenv("POSTGRES_USER"))
	log.Printf("POSTGRES_DB: %s", os.Getenv("POSTGRES_DB"))

	log.Printf("Viper configuration:")
	log.Printf("postgres.host: %s", viper.GetString("postgres.host"))
	log.Printf("postgres.port: %s", viper.GetString("postgres.port"))
	log.Printf("postgres.user: %s", viper.GetString("postgres.user"))
	log.Printf("postgres.db: %s", viper.GetString("postgres.db"))

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
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
		log.Fatalf("Failed to create resource service: %v", err)
	}

	// Initialize MinIO client with config from viper
	minioClient, err := minio.New(viper.GetString("minio.endpoint"), &minio.Options{
		Creds: credentials.NewStaticV4(
			viper.GetString("minio.access_key"),
			viper.GetString("minio.secret_key"),
			"",
		),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	// Initialize handlers
	pdfHandler, err := httpHdlr.NewPDFHandler(
		minioClient,
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
		minioClient,
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

	// Register routes
	r.GET("/pdfs", pdfHandler.List)
	r.POST("/pdfs", pdfHandler.Upload)
	r.POST("/conversion", conversionHandler.Convert)

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
