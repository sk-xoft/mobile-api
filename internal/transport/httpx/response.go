package httpx

import (
	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/constants"
)

const SuccessCode = constants.CodeSuccess

type Response struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Data        any    `json:"data,omitempty"`
}

func Success(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Code:        SuccessCode,
		Description: successDescription(c.Get("Accept-Language")),
		Data:        data,
	})
}

func Error(c *fiber.Ctx, status int, code, description string) error {
	return c.Status(status).JSON(Response{Code: code, Description: description})
}

func successDescription(language string) string {
	if language == "en" || language == "EN" {
		return "Success"
	}
	return "สำเร็จ"
}
