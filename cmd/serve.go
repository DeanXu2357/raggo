/*
Copyright © 2024 Dean
*/
package cmd

import (
	"context"
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

	httpHdlr "raggo/handler/http"
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

	// Initialize PDF handler with MinIO client and config
	pdfHandler, err := httpHdlr.NewPDFHandler(
		minioClient,
		viper.GetString("minio.pdf_bucket"),
		viper.GetString("minio.domain"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize PDF handler: %v", err)
	}

	// Setup gin router
	r := gin.Default()

	// Register routes
	r.POST("/pdfs", pdfHandler.Upload)

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

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func settingDefaultConfig() {
	// Set default values
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.domain", "http://localhost:9000")
	viper.SetDefault("minio.access_key", "minioadmin")
	viper.SetDefault("minio.secret_key", "minioadmin")
	viper.SetDefault("minio.pdf_bucket", "pdfs")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.shutdown_timeout", "5s")

	// Map environment variables to Viper keys
	viper.BindEnv("minio.endpoint", "MINIO_ENDPOINT")
	viper.BindEnv("minio.domain", "MINIO_DOMAIN")
	viper.BindEnv("minio.access_key", "MINIO_ACCESS_KEY")
	viper.BindEnv("minio.secret_key", "MINIO_SECRET_KEY")
	viper.BindEnv("minio.pdf_bucket", "MINIO_PDF_BUCKET")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")

	viper.AutomaticEnv()
}
