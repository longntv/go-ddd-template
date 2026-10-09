package config

// ObservabilityConfig holds the observability configuration.
type ObservabilityConfig struct {
	DataDog DataDogConfig
}

// DataDogConfig holds the DataDog configuration.
type DataDogConfig struct {
	Enabled     bool
	ServiceName string
	AgentHost   string
	TracePort   string
}
