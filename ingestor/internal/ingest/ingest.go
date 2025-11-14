package ingest

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/alex-labbe/runsense/ingestor/internal/db"
)

// AccBlock holds accelerometer samples for each axis.

type AccBlock struct {
	X []float64 `json:"x"`
	Y []float64 `json:"y"`
	Z []float64 `json:"z"`
}

// IMUPayload represents a single window of IMU data coming from devices.

type IMUPayload struct {
	DeviceID  string   `json:"device_id"`
	Seq       int64    `json:"seq"`
	TsStart   string   `json:"ts_start"` // ISO8601 string
	FsHz      float64  `json:"fs_hz"`
	DurationS float64  `json:"duration_s"`
	Label     *string  `json:"label,omitempty"`
	Acc       AccBlock `json:"acc"`
}

const (
	minSamples = 64
)

func ParseWindow(payload []byte) (db.RawWindow, error) {
	var p IMUPayload

	if err := json.Unmarshal(payload, &p); err != nil {
		return db.RawWindow{}, fmt.Errorf("unmarshal payload: %w", err)
	}

	// basic validation
	if p.DeviceID == "" {
		return db.RawWindow{}, fmt.Errorf("device_id is required")
	}
	if p.Seq < 0 {
		return db.RawWindow{}, fmt.Errorf("seq must be non-negative")
	}
	if p.FsHz <= 0 {
		return db.RawWindow{}, fmt.Errorf("fs_hz must be positive")
	}
	if p.DurationS <= 0 {
		return db.RawWindow{}, fmt.Errorf("duration_s must be positive")
	}

	// Acc array validation
	nx := len(p.Acc.X)
	ny := len(p.Acc.Y)
	nz := len(p.Acc.Z)

	if nx < minSamples || ny < minSamples || nz < minSamples {
		return db.RawWindow{}, fmt.Errorf("not enough samples: x=%d y=%d z=%d (min %d)", nx, ny, nz, minSamples)
	}

	// parse ts_start
	ts, err := time.Parse(time.RFC3339Nano, p.TsStart)
	if err != nil {
		return db.RawWindow{}, fmt.Errorf("invalid ts_start: %w", err)
	}

	// build db.RawWindow (validated and typed) keep payload as-is
	return db.RawWindow{
		DeviceID:  p.DeviceID,
		Seq:       p.Seq,
		TsStart:   ts,
		FsHz:      int(math.Round(p.FsHz)),
		DurationS: float32(p.DurationS),
		Label:     p.Label,
		Payload:   json.RawMessage(payload),
	}, nil
}
