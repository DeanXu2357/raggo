package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"raggo/src/jobctrl"
)

var enqueueExampleCmd = &cobra.Command{
	Use:   "enqueue-example",
	Short: "Enqueue an example job",
	RunE:  runEnqueueExample,
}

func init() {
	rootCmd.AddCommand(enqueueExampleCmd)
	settingDefaultConfig()
}

func runEnqueueExample(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logger := watermill.NewStdLogger(false, false)

	// Log configuration values
	log.Printf("PostgreSQL Configuration:")
	log.Printf("host: %s", viper.GetString("postgres.host"))
	log.Printf("port: %s", viper.GetString("postgres.port"))
	log.Printf("user: %s", viper.GetString("postgres.user"))
	log.Printf("db: %s", viper.GetString("postgres.db"))

	log.Printf("\nAMQP Configuration:")
	log.Printf("url: %s", viper.GetString("amqp.url"))

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
	publisher, err := amqp.NewPublisher(
		amqp.NewDurableQueueConfig(viper.GetString("amqp.url")),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}
	defer publisher.Close()

	// Initialize job repository and service
	jobRepo := jobctrl.NewPostgresJobRepository(db)
	jobService := jobctrl.NewJobService(publisher, jobRepo, logger, nil)

	// Create test payload
	payload := jobctrl.TestPayload{
		Print: "hello world",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Enqueue job
	ctx := context.Background()
	job, err := jobService.EnqueueJob(ctx, "test", payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	fmt.Printf("Successfully enqueued job with ID: %d\n", job.ID)
	return nil
}
