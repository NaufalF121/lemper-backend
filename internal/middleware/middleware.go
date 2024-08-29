package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Middleware for validating JWT token
func Checktoken(c *fiber.Ctx) error {
	// Get the JWT token from the request header "Authorization
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": "Unauthorized",
		})
	}

	// Parse the token
	token, err := jwt.Parse(tokenString[7:], func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte("secret"), nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": err.Error(),
			"token":   tokenString,
		})
	}

	// Check if the token is valid
	if !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": "Unauthorized",
		})
	}

	// Extract the token claims
	claims := token.Claims.(jwt.MapClaims)
	c.Locals("user", token.Claims)
	c.Locals("name", claims["name"])

	// Continue stack
	return c.Next()

}
