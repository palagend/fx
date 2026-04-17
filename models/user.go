package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email     string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type RefreshToken struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;size:500;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &RefreshToken{}, &Trade{}, &Holding{}, &Investment{})
}
