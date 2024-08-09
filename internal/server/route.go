package server

import (
	"backend/internal/handler"
	"backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func (s *FiberServer) RegisterFiberRoutes() {

	s.App.Get("/", s.HelloWorldHandler)
	s.App.Post("/api/auth/login", handler.Login)
	s.App.Get("/homepage", middleware.Checktoken, handler.Accessible)
	s.App.Get("/api/auth/restricted", middleware.Checktoken, handler.Restricted)

}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World",
	}

	return c.JSON(resp)
}
