package model

import (
	"encoding/json"
	"time"
)

type TelemetryRecord struct {
	ID         int64
	DeviceID   string
	Payload    json.RawMessage
	ReceivedAt time.Time
}

type TelemetryResponse struct {
	ID         int64           `json:"id"`
	DeviceID   string          `json:"device_id"`
	Payload    json.RawMessage `json:"payload"  swaggertype:"object"`
	ReceivedAt time.Time       `json:"received_at"`
}

type IngestResponse struct {
	Status     string    `json:"status"`
	ReceivedAt time.Time `json:"received_at"`
}
