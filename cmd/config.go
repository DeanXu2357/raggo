package cmd

import "github.com/spf13/viper"

func settingDefaultConfig() {
	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Map environment variables to Viper keys for PostgreSQL
	viper.BindEnv("postgres.host", "POSTGRES_HOST")
	viper.BindEnv("postgres.port", "POSTGRES_PORT")
	viper.BindEnv("postgres.user", "POSTGRES_USER")
	viper.BindEnv("postgres.password", "POSTGRES_PASSWORD")
	viper.BindEnv("postgres.db", "POSTGRES_DB")

	// Map environment variables to Viper keys for MinIO and Server
	viper.BindEnv("minio.endpoint", "MINIO_ENDPOINT")
	viper.BindEnv("minio.domain", "MINIO_DOMAIN")
	viper.BindEnv("minio.access_key", "MINIO_ACCESS_KEY")
	viper.BindEnv("minio.secret_key", "MINIO_SECRET_KEY")
	viper.BindEnv("minio.pdf_bucket", "MINIO_PDF_BUCKET")
	viper.BindEnv("minio.chunk_bucket", "MINIO_CHUNKS_BUCKET")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")

	// Map environment variables to Viper keys for RabbitMQ
	viper.BindEnv("amqp.url", "AMQP_URL")

	// Set default values for PostgreSQL
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", "5432")
	viper.SetDefault("postgres.user", "postgres")
	viper.SetDefault("postgres.password", "postgres")
	viper.SetDefault("postgres.db", "raggo")

	// Set default values for MinIO and Server
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.domain", "http://localhost:9000")
	viper.SetDefault("minio.access_key", "minioadmin")
	viper.SetDefault("minio.secret_key", "minioadmin")
	viper.SetDefault("minio.pdf_bucket", "pdfs")
	viper.SetDefault("minio.chunk_bucket", "chunks")

	// Set default values for RabbitMQ
	viper.SetDefault("amqp.url", "amqp://guest:guest@localhost:5672/")

	// Set default values for Unstructured API
	viper.BindEnv("unstructured.url", "UNSTRUCTURED_API_URL")
	viper.SetDefault("unstructured.url", "http://unstructured_api:8000")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.shutdown_timeout", "5s")

	viper.BindEnv("weaviate.url", "WEAVIATE_URL")
	viper.SetDefault("weaviate.url", "http://weaviate:8080")

	viper.BindEnv("ollama.url", "OLLAMA_URL")
	viper.SetDefault("ollama.url", "http://ollama:11434/api")
}
