package helpers

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

	fmt.Printf("%s%sUPDATING%s\tmigrations.go\n", Bold, Yellow, Reset)
	appendMigrationCode(modelName)
	fmt.Printf("%s%sGENERATING%s\thandlers\n", Bold, Yellow, Reset)
	generateHandlerFile(modelName)
	generateAndWriteViewFiles(tableName, fields)
	appendModelToJSON(modelName, fields)
}

func appendModelToJSON(modelName string, fields []Field) {
	models, err := ReadModelsFromJSON()
	if err != nil {
		log.Fatalf("Failed to read models from JSON: %v", err)
	}

	var modelFields []Field
	for _, field := range fields {
		modelFields = append(modelFields, Field{
			Name: field.Name,
			Type: field.Type,
		})
	}

	models[modelName] = modelFields

	err = WriteModelsToJSON(models)
	if err != nil {
		log.Fatalf("Failed to write models to JSON: %v", err)
	}
	fmt.Printf("%s%sUPDATE%s\tmodels.json\t\n", Bold, Yellow, Reset)
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

	fmt.Printf("%s%sUPDATE%s\tmodels.go\t\n", Bold, Yellow, Reset)

	return modelBuilder.String()
}

func writeToFile(filename, content string) {
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write file %s: %v", filename, err)
	}
}

func appendMigrationCode(modelName string) {
	migrationFileName := "helpers/migrations.go"
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
	} else {
		content, err := os.ReadFile(migrationFileName)
		if err != nil {
			log.Fatalf("Failed to read migration file: %v", err)
		}

		contentStr := string(content)
		contentStr = strings.TrimSuffix(contentStr, "}\n") + "\n" + migrationCode + "}\n"
		writeToFile(migrationFileName, contentStr)
		fmt.Printf("%s%sUPDATE%s\tmigrations.go\t\n", Bold, Yellow, Reset)
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
	fmt.Printf("%s%sUPDATED%s\troutes.go\n", Bold, Green, Reset)
	return ioutil.WriteFile(filePath, []byte(newContent), 0644)
}

