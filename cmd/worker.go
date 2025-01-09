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

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"raggo/src/infrastructure/integrations/ollama"
	jobctrl "raggo/src/infrastructure/job"
	"raggo/src/storage/minioctrl"
	"raggo/src/storage/postgres/chunkctrl"
	"raggo/src/storage/postgres/resourcectrl"
	"raggo/src/storage/postgres/translatedchunkctrl"
	"raggo/src/storage/postgres/translatedresourcectrl"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the background job worker",
	RunE:  runWorker,
}

func init() {
	rootCmd.AddCommand(workerCmd)
	settingDefaultConfig()
}

func runWorker(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logger := watermill.NewStdLogger(false, false)

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
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Get underlying *sql.DB for cleanup
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying *sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Initialize AMQP publisher
	amqpPublisher, err := amqp.NewPublisher(
		amqp.NewDurableQueueConfig(viper.GetString("amqp.url")),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return err
	}
	defer amqpPublisher.Close()

	// Initialize AMQP subscriber
	subscriberConfig := amqp.NewDurableQueueConfig(viper.GetString("amqp.url"))
	subscriberConfig.Consume.NoRequeueOnNack = true
	amqpSubscriber, err := amqp.NewSubscriber(
		subscriberConfig,
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return err
	}
	defer amqpSubscriber.Close()

	// Initialize router
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return err
	}

	// Add middleware
	router.AddMiddleware(
		middleware.Recoverer,
		middleware.CorrelationID,
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Second,
			Logger:          logger,
		}.Middleware,
	)

	// Initialize MinioService
	minioService, err := minioctrl.NewMinioService(
		viper.GetString("minio.endpoint"),
		viper.GetString("minio.access_key"),
		viper.GetString("minio.secret_key"),
		viper.GetBool("minio.use_ssl"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize minio service: %v", err)
	}

	// Initialize OllamaClient
	ollamaClient := ollama.NewClient("http://ollama:11434/api", &http.Client{})

	// Initialize ResourceService
	resourceService, err := resourcectrl.NewResourceService(db)
	if err != nil {
		return fmt.Errorf("failed to initialize resource service: %v", err)
	}

	// Initialize ChunkService
	chunkService, err := chunkctrl.NewChunkService(db)
	if err != nil {
		return fmt.Errorf("failed to initialize chunk service: %v", err)
	}

	// Initialize TranslatedResourceService
	translatedResourceService, err := translatedresourcectrl.NewTranslatedResourceService(db)
	if err != nil {
		return fmt.Errorf("failed to initialize translated resource service: %v", err)
	}

	// Initialize TranslatedChunkService
	translatedChunkService, err := translatedchunkctrl.NewTranslatedChunkService(db)
	if err != nil {
		return fmt.Errorf("failed to initialize translated chunk service: %v", err)
	}

	// Initialize TranslationTask
	translationTask := jobctrl.NewTranslationTask(
		resourceService,
		chunkService,
		translatedResourceService,
		translatedChunkService,
		minioService,
		ollamaClient,
	)

	// Initialize job repository and service
	jobRepo := jobctrl.NewPostgresJobRepository(db)
	jobService := jobctrl.NewJobService(amqpPublisher, jobRepo, logger, translationTask)

	// Add handler for processing jobs
	router.AddNoPublisherHandler(
		"job_processor",
		"jobs",
		amqpSubscriber,
		func(msg *message.Message) error {
			return jobService.ProcessJobMessage(msg)
		},
	)

	// Run the router
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := router.Run(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")
	cancel()
	<-router.Running()
	log.Println("Router stopped")

	return nil
}
