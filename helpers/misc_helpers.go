package helpers

import (
	"regexp"
	"strings"
)

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

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

func GetHTMLInputType(goType string) string {
	switch goType {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return "number"
	case "bool":
		return "checkbox"
	default:
		return "text"
	}
}
