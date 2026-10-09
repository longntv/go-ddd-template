package config

// AWSConfig holds the AWS configuration.
type AWSConfig struct {
	Region      string
	EndpointURL string
}

// S3Config holds the S3 configuration.
type S3Config struct {
	BucketName string
	Prefix     string
}

// SNSConfig holds the SNS configuration.
type SNSConfig struct {
	TopicARN string
}

// SQSConfig holds the SQS configuration.
type SQSConfig struct {
	QueueURL        string
	MaxMessages     int32
	WaitTimeSeconds int32
}

// BedrockConfig holds the Bedrock configuration.
type BedrockConfig struct {
	Region  string
	ModelID string
}
