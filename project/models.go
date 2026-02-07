package project

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:64;not null"`
	Email string `gorm:"size:64;uniqueIndex;not null"`
	Posts []Post
}

type Post struct {
	ID        uint   `gorm:"primaryKey"`
	Subject   string `gorm:"size:255;not null"`
	Content   string `gorm:"not null"`
	UserID    uint   `gorm:"index:idx_user_created,priority:1"` // FK
	Tags      []Tag  `gorm:"many2many:post_tags"`
	Comments  []Comment
	CreatedAt time.Time `gorm:"autoCreateAt;index:idx_user_created,priority:2"`
}

type Tag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:64;uniqueIndex;not null"`
	Posts []Post `gorm:"many2many:post_tags"`
}

type Comment struct {
	ID        uint           `gorm:"primaryKey"`
	PostID    uint           `gorm:"index"` // FK
	Content   string         `gorm:"not null"`
	UserID    uint           `gorm:"index"` // FK
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
