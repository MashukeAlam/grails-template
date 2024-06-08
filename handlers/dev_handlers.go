package handlers

import (
	"github.com/MashukeAlam/grails-template/helpers"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ScaffoldData struct {
	TableName    string          `json:"tableName"`
	RefTableName string          `json:"refTableName"`
	Fields       []helpers.Field `json:"fields"`
}

func GetDevView() fiber.Handler {
	return func(c *fiber.Ctx) error {
		modelNames, err := helpers.GetModelNames()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read model names",
			})
		}
		return c.Render("_dev/_dev_index", fiber.Map{
			"Title":      "Everything Center",
			"ModelNames": modelNames,
		}, "layouts/main")
	}
}

func GetMigration(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		helpers.Migrate(db)
		return c.Render("_dev/_dev_index", fiber.Map{
			"Title": "Everything Center",
		}, "layouts/main")
	}
}

func ProcessIncomingScaffoldData(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data struct {
			ScaffoldData ScaffoldData `json:"scaffoldData"`
		}

		// Parse the JSON request body
		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}

		for _, field := range data.ScaffoldData.Fields {
			fmt.Printf("Field Name: %s, Field Type: %s\n", field.Name, field.Type)
		}

		tableName := data.ScaffoldData.TableName
		refTableName := data.ScaffoldData.RefTableName
		fields := data.ScaffoldData.Fields

		if refTableName != "" {
			helpers.CreateModel(tableName, fields, refTableName)
		} else {
			helpers.CreateModel(tableName, fields)
		}
		return c.JSON(fiber.Map{
			"message":     "Scaffold created successfully",
			"action":      "migrate",
			"actionParam": data.ScaffoldData.TableName,
		})
	}
}
