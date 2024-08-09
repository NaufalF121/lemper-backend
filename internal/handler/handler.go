package handler

import (
	"backend/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func Login(c *fiber.Ctx) error {
	var user model.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	// Throws Unauthorized error
	if user.Name != "john" || user.Pass != "doe" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": "Unauthorized",
		})
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"name":  "John Doe",
		"admin": true,
		"exp":   time.Now().Add(time.Second * 24).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"token":      t,
		"real token": token,
	})
}

func Accessible(c *fiber.Ctx) error {
	return c.SendString("Accessible")
}

func Restricted(c *fiber.Ctx) error {
	claims := c.Locals("user").(jwt.MapClaims)
	name := claims["name"].(string)
	return c.SendString("Welcome " + name)
}