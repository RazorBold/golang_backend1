package model

import (
	"encoding/json"
	"time"
)

type Device struct {
	ID            string
	ApplicationID string
	Name          string
	Type          string
	Status        string
	LastSeenAt    *time.Time
	Metadata      json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateDeviceRequest struct {
	Name     string          `json:"name"     validate:"required,min=2,max=100"`
	Type     string          `json:"type"     validate:"omitempty,oneof=sensor actuator gateway"`
	Metadata json.RawMessage `json:"metadata"  swaggertype:"object"`
}

type UpdateDeviceRequest struct {
	Name     string          `json:"name"     validate:"omitempty,min=2,max=100"`
	Type     string          `json:"type"     validate:"omitempty,oneof=sensor actuator gateway"`
	Metadata json.RawMessage `json:"metadata"  swaggertype:"object"`
}

type DeviceResponse struct {
	ID            string          `json:"id"`
	ApplicationID string          `json:"application_id"`
	Name          string          `json:"name"`
	Type          string          `json:"type,omitempty"`
	Status        string          `json:"status"`
	LastSeenAt    *time.Time      `json:"last_seen_at"`
	Metadata      json.RawMessage `json:"metadata,omitempty"  swaggertype:"object"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
