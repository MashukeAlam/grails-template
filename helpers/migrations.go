// Package helpers Never TOUCH this file please.
package helpers

import (
	"gorm.io/gorm"
	"github.com/MashukeAlam/grails-template/models"
)

func Migrate(db *gorm.DB) {
	db.AutoMigrate(models.User{})
}
