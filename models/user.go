package models
// may wanna delete this file because it's only needed for initial setup

import (
    "gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
    gorm.Model
    Name     string `gorm:"size:255;not null" json:"name"`
    Email    string `gorm:"size:255;unique;not null" json:"email"`
    Password string `gorm:"size:255;not null" json:"password"`
}