func generateAndWriteViewFiles(tableName string, fields []Field) {
	viewDirPlural := strings.ToLower(tableName) + "s"
	viewDir := filepath.Join("views", viewDirPlural)

	if err := os.MkdirAll(viewDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create views directory: %v", err)
	}
	fmt.Printf("%s%sTRYING%s\tviews\t\n", Bold, Yellow, Reset)

	indexViewContent := generateIndexViewContent(tableName, fields)
	indexViewFileName := filepath.Join(viewDir, "index.html")
	writeToFile(indexViewFileName, indexViewContent)

	insertViewContent := generateInsertViewContent(tableName, fields)
	insertViewFileName := filepath.Join(viewDir, "insert.html")
	writeToFile(insertViewFileName, insertViewContent)

	showViewContent := generateShowViewContent(tableName, fields)
	showViewFileName := filepath.Join(viewDir, "show.html")
	writeToFile(showViewFileName, showViewContent)

	editViewContent := generateEditViewContent(tableName, fields)
	editViewFileName := filepath.Join(viewDir, "edit.html")
	writeToFile(editViewFileName, editViewContent)

	deleteViewContent := generateDeleteViewContent(tableName, fields)
	deleteViewFileName := filepath.Join(viewDir, "delete.html")
	writeToFile(deleteViewFileName, deleteViewContent)
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

	fmt.Printf("%s%sGENERATED%s\tindex.html\n", Bold, Green, Reset)

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
        `, field.Name, field.Name, GetHTMLInputType(field.Type), field.Name, field.Name))
	}

	fmt.Printf("%s%sGENERATED%s\tinsert.html\n", Bold, Green, Reset)

	return fmt.Sprintf(`
    <h2>Add %s</h2>
    <form action="/%ss" method="POST">
        %s
        <button type="submit">Add %s</button>
    </form>
    `, tableName, tableName, formFields.String(), tableName)
}

func generateShowViewContent(tableName string, fields []Field) string {
	var tableRows strings.Builder

	for _, field := range fields {
		tableRows.WriteString(fmt.Sprintf("<tr><th>%s</th><td>{{.%s}}</td></tr>", field.Name, ToCamelCase(field.Name)))
	}

	fmt.Printf("%s%sGENERATED%s\tshow.html\n", Bold, Green, Reset)

	return fmt.Sprintf(`
    <h2>Show %s</h2>
    <table>
        <tbody>%s</tbody>
    </table>
    <a href="/%ss">Back</a>
    `, tableName, tableRows.String(), tableName)
}

func generateEditViewContent(tableName string, fields []Field) string {
	var formFields strings.Builder
	for _, field := range fields {
		formFields.WriteString(fmt.Sprintf(`
            <label for="%s">%s:</label>
            <input type="%s" id="%s" name="%s" value="{{.%s.%s}}" required>
        `, field.Name, field.Name, GetHTMLInputType(field.Type), field.Name, field.Name, strings.ToLower(tableName), ToCamelCase(field.Name)))
	}

	fmt.Printf("%s%sGENERATED%s\tedit.html\n", Bold, Green, Reset)

	return fmt.Sprintf(`
    <h2>Edit %s</h2>
    <form id="editForm">
        %s
        <button type="submit">Update %s</button>
    </form>

    <script>
        document.getElementById('editForm').addEventListener('submit', async function(event) {
            event.preventDefault();
            const form = event.target;
            const data = new FormData(form);
            const jsonData = {};
			
			Array.from(form.elements).forEach(input => {
                if (input.name) {
                    if (input.type === 'number') {
                        jsonData[input.name] = parseInt(input.value, 10);
                    } else {
                        jsonData[input.name] = input.value;
                    }
                }
            });

            try {
                const response = await fetch('/%ss/{{.%s.ID}}', {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(jsonData)
                });

                if (response.ok) {
                    alert('Update successful!');
                    window.location.href = '/%ss';
                } else {
                    const errorData = await response.json();
                    alert('Error: ' + errorData.error);
                }
            } catch (error) {
                console.error('Error:', error);
                alert('An error occurred while updating.');
            }
        });
    </script>
    `, tableName, formFields.String(), tableName, tableName, strings.ToLower(tableName), tableName)
}

func generateDeleteViewContent(tableName string, fields []Field) string {
	var tableRows strings.Builder

	for _, field := range fields {
		tableRows.WriteString(fmt.Sprintf("<tr><th>%s</th><td>{{.%s}}</td></tr>", field.Name, ToCamelCase(field.Name)))
	}

	fmt.Printf("%s%sGENERATED%s\tdelete.html\n", Bold, Green, Reset)

	return fmt.Sprintf(`
    <h2>Delete %s</h2>
    <table>
        <tbody>%s</tbody>
    </table>
    <form id="deleteForm">
        <button type="submit">Delete</button>
    </form>
    <a href="/%ss">Back</a>

    <script>
        document.getElementById('deleteForm').addEventListener('submit', async function(event) {
            event.preventDefault();

            if (!confirm('Are you sure you want to delete this?')) {
                return;
            }

            try {
                const response = await fetch('/%ss/{{.%s.ID}}', {
                    method: 'DELETE',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                });

                if (response.ok) {
                    alert('Delete successful!');
                    window.location.href = '/%ss';
                } else {
                    const errorData = await response.json();
                    alert('Error: ' + errorData.error);
                }
            } catch (error) {
                console.error('Error:', error);
                alert('An error occurred while deleting.');
            }
        });
    </script>
    `, tableName, tableRows.String(), tableName, tableName, strings.ToLower(tableName), tableName)
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

// Show{{.ModelName}} renders the details view for a specific {{.ModelName}}
func Show{{.ModelName}}(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var {{.ModelNameLowercase}} models.{{.ModelName}}
		if err := db.First(&{{.ModelNameLowercase}}, c.Params("id")).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "{{.ModelName}} not found",
			})
		}
		return c.Render("{{.ModelNameLowercase}}s/show", fiber.Map{"{{.ModelNameLowercase}}": {{.ModelNameLowercase}}, "Title": "Show Entry"}, "layouts/main")
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
		return c.JSON(fiber.Map{"redirectUrl": "/{{.ModelNameLowercase}}s"})
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
	%s.Get("/:id", handlers.Show%s(dbGorm))
	%s.Get("/:id/edit", handlers.Edit%s(dbGorm))
	%s.Put("/:id", handlers.Update%s(dbGorm))
	%s.Get("/:id/delete", handlers.Delete%s(dbGorm))
	%s.Delete("/:id", handlers.Destroy%s(dbGorm))
`, strings.Title(modelName), modelName, modelName, modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, strings.Title(modelName), modelName, modelName)

	if err := appendRoutesCode(routeRegistration); err != nil {
		log.Fatalf("Failed to append routes code: %v", err)
	}
	fmt.Printf("%s%sGENERATED%s\t%shandlers.go\n", Bold, Green, Reset, modelName)
}
