package handler

import (
	"errors"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/pkg/response"
	"github.com/RazorBold/golang_backend1/internal/pkg/validator"
	"github.com/RazorBold/golang_backend1/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ApplicationHandler struct {
	appSvc *service.ApplicationService
}

func NewApplicationHandler(appSvc *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appSvc: appSvc}
}

// Create godoc
// @Summary      Create application
// @Tags         applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateApplicationRequest  true  "Application payload"
// @Success      201   {object}  model.ApplicationResponse
// @Failure      401   {object}  map[string]interface{}
// @Router       /applications [post]
func (h *ApplicationHandler) Create(c *fiber.Ctx) error {
	var req model.CreateApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if errs := validator.Validate(req); errs != nil {
		return response.UnprocessableEntity(c, errs)
	}

	userID := c.Locals("user_id").(string)
	app, err := h.appSvc.Create(c.Context(), userID, req)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Created(c, app)
}

// List godoc
// @Summary      List applications
// @Tags         applications
// @Produce      json
// @Security     BearerAuth
// @Success      200   {array}   model.ApplicationResponse
// @Failure      401   {object}  map[string]interface{}
// @Router       /applications [get]
func (h *ApplicationHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	apps, err := h.appSvc.List(c.Context(), userID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, apps)
}

// Get godoc
// @Summary      Get application by ID
// @Tags         applications
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Application ID"
// @Success      200   {object}  model.ApplicationResponse
// @Failure      404   {object}  map[string]interface{}
// @Router       /applications/{id} [get]
func (h *ApplicationHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	app, err := h.appSvc.Get(c.Context(), userID, c.Params("id"))
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "application not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, app)
}

// Update godoc
// @Summary      Update application
// @Tags         applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                         true  "Application ID"
// @Param        body  body      model.UpdateApplicationRequest true  "Update payload"
// @Success      200   {object}  model.ApplicationResponse
// @Failure      404   {object}  map[string]interface{}
// @Router       /applications/{id} [put]
func (h *ApplicationHandler) Update(c *fiber.Ctx) error {
	var req model.UpdateApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if errs := validator.Validate(req); errs != nil {
		return response.UnprocessableEntity(c, errs)
	}

	userID := c.Locals("user_id").(string)
	app, err := h.appSvc.Update(c.Context(), userID, c.Params("id"), req)
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "application not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, app)
}

// Delete godoc
// @Summary      Delete application
// @Tags         applications
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Application ID"
// @Success      200   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Router       /applications/{id} [delete]
func (h *ApplicationHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	err := h.appSvc.Delete(c.Context(), userID, c.Params("id"))
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "application not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, fiber.Map{"message": "application deleted"})
}

// RegenerateKey godoc
// @Summary      Regenerate API key for an application
// @Tags         applications
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Application ID"
// @Success      200   {object}  model.ApplicationResponse
// @Failure      404   {object}  map[string]interface{}
// @Router       /applications/{id}/regenerate-key [post]
func (h *ApplicationHandler) RegenerateKey(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	app, err := h.appSvc.RegenerateAPIKey(c.Context(), userID, c.Params("id"))
	if errors.Is(err, service.ErrNotFound) {
		return response.NotFound(c, "application not found")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, app)
}
