package response

import "github.com/gofiber/fiber/v2"

type Success struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type Error struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Errors  any    `json:"errors,omitempty"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Success{Status: "ok", Data: data})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Success{Status: "ok", Data: data})
}

func BadRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(Error{Status: "error", Message: msg})
}

func Unauthorized(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(Error{Status: "error", Message: msg})
}

func Forbidden(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusForbidden).JSON(Error{Status: "error", Message: msg})
}

func NotFound(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(Error{Status: "error", Message: msg})
}

func UnprocessableEntity(c *fiber.Ctx, errs any) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(Error{
		Status:  "error",
		Message: "validation failed",
		Errors:  errs,
	})
}

func InternalError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(Error{
		Status:  "error",
		Message: "internal server error",
	})
}
