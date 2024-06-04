package helpers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"io/ioutil"
)

func CreateModel(tableName string, fields []Field, reference ...string) {
	modelDir := "models"
	if err := os.MkdirAll(modelDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create models directory: %v", err)
	}

	modelName := ToCamelCase(tableName)
	modelContent := generateModelContent(modelName, fields, reference...)
	modelFileName := filepath.Join(modelDir, fmt.Sprintf("%s.go", tableName))
	writeToFile(modelFileName, modelContent)
	fmt.Printf("Model file %s created successfully.\n", modelFileName)

	appendMigrationCode(modelName)

	// TODO: autoMigrate here gorm model.
	fmt.Printf("%s\n\n\ndbGorm.AutoMigrate(&models.%s{})%s\n\n\n", Green, modelName, Reset)

	generateHandlerFile(modelName)

	generateAndWriteViewFiles(tableName, fields)
}

func generateModelContent(modelName string, fields []Field, reference ...string) string {
	var modelBuilder strings.Builder

	modelBuilder.WriteString(fmt.Sprintf("package models\n\nimport \"gorm.io/gorm\"\n\n// %s model\ntype %s struct {\n", modelName, modelName))
	modelBuilder.WriteString("	gorm.Model\n")
	for _, field := range fields {
		fieldName := ToCamelCase(field.Name)
		modelBuilder.WriteString(fmt.Sprintf("	%s %s\n", fieldName, field.Type))
	}
	if len(reference) > 0 {
		referenceTable := reference[0]
		referenceField := ToCamelCase(referenceTable)
		modelBuilder.WriteString(fmt.Sprintf("	%sID int\n", referenceField))
		modelBuilder.WriteString(fmt.Sprintf("	%s %s `gorm:\"foreignKey:%sID;references:ID\"`\n", referenceField, referenceField, referenceField))
	}
	modelBuilder.WriteString("}\n")

	return modelBuilder.String()
}

func writeToFile(filename, content string) {
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write file %s: %v", filename, err)
	}
}

func appendMigrationCode(modelName string) {
	migrationFileName := "internals/migrations.go"
	migrationFunction := `package internals

import (
	"gorm.io/gorm"
	"github.com/MashukeAlam/grails/models"
)

func Migrate(db *gorm.DB) {
`
	migrationCode := fmt.Sprintf("\tdb.AutoMigrate(&models.%s{})\n", modelName)

	if _, err := os.Stat(migrationFileName); os.IsNotExist(err) {
		content := migrationFunction + migrationCode + "}\n"
		writeToFile(migrationFileName, content)
		fmt.Printf("Migration file %s created successfully.\n", migrationFileName)
	} else {
		content, err := os.ReadFile(migrationFileName)
		if err != nil {
			log.Fatalf("Failed to read migration file: %v", err)
		}

		contentStr := string(content)
		contentStr = strings.TrimSuffix(contentStr, "}\n") + "\n" + migrationCode + "}\n"
		writeToFile(migrationFileName, contentStr)
		fmt.Printf("Migration for %s appended to %s successfully.\n", modelName, migrationFileName)
	}
}

func appendRoutesCode(codeToAdd string) error {
	filePath := "internals/routes.go"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.Create(filePath); err != nil {
			return err
		}
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileContent := string(content)
	idx := strings.LastIndex(fileContent, "}")

	newContent := fileContent[:idx] + codeToAdd + "\n}" + fileContent[idx+1:]
	return ioutil.WriteFile(filePath, []byte(newContent), 0644)
}

func generateAndWriteViewFiles(tableName string, fields []Field) {
	viewDirPlural := strings.ToLower(tableName) + "s"
	viewDir := filepath.Join("views", viewDirPlural)

	if err := os.MkdirAll(viewDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create views directory: %v", err)
	}

	indexViewContent := generateIndexViewContent(tableName, fields)
	indexViewFileName := filepath.Join(viewDir, "index.html")
	writeToFile(indexViewFileName, indexViewContent)
	fmt.Printf("Index View file %s created successfully.\n", indexViewFileName)

	insertViewContent := generateInsertViewContent(tableName, fields)
	insertViewFileName := filepath.Join(viewDir, "insert.html")
	writeToFile(insertViewFileName, insertViewContent)
	fmt.Printf("Insert View file %s created successfully.\n", insertViewFileName)
}

func generateIndexViewContent(tableName string, fields []Field) string {
	var tableHeaders, tableRows strings.Builder

	for _, field := range fields {
		tableHeaders.WriteString(fmt.Sprintf("<th>%s</th>", field.Name))
	}

	tableRows.WriteString("{{range .Records}}<tr>")
	for _, field := range fields {
		tableRows.WriteString(fmt.Sprintf("<td>{{.%s}}</td>", ToCamelCase(field.Name)))
	}
	tableRows.WriteString(fmt.Sprintf(`
        <td>
            <a href="%ss/{{.ID}}/edit">Edit</a> |
            <a href="%ss/{{.ID}}/delete">Delete</a>
        </td>
        <td>{{.CreatedAt}}</td>
    </tr>{{end}}`, strings.ToLower(tableName), strings.ToLower(tableName)))

	return fmt.Sprintf(`
    <h2>All %s</h2>
    <a href="/%ss/insert">Add +</a>
    <table>
        <thead>
            <tr>%s<th>Actions</th><th>Created At</th></tr>
        </thead>
        <tbody>%s</tbody>
    </table>
    `, tableName, tableName, tableHeaders.String(), tableRows.String())
}

