package server

import (
	"backend/internal/database"
	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "Backend Lemper API",
			AppName:      "Backend Lemper API",
		}),
		db: database.New(),
	}

	return server
}
