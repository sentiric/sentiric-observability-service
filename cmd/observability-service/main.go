// sentiric-observability-service/cmd/observability-service/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/sentiric/sentiric-observability-service/internal/config"
	"github.com/sentiric/sentiric-observability-service/internal/logger"
	"github.com/sentiric/sentiric-observability-service/internal/server"

	controlv1 "github.com/sentiric/sentiric-contracts/gen/go/sentiric/control/v1"
)

var (
	ServiceVersion string
	GitCommit      string
	BuildDate      string
)

const serviceName = "observability-service"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Kritik Hata: Konfigürasyon yüklenemedi: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(serviceName, cfg.Env, cfg.LogLevel)

	log.Info().
		Str("version", ServiceVersion).
		Str("commit", GitCommit).
		Str("build_date", BuildDate).
		Str("profile", cfg.Env).
		Msg("🚀 Sentiric Observability Service başlatılıyor...")

	// HTTP ve gRPC sunucularını oluştur
	grpcServer := server.NewGrpcServer(cfg.CertPath, cfg.KeyPath, cfg.CaPath, log)
	httpServer := startHttpServer(cfg.HttpPort, log)

	// gRPC Handler'ı kaydet
	controlv1.RegisterObservabilityServiceServer(grpcServer, &observabilityHandler{})

	// gRPC sunucusunu bir goroutine'de başlat
	go func() {
		log.Info().Str("port", cfg.GRPCPort).Msg("gRPC sunucusu dinleniyor...")
		if err := server.Start(grpcServer, cfg.GRPCPort); err != nil && err.Error() != "http: Server closed" {
			log.Error().Err(err).Msg("gRPC sunucusu başlatılamadı")
		}
	}()

	// Graceful shutdown için sinyal dinleyicisi
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Warn().Msg("Kapatma sinyali alındı, servisler durduruluyor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.Stop(grpcServer)
	log.Info().Msg("gRPC sunucusu durduruldu.")

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP sunucusu düzgün kapatılamadı.")
	} else {
		log.Info().Msg("HTTP sunucusu durduruldu.")
	}

	log.Info().Msg("Servis başarıyla durduruldu.")
}

func startHttpServer(port string, log zerolog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok"}`)
	})

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		log.Info().Str("port", port).Msg("HTTP sunucusu (health) dinleniyor")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP sunucusu başlatılamadı")
		}
	}()
	return srv
}

// =================================================================
// GRPC HANDLER IMPLEMENTASYONU (Placeholder)
// =================================================================

type observabilityHandler struct {
	controlv1.UnimplementedObservabilityServiceServer
}

func (*observabilityHandler) GetMetrics(ctx context.Context, req *controlv1.GetMetricsRequest) (*controlv1.GetMetricsResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("rpc", "GetMetrics").Str("service_name", req.GetServiceName()).Logger()
	log.Info().Msg("GetMetrics isteği alındı (Placeholder)")

	// Simüle edilmiş metrikler
	metrics := map[string]float64{
		"cpu_usage":       0.5,
		"active_requests": 15.0,
	}
	return &controlv1.GetMetricsResponse{
		Metrics: metrics,
	}, nil
}

func (*observabilityHandler) GetLogs(ctx context.Context, req *controlv1.GetLogsRequest) (*controlv1.GetLogsResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("rpc", "GetLogs").Str("service_name", req.GetServiceName()).Logger()
	log.Info().Int32("limit", req.GetLimit()).Msg("GetLogs isteği alındı (Placeholder)")

	// Simüle edilmiş log girdileri
	logs := []string{
		"2025-10-06 INFO: Service started successfully.",
		fmt.Sprintf("2025-10-06 WARN: Low disk space on %s.", req.GetServiceName()),
	}
	return &controlv1.GetLogsResponse{
		LogEntries: logs,
	}, nil
}
