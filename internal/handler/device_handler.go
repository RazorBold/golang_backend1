package handler

import (
	"errors"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/pkg/response"
	"github.com/RazorBold/golang_backend1/internal/pkg/validator"
	"github.com/RazorBold/golang_backend1/internal/service"
	"github.com/gofiber/fiber/v2"
)

type DeviceHandler struct {
	deviceSvc *service.DeviceService
}

func NewDeviceHandler(deviceSvc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceSvc: deviceSvc}
}

// Create godoc
// @Summary      Create device
// @Tags         devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id  path      string                    true  "Application ID"
// @Param        body    body      model.CreateDeviceRequest true  "Device payload"
// @Success      201     {object}  model.DeviceResponse
// @Failure      404     {object}  map[string]interface{}
// @Router       /applications/{app_id}/devices [post]
func (h *DeviceHandler) Create(c *fiber.Ctx) error {
	var req model.CreateDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if errs := validator.Validate(req); errs != nil {
		return response.UnprocessableEntity(c, errs)
	}

	userID := c.Locals("user_id").(string)
	device, err := h.deviceSvc.Create(c.Context(), userID, c.Params("app_id"), req)
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "application not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.Created(c, device)
}

// List godoc
// @Summary      List devices for an application
// @Tags         devices
// @Produce      json
// @Security     BearerAuth
// @Param        app_id  path      string  true  "Application ID"
// @Success      200     {array}   model.DeviceResponse
// @Failure      404     {object}  map[string]interface{}
// @Router       /applications/{app_id}/devices [get]
func (h *DeviceHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	devices, err := h.deviceSvc.List(c.Context(), userID, c.Params("app_id"))
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "application not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, devices)
}

// Get godoc
// @Summary      Get device by ID
// @Tags         devices
// @Produce      json
// @Security     BearerAuth
// @Param        app_id  path      string  true  "Application ID"
// @Param        id      path      string  true  "Device ID"
// @Success      200     {object}  model.DeviceResponse
// @Failure      404     {object}  map[string]interface{}
// @Router       /applications/{app_id}/devices/{id} [get]
func (h *DeviceHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	device, err := h.deviceSvc.Get(c.Context(), userID, c.Params("app_id"), c.Params("id"))
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "device not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, device)
}

// Update godoc
// @Summary      Update device
// @Tags         devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        app_id  path      string                   true  "Application ID"
// @Param        id      path      string                   true  "Device ID"
// @Param        body    body      model.UpdateDeviceRequest true "Update payload"
// @Success      200     {object}  model.DeviceResponse
// @Failure      404     {object}  map[string]interface{}
// @Router       /applications/{app_id}/devices/{id} [put]
func (h *DeviceHandler) Update(c *fiber.Ctx) error {
	var req model.UpdateDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if errs := validator.Validate(req); errs != nil {
		return response.UnprocessableEntity(c, errs)
	}

	userID := c.Locals("user_id").(string)
	device, err := h.deviceSvc.Update(c.Context(), userID, c.Params("app_id"), c.Params("id"), req)
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "device not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, device)
}

// Delete godoc
// @Summary      Delete device
// @Tags         devices
// @Produce      json
// @Security     BearerAuth
// @Param        app_id  path      string  true  "Application ID"
// @Param        id      path      string  true  "Device ID"
// @Success      200     {object}  map[string]interface{}
// @Failure      404     {object}  map[string]interface{}
// @Router       /applications/{app_id}/devices/{id} [delete]
func (h *DeviceHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	err := h.deviceSvc.Delete(c.Context(), userID, c.Params("app_id"), c.Params("id"))
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "device not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, fiber.Map{"message": "device deleted"})
}
