package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alex-labbe/runsense/ingestor/internal/config"
	"github.com/alex-labbe/runsense/ingestor/internal/db"
	"github.com/alex-labbe/runsense/ingestor/internal/metrics"
	"github.com/alex-labbe/runsense/ingestor/internal/mqtt"
)

func parseWindow(payload []byte) (db.RawWindow, error) {

	return db.RawWindow{}, errors.New("not implemented")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Load configuration from environment variables
	cfg := config.Load()

	log.Printf("starting ingestor service on port %d", cfg.MQTTPort)

	// Initialize metrics server
	metrics.InitMetricsServer()

	// Initialize database connection
	store, err := db.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer store.Close()

	// Define MQTT message handler
	// topic is mqtt_topic
	handler := func(topic string, payload []byte) {

		msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// Parse and validate the incoming message
		w, err := parseWindow(payload)
		if err != nil {
			metrics.IngestorParseErrorsTotal.Inc()
			log.Printf("failed to parse message on topic %s: %v", topic, err)
			return
		}

		// Insert the raw window into the database
		_, err = store.InsertRawWindow(msgCtx, w)
		if err != nil {
			if errors.Is(err, db.ErrDuplicateWindow) {
				// just a duplicate window
				metrics.IngestorDubSkippedTotal.Inc()
				return
			}
			// real db error
			log.Printf("db insert error: %v", err)
		}

		metrics.IngestorDBInsertsTotal.Inc()
		metrics.IngestorMessagesTotal.Inc()
	}

	// init MQTT client
	mq := mqtt.New(cfg, handler)
	if err := mq.Start(); err != nil {
		log.Fatalf("failed to start mqtt client: %v", err)
	}
	// no defer here

	// HTTP server /health and /metrics
	mux := http.NewServeMux()
	// health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		hctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// ping db
		if err := store.Ping(hctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("db not healthy\n"))
			return
		}

		if !mq.Healthy() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("mqtt not healthy"))
			return
		}

		// if we reach here, both are healthy
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.Handle("/metrics", metrics.Handler())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: mux,
	}

	// start HTTP server in background
	go func() {
		log.Printf("listening on port: %d", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http server error: %v", err)
			cancel()
		}
	}()

	// shutdown on SIGINT/SIGTERM

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

}
