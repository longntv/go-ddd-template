package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	App           AppConfig
	Database      DatabaseConfig
	AWS           AWSConfig
	S3            S3Config
	SNS           SNSConfig
	SQS           SQSConfig
	Bedrock       BedrockConfig
	Observability ObservabilityConfig
}

// Load loads the configuration from environment variables.
func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			Port:     getEnv("APP_PORT", "8080"),
			LogLevel: getEnv("LOG_LEVEL", "debug"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			Name:            getEnv("DB_NAME", "go_ddd_template"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600),
		},
		AWS: AWSConfig{
			Region:      getEnv("AWS_REGION", "ap-northeast-1"),
			EndpointURL: getEnv("AWS_ENDPOINT_URL", ""),
		},
		S3: S3Config{
			BucketName: getEnv("S3_BUCKET_NAME", "go-ddd-template-bucket"),
			Prefix:     getEnv("S3_BUCKET_PREFIX", "uploads/"),
		},
		SNS: SNSConfig{
			TopicARN: getEnv("SNS_TOPIC_ARN", ""),
		},
		SQS: SQSConfig{
			QueueURL:        getEnv("SQS_QUEUE_URL", ""),
			MaxMessages:     getEnvAsInt("SQS_MAX_MESSAGES", 10),
			WaitTimeSeconds: getEnvAsInt("SQS_WAIT_TIME_SECONDS", 20),
		},
		Bedrock: BedrockConfig{
			Region:  getEnv("BEDROCK_REGION", "ap-northeast-1"),
			ModelID: getEnv("BEDROCK_MODEL_ID", "anthropic.claude-3-5-sonnet-20241022-v2:0"),
		},
		Observability: ObservabilityConfig{
			DataDog: DataDogConfig{
				Enabled:     getEnvAsBool("DD_ENABLED", false),
				ServiceName: getEnv("DD_SERVICE_NAME", "go-ddd-template"),
				AgentHost:   getEnv("DD_AGENT_HOST", "localhost"),
				TracePort:   getEnv("DD_TRACE_AGENT_PORT", "8126"),
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