func generateInsertViewContent(tableName string, fields []Field) string {
	var formFields strings.Builder
	for _, field := range fields {
		formFields.WriteString(fmt.Sprintf(`
            <label for="%s">%s:</label>
            <input type="%s" id="%s" name="%s" required>
        `, field.Name, field.Name, field.Type, field.Name, field.Name))
	}

	return fmt.Sprintf(`
    <h2>Add %s</h2>
    <form action="/%ss" method="POST">
        %s
        <button type="submit">Add %s</button>
    </form>
    `, tableName, tableName, formFields.String(), tableName)
}

func generateHandlerFile(modelName string) {
	const handlerTemplate = `package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"{{.ProjectName}}/models" // Adjust the import path accordingly
)

// Get{{.ModelName}}s retrieves all {{.ModelName}}s from the database
func Get{{.ModelName}}s(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var {{.ModelNamePlural}} []models.{{.ModelName}}
		if result := db.Find(&{{.ModelNamePlural}}); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": result.Error.Error(),
			})
		}
		return c.Render("{{.ModelNameLowercase}}s/index", fiber.Map{
			"Title": "All {{.ModelName}}s",
			"Records": {{.ModelNamePlural}},
		}, "layouts/main")
	}
}

// Insert{{.ModelName}} renders the insert form
func Insert{{.ModelName}}() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("{{.ModelNameLowercase}}s/insert", fiber.Map{
			"Title": "Add New {{.ModelName}}",
		}, "layouts/main")
	}
}

// Create{{.ModelName}} handles the form submission for creating a new {{.ModelName}}
func Create{{.ModelName}}(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		{{.ModelNameLowercase}} := new(models.{{.ModelName}})
		if err := c.BodyParser({{.ModelNameLowercase}}); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		if result := db.Create({{.ModelNameLowercase}}); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": result.Error.Error(),
			})
		}
		return c.Redirect("/{{.ModelNameLowercase}}s")
	}
}

// Edit{{.ModelName}} renders the edit form for a specific {{.ModelName}}
func Edit{{.ModelName}}(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var {{.ModelNameLowercase}} models.{{.ModelName}}
		if err := db.First(&{{.ModelNameLowercase}}, c.Params("id")).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "{{.ModelName}} not found",
			})
		}
		return c.Render("{{.ModelNameLowercase}}s/edit", fiber.Map{"{{.ModelNameLowercase}}": {{.ModelNameLowercase}}, "Title": "Edit Entry"}, "layouts/main")
	}
}

// Update{{.ModelName}} handles the form submission for updating a {{.ModelName}}
func Update{{.ModelName}}(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var {{.ModelNameLowercase}} models.{{.ModelName}}
		if err := db.First(&{{.ModelNameLowercase}}, c.Params("id")).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "{{.ModelName}} not found",
			})
		}
		if err := c.BodyParser(&{{.ModelNameLowercase}}); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		if err := db.Save(&{{.ModelNameLowercase}}).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update {{.ModelName}}",
			})
		}
		return c.Redirect("/{{.ModelNameLowercase}}s")
	}
}

// Delete{{.ModelName}} renders the delete confirmation view for a specific {{.ModelName}}
func Delete{{.ModelName}}(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var {{.ModelNameLowercase}} models.{{.ModelName}}
		if err := db.First(&{{.ModelNameLowercase}}, c.Params("id")).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "{{.ModelName}} not found",
			})
		}
		return c.Render("{{.ModelNameLowercase}}s/delete", fiber.Map{"{{.ModelNameLowercase}}": {{.ModelNameLowercase}}, "Title": "Delete Entry"}, "layouts/main")
	}
}

// Destroy{{.ModelName}} handles the deletion of a {{.ModelName}}
func Destroy{{.ModelName}}(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var {{.ModelNameLowercase}} models.{{.ModelName}}
		if err := db.First(&{{.ModelNameLowercase}}, c.Params("id")).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "{{.ModelName}} not found",
			})
		}
		if err := db.Unscoped().Delete(&{{.ModelNameLowercase}}).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete {{.ModelName}}",
			})
		}
		return c.JSON(fiber.Map{"redirectUrl": "/{{.ModelNameLowercase}}s"})
	}
}
`
	data := struct {
		ModelName          string
		ModelNamePlural    string
		ModelNameLowercase string
		ProjectName        string
	}{
		ModelName:          strings.Title(modelName),
		ModelNamePlural:    strings.Title(modelName) + "s",
		ModelNameLowercase: strings.ToLower(modelName),
		ProjectName:        os.Getenv("PROJECT_NAME"),
	}

	tmpl, err := template.New("handler").Parse(handlerTemplate)
	if err != nil {
		log.Fatalf("Failed to parse handler template: %v", err)
	}

	handlerFileName := filepath.Join("handlers", fmt.Sprintf("%s_handlers.go", strings.ToLower(modelName)))
	file, err := os.Create(handlerFileName)
	if err != nil {
		log.Fatalf("Failed to create handler file %s: %v", handlerFileName, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		log.Fatalf("Failed to execute handler template: %v", err)
	}

	routeRegistration := fmt.Sprintf(`
	// %s routes
	%s := app.Group("/%ss")
	%s.Get("/", handlers.Get%ss(dbGorm))
	%s.Get("/insert", handlers.Insert%s())
	%s.Post("/", handlers.Create%s(dbGorm))
	%s.Get("/:id/edit", handlers.Edit%s(dbGorm))
	%s.Put("/:id", handlers.Update%s(dbGorm))
	%s.Get("/:id/delete", handlers.Delete%s(dbGorm))
	%s.Delete("/:id", handlers.Destroy%s(dbGorm))
`, strings.Title(modelName), modelName, modelName, modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, modelName)

	fmt.Println("\033[33m" + routeRegistration + "\033[0m")

	if err := appendRoutesCode(routeRegistration); err != nil {
		log.Fatalf("Failed to append routes code: %v", err)
	}
	fmt.Printf("Handler file %s created successfully.\n", handlerFileName)
}
