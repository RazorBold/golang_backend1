package broker

import (
	"encoding/json"
	"time"
)

type TelemetryMessage struct {
	EventID    string          `json:"event_id"`
	DeviceID   string          `json:"device_id"`
	AppID      string          `json:"app_id"`
	ReceivedAt time.Time       `json:"received_at"`
	Data       json.RawMessage `json:"data"`
}
