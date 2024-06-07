package helpers

import (
	"github.com/MashukeAlam/grails-template/models"
	"regexp"
	"strings"
)

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var models_list []interface{}

func ToGoType(sqlType string) string {
	// Regular expression to match SQL types with optional length or precision
	re := regexp.MustCompile(`([a-zA-Z]+)(\(\d+\))?`)

	// Extract base type and optional length/precision
	matches := re.FindStringSubmatch(strings.ToUpper(sqlType))
	if len(matches) < 2 {
		return "string"
	}
	baseType := matches[1]

	switch baseType {
	case "VARCHAR", "CHAR", "NVARCHAR", "NCHAR", "CLOB", "TEXT":
		return "string"
	case "INT", "INTEGER", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT":
		return "int"
	case "FLOAT", "DOUBLE", "REAL", "DECIMAL", "NUMERIC":
		return "float64"
	case "DATE", "DATETIME", "TIMESTAMP", "TIME", "YEAR":
		return "time.Time"
	case "BINARY", "VARBINARY", "BLOB", "LONGBLOB", "MEDIUMBLOB", "TINYBLOB":
		return "[]byte"
	case "BOOL", "BOOLEAN":
		return "bool"
	default:
		return "string"
	}
}

func init() {
	models_list = append(models_list, models.User{})
}
