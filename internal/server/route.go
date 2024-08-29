package server

import (
	"backend/internal/handler"
	"backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func (s *FiberServer) RegisterFiberRoutes() {

	s.App.Use(cors.New())
	s.App.Get("/", s.HelloWorldHandler)
	s.App.Post("/api/auth/login", handler.Login)
	s.App.Get("/homepage", middleware.Checktoken, handler.Accessible)
	s.App.Get("/api/auth/restricted", middleware.Checktoken, handler.Restricted)
	s.App.Post("/api/restricted/Upload", middleware.Checktoken, handler.UpFile)

}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World",
	}

	return c.JSON(resp)
}
