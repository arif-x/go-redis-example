package main

import (
	"fmt"
	"net/http"
	"strconv"

	"example.com/go-crud-redis/cache"

	"github.com/gofiber/fiber/v2"
)

var (
	redisCache = cache.NewRedisCache("localhost:6379", 0, 1)
)

func main() {
	app := fiber.New()

	app.Post("/persons", func(ctx *fiber.Ctx) error {
		var person cache.Person
		if err := ctx.BodyParser(&person); err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		res, err := redisCache.CreatePerson(&person)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(http.StatusOK).JSON(fiber.Map{"person": res})
	})

	app.Get("/persons", func(ctx *fiber.Ctx) error {
		pageReq := ctx.Query("page")
		sizeReq := ctx.Query("size")

		if pageReq == "" {
			pageReq = "1"
		}

		if sizeReq == "" {
			sizeReq = "10"
		}

		page, err := strconv.Atoi(pageReq)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		size, err := strconv.Atoi(sizeReq)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		persons, err := redisCache.GetPersons(page, size)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(http.StatusOK).JSON(fiber.Map{"persons": persons})
	})

	app.Get("/persons/:id", func(ctx *fiber.Ctx) error {
		id := ctx.Params("id")
		person, err := redisCache.GetPerson(id)
		if err != nil {
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"message": "person not found"})
		}
		return ctx.Status(http.StatusOK).JSON(fiber.Map{"person": person})
	})

	app.Put("/persons/:id", func(ctx *fiber.Ctx) error {
		id := ctx.Params("id")
		res, err := redisCache.GetPerson(id)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		var person cache.Person
		if err := ctx.BodyParser(&person); err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		res.Name = person.Name
		res.Age = person.Age
		res, err = redisCache.UpdatePerson(res)

		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(http.StatusOK).JSON(fiber.Map{"person": res})
	})

	app.Delete("/persons/:id", func(ctx *fiber.Ctx) error {
		id := ctx.Params("id")
		err := redisCache.DeletePerson(id)
		if err != nil {
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(http.StatusOK).JSON(fiber.Map{"message": "person deleted successfully"})
	})

	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server is running on port 3000")
}
