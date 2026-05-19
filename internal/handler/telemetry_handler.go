package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/RazorBold/golang_backend1/internal/pkg/response"
	"github.com/RazorBold/golang_backend1/internal/service"
	"github.com/gofiber/fiber/v2"
)

type TelemetryHandler struct {
	telemetrySvc *service.TelemetryService
}

func NewTelemetryHandler(telemetrySvc *service.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{telemetrySvc: telemetrySvc}
}

// Ingest godoc
// @Summary      Ingest telemetry from a device
// @Tags         telemetry
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        device_id  path      string          true  "Device ID"
// @Param        body       body      object          true  "Arbitrary JSON telemetry payload"
// @Success      201        {object}  model.IngestResponse
// @Failure      404        {object}  map[string]interface{}
// @Router       /ingest/{device_id} [post]
func (h *TelemetryHandler) Ingest(c *fiber.Ctx) error {
	appID := c.Locals("app_id").(string)
	deviceID := c.Params("device_id")

	body := c.Body()
	if len(body) == 0 {
		return response.BadRequest(c, "request body cannot be empty")
	}

	resp, err := h.telemetrySvc.Ingest(c.Context(), appID, deviceID, body)
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "device not found or does not belong to this application")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.Created(c, resp)
}

// GetData godoc
// @Summary      Query historical telemetry
// @Tags         telemetry
// @Produce      json
// @Security     BearerAuth
// @Param        device_id  path      string  true  "Device ID"
// @Param        from       query     string  false "RFC3339 start time"
// @Param        to         query     string  false "RFC3339 end time"
// @Param        limit      query     int     false "Max records (default 50, max 500)"
// @Success      200        {array}   model.TelemetryResponse
// @Failure      404        {object}  map[string]interface{}
// @Router       /devices/{device_id}/data [get]
func (h *TelemetryHandler) GetData(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	deviceID := c.Params("device_id")

	from, to, limit, err := parseTelemetryQuery(c)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	records, err := h.telemetrySvc.GetData(c.Context(), userID, deviceID, from, to, limit)
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "device not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, records)
}

// GetLatest godoc
// @Summary      Get latest telemetry for a device
// @Tags         telemetry
// @Produce      json
// @Security     BearerAuth
// @Param        device_id  path      string  true  "Device ID"
// @Success      200        {object}  model.TelemetryResponse
// @Failure      404        {object}  map[string]interface{}
// @Router       /devices/{device_id}/latest [get]
func (h *TelemetryHandler) GetLatest(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	deviceID := c.Params("device_id")

	rec, err := h.telemetrySvc.GetLatest(c.Context(), userID, deviceID)
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "no data found for this device")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, rec)
}

func parseTelemetryQuery(c *fiber.Ctx) (from, to *time.Time, limit int, err error) {
	if fromStr := c.Query("from"); fromStr != "" {
		t, parseErr := time.Parse(time.RFC3339, fromStr)
		if parseErr != nil {
			return nil, nil, 0, errors.New("invalid 'from' timestamp, use RFC3339 (e.g. 2026-05-19T00:00:00Z)")
		}
		from = &t
	}
	if toStr := c.Query("to"); toStr != "" {
		t, parseErr := time.Parse(time.RFC3339, toStr)
		if parseErr != nil {
			return nil, nil, 0, errors.New("invalid 'to' timestamp, use RFC3339")
		}
		to = &t
	}

	limit = 50
	if limitStr := c.Query("limit"); limitStr != "" {
		n, parseErr := strconv.Atoi(limitStr)
		if parseErr != nil || n < 1 {
			return nil, nil, 0, errors.New("invalid 'limit' value")
		}
		limit = n
	}
	if limit > 500 {
		limit = 500
	}
	return
}
