package handlers

import "github.com/gofiber/fiber/v2"

func HandleInternalSerErr(c *fiber.Ctx, err error, message string) error {
	return HandleError(c, err, fiber.StatusInternalServerError, message)
}

func HandleOk(c *fiber.Ctx, message string) error {
	return HandleError(c, nil, fiber.StatusOK, message)
}

func HandleUnAuth(c *fiber.Ctx) error {
	return HandleError(c, nil, fiber.StatusUnauthorized, "not authorized")
}

func HandleError(c *fiber.Ctx, err error, status int, message string) error {
	ending := ""
	if err != nil {
		ending = ": " + err.Error()
	}
	return c.Status(status).JSON(fiber.Map{
		"message": message + ending,
	})
}
