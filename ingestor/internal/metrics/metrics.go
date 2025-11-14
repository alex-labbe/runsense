package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// MQTT
	IngestorMessagesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_messages_total",
			Help: "Total number of MQTT messages received",
		},
	)

	IngestorDBInsertsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_db_inserts_total",
			Help: "Total number of database inserts performed into raw_windows",
		},
	)

	IngestorParseErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_parse_errors_total",
			Help: "Total number of errors encountered while parsing messages",
		},
	)

	IngestorDubSkippedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_dup_skipped_total",
			Help: "Total number of duplicate windows skipped during insertion",
		},
	)

	IngestorMqttReconnectsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_mqtt_reconnects_total",
			Help: "Total number of MQTT reconnections",
		},
	)
)

func InitMetricsServer() {
	prometheus.MustRegister(
		IngestorMessagesTotal,
		IngestorDBInsertsTotal,
		IngestorParseErrorsTotal,
		IngestorDubSkippedTotal,
		IngestorMqttReconnectsTotal,
	)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
