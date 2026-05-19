package handler

import (
	"errors"
	"time"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/pkg/response"
	"github.com/RazorBold/golang_backend1/internal/pkg/validator"
	"github.com/RazorBold/golang_backend1/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.RegisterRequest  true  "Registration payload"
// @Success      201   {object}  model.UserResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      422   {object}  map[string]interface{}
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if errs := validator.Validate(req); errs != nil {
		return response.UnprocessableEntity(c, errs)
	}

	user, err := h.authSvc.Register(c.Context(), req)
	if errors.Is(err, service.ErrEmailTaken) {
		return response.BadRequest(c, "email already registered")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.Created(c, user)
}

// Login godoc
// @Summary      Login and obtain tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.LoginRequest  true  "Login credentials"
// @Success      200   {object}  model.AuthResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if errs := validator.Validate(req); errs != nil {
		return response.UnprocessableEntity(c, errs)
	}

	authResp, err := h.authSvc.Login(c.Context(), req)
	if errors.Is(err, service.ErrInvalidCreds) {
		return response.Unauthorized(c, "invalid email or password")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, authResp)
}

// Refresh godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.RefreshRequest  true  "Refresh token"
// @Success      200   {object}  model.AuthResponse
// @Failure      401   {object}  map[string]interface{}
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if errs := validator.Validate(req); errs != nil {
		return response.UnprocessableEntity(c, errs)
	}

	authResp, err := h.authSvc.Refresh(c.Context(), req.RefreshToken)
	if errors.Is(err, service.ErrInvalidToken) {
		return response.Unauthorized(c, "invalid or expired refresh token")
	}
	if err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, authResp)
}

// Logout godoc
// @Summary      Logout (blacklist access token)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.LogoutRequest  true  "Refresh token to revoke"
// @Success      200   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req model.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if errs := validator.Validate(req); errs != nil {
		return response.UnprocessableEntity(c, errs)
	}

	jti := c.Locals("jti").(string)
	expiresAt := c.Locals("expires_at").(time.Time)

	if err := h.authSvc.Logout(c.Context(), req.RefreshToken, jti, expiresAt); err != nil {
		return response.InternalError(c)
	}
	return response.OK(c, fiber.Map{"message": "logged out successfully"})
}
