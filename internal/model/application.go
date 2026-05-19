package model

import "time"

type Application struct {
	ID          string
	UserID      string
	Name        string
	Description string
	APIKey      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateApplicationRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateApplicationRequest struct {
	Name        string `json:"name"        validate:"omitempty,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type ApplicationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	APIKey      string    `json:"api_key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
