package server

import (
	"backend/internal/handler"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func (s *FiberServer) RegisterFiberRoutes() {

	s.App.Use(cors.New())
	s.App.Get("/", s.CheckHealth)
	s.App.Post("/api/auth/login", handler.Login)
	s.App.Get("/homepage", middleware.Checktoken, handler.Accessible)
	s.App.Get("/api/auth/restricted", middleware.Checktoken, handler.Restricted)
	s.App.Post("/api/restricted/Upload", middleware.Checktoken, handler.UpFile)
	s.App.Get("/api/restricted/Problems/:problem", handler.GetProblem)
	s.App.Get("/api/restricted/Sandbox", handler.Sandbox)

}

func (s *FiberServer) CheckHealth(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{
		"status": "ok",
	})
}
