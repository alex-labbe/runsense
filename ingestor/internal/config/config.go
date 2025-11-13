package config

import (
	"os"
	"strconv"
)

type Config struct {
	// MQTT
	MQTTHost    			string
	MQTTPort    			string
	MQTTUsername 			string
	MQTTPassword 			string
	MQTTTopics    			string
	MQTTClientID			string
	MQTTKeepAlive			int
	MQTTQos					byte
	MQTTCleanSession		bool

	// Postgres
	PGHost					string
	PGPort					int
	PGDB					string
	PGUser					string
	PGPassword				string

	// Tables
	RawTable				string
	FeatTable				string

	// Service
	HTTPPort				int
	LogLevel				string
}

func Load() Config {
	return Config{
		// MQTT
		MQTTHOST: os.Getenv("MQTT_HOST")
		MQTTPORT: os.Getenv("MQTT_PORT")
		MQTTUsername: os.Getenv("MQTT_USERNAME")
		MQTTPassword: os.Getenv("MQTT_PASSWORD")
		MQTTTopics: os.Getenv("MQTT_TOPICS")
		MQTTClientID: os.Getenv("MQTT_CLIENT_ID")
		MQTTKeepAlive: mustInt("MQTT_KEEPALIVE")
		MQTTQos: mustInt("MQQT_QOS")
		MQTTCleanSession: mustBool("MQTT_CLEAN_SESSION")

		// Postgres
		PGHost: os.Getenv("PG_HOST")
		PGPort: os.Getenv("PG_HOST")
		PGDB: os.Getenv("PG_DB")
		PGUser: os.Getenv("PG_USER_ENV")
		PGPassword: os.Getenv("PG_PASSWORD_ENV")

		// Tables
		RawTable: os.Getenv("RAW_TABLE")
		FeatTable: os.Getenv("FEAT_TABLE")

		// Service
		HTTPPort: os.Getenv("HTTP_PORT")
		LogLevel: os.Getenv("LOG_LEVEL")
	}
}

func mustInt(name string) int {
	v := os.Getenv(name)
	n, err := strconv.Atoi(v)
	if err != nil {
		panic("invalid int env var: " + name + "=" + v)
	}
	return n
}

func mustBool(name string) bool {
	v := os.Getenv(name)
	switch v {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		panic("invalid bool env var: " + name + "=" + v)
	}
}