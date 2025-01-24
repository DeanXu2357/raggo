package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	weaviateClient "github.com/weaviate/weaviate-go-client/v4/weaviate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	v2 "raggo/handler/http/v2"
	"raggo/src/core/knowledgebase"
	"raggo/src/core/knowledgebase/valkey"
	"raggo/src/fsutil"
	"raggo/src/infrastructure/integrations/ollama"
	"raggo/src/infrastructure/log"
	"raggo/src/storage/weaviate"
)

// serveV2Cmd represents the serveV2 command
var serveV2Cmd = &cobra.Command{
	Use:   "serveV2",
	Short: "Run a rag server with v2 APIs",
	Long:  `The serveV2 command starts an HTTP server that provides the rag service with v2 APIs`,
	Run:   RunServerV2,
}

func init() {
	rootCmd.AddCommand(serveV2Cmd)
}

func RunServerV2(cmd *cobra.Command, args []string) {
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
		log.Error(err, "Failed to connect to database")
		return
	}

	// Initialize Ollama client
	oc := ollama.NewClient(viper.GetString("ollama.url"), &http.Client{
		Timeout: 30 * time.Second,
	})

	// Initialize Weaviate client
	wc := weaviateClient.New(weaviateClient.Config{
		Host:   viper.GetString("weaviate.url"),
		Scheme: "http",
	})
	wsdk := weaviate.NewSDK(wc)

	// Initialize Valkey store with connection string
	valkeyConn := fmt.Sprintf("valkey://%s:%d",
		viper.GetString("valkey.host"),
		viper.GetInt("valkey.port"))
	valkeyStore, err := valkey.NewValkeyStore(valkeyConn)
	if err != nil {
		log.Error(err, "Failed to create valkey store")
		return
	}

	// Initialize file store
	fs := fsutil.NewLocalFileStore()

	// Initialize local knowledge base with all dependencies
	kb, err := knowledgebase.NewV2Service(
		viper.GetString("rag.data_root"),
		valkeyStore,
		wsdk,
		oc,
	)
	if err != nil {
		log.Error(err, "Failed to create knowledge base service")
		return
	}

	// Initialize HTTP handler with individual services
	handler := v2.NewHandler(
		kb, // KnowledgeBaseService
		knowledgebase.NewResourceService(viper.GetString("rag.data_root"), fs),
		knowledgebase.NewSearchService(wsdk, oc),
		knowledgebase.NewChatService(kb),
		knowledgebase.NewSystemService(kb),
	)

	// Setup gin router
	r := gin.Default()

	// Register routes
	handler.RegisterRoutes(r)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + viper.GetString("server.port"),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err, "Failed to start server")
			return
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Parse shutdown timeout
	timeout, err := time.ParseDuration(viper.GetString("server.shutdown_timeout"))
	if err != nil {
		log.Error(err, "Invalid shutdown timeout, using default 5s")
		timeout = 5 * time.Second
	}

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Error(err, "Failed to get underlying *sql.DB")
	} else {
		// Close database connection
		if err := sqlDB.Close(); err != nil {
			log.Error(err, "Error closing database connection")
		}
	}

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Error(err, "Server forced to shutdown")
	}

	log.Info("Server exited")
}
