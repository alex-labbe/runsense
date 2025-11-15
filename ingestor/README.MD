# Ingestor Overview

This service ingests windowed sensor data from MQTT, stores it in Postgres, and exposes Prometheus-compatible metrics.

## Architecture

### `main.go`
- Reads configuration.
- Initializes dependencies (Postgres DB, MQTT client, metrics).
- Starts all background processes.
- Manages graceful shutdown.

### `internal/config`
- Loads environment variables.
- Provides a `Config` struct consumed by `main.go`.

### `internal/db`
- Opens a Postgres connection.
- Provides helpers like `InsertRawWindow()` and `Ping()`.

### `internal/mqtt`
- Creates and configures the MQTT client.
- Subscribes to relevant topics and receives messages.
- Invokes a callback (defined in `main.go`) whenever a new sensor window payload arrives.

### `internal/metrics`
- Defines counters and gauges.
- Increments metrics from other parts of the code.
- Registers the `/metrics` HTTP handler for Prometheus scraping.

## Example Payload
example MQTT payload representing a window of sensor data:

```json
{
  "device_id": "stm32-xx-xx",
  "seq": "xx",
  "ts_start": "2025-11-12T20:15:00Z",
  "fs_hz": 100.0,
  "duration_s": 2.0,
  "acc": {
    "x": [/* 200 floats */],
    "y": [/* 200 floats */],
    "z": [/* 200 floats */]
  },
  "label": "walk"
}
```



