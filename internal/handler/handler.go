package handler

import (
	"backend/internal/model"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Login(c *fiber.Ctx) error {
	var user model.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	if user.Name != "john" || user.Pass != "doe" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": "Unauthorized",
		})
	}

	claims := jwt.MapClaims{
		"name":  "John Doe",
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"token": t,
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

func UpFile(c *fiber.Ctx) error {

	filename := "internal/Content/Solution/Problem1/A/input/1.txt"
	lines, err := parseTxtFile(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	file := c.FormValue("file")

	if file == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": "Value empty",
		})
	}

	err = saveStringToFile("./internal/temp/main.go", file)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})

	}

	out, err := execFile("./internal/temp/main.go", lines)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code": fiber.StatusOK,
		"data": fiber.Map{
			"output": out,
		},
	})

}

func GetProblem(c *fiber.Ctx) error {
	problem := c.Params("problem")
	return c.SendFile("./internal/Content/Question/" + problem)
}

func Sandbox(c *fiber.Ctx) error {
	var requestData model.Judge

	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	byteCode, err := hex.DecodeString(requestData.Code)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	buff, err := CreateTar(requestData.Problem, byteCode)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	imgName, err := buildDockerImage(cli, ctx, buff, requestData.User, requestData.Problem)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}
	out, err := runDockerContainer(cli, ctx, imgName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code": fiber.StatusOK,
		"data": out,
	})
}
