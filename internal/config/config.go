// sentiric-observability-service/internal/config/config.go
package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	GRPCPort     string
	HttpPort     string
	CertPath     string
	KeyPath      string
	CaPath       string
	LogLevel     string
	Env          string
    
    // Observability bağımlılıkları (Placeholder)
    PrometheusURL string
    LokiURL string
}

func Load() (*Config, error) {
	godotenv.Load()

	// Harmonik Mimari Portlar (Control Plane, 110XX bloğu atandı)
	return &Config{
		GRPCPort:     GetEnv("OBSERVABILITY_SERVICE_GRPC_PORT", "11011"),
		HttpPort:     GetEnv("OBSERVABILITY_SERVICE_HTTP_PORT", "11010"),
		
		CertPath:     GetEnvOrFail("OBSERVABILITY_SERVICE_CERT_PATH"),
		KeyPath:      GetEnvOrFail("OBSERVABILITY_SERVICE_KEY_PATH"),
		CaPath:       GetEnvOrFail("GRPC_TLS_CA_PATH"),
		LogLevel:     GetEnv("LOG_LEVEL", "info"),
		Env:          GetEnv("ENV", "production"),

        PrometheusURL: GetEnv("PROMETHEUS_URL", "http://prometheus:9090"),
        LokiURL: GetEnv("LOKI_URL", "http://loki:3100"),
	}, nil
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func GetEnvOrFail(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal().Str("variable", key).Msg("Gerekli ortam değişkeni tanımlı değil")
	}
	return value
}