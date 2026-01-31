package basis

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:64;not null"`
	Email     string    `gorm:"size:128;uniqueIndex;not null"`
	Age       uint8     `gorm:"not null"`
	Status    string    `gorm:"size:16;default:active;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func CrudTest() {
	dsn := "crud.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to open %s, %v\n", dsn, err)
	}

	if err = db.AutoMigrate(&User{}); err != nil {
		fmt.Println("failed to auto migrate, ", err)
	}

	// clear the users table before crud
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		fmt.Println("failed to clear users table: ", err)
	}

	create(db)
}

func create(db *gorm.DB) {
	users := []User{
		{
			Name:   "Alice",
			Email:  "alice@example.com",
			Age:    25,
			Status: "active",
		},
		{
			Name:   "Bob",
			Email:  "bob@example.com",
			Age:    30,
			Status: "active",
		},
		{
			Name:   "Charlie",
			Email:  "charlie@example.com",
			Age:    28,
			Status: "inactive",
		},
		{
			Name:   "Diana",
			Email:  "diana@example.com",
			Age:    35,
			Status: "active",
		},
		{
			Name:   "Ethan",
			Email:  "ethan@example.com",
			Age:    22,
			Status: "pending",
		},
	}

	if err := db.Create(&users).Error; err != nil {
		fmt.Println("failed to inserts users, ", err)
	} else {
		fmt.Printf("created %d users\n", len(users))
	}

	u := User{
		Name:   "Fiona",
		Email:  "fiona@example.com",
		Age:    40,
		Status: "active",
	}

	if err := db.Create(&u).Error; err != nil {
		fmt.Println("failed to insert user, ", u)
	} else {
		fmt.Printf("new user id: %d\n", u.ID)
	}
}
