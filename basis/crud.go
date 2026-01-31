package basis

import (
	"errors"
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
	read(db)
	update(db)
}

func create(db *gorm.DB) {
	users := []User{
		{
			Name:   "Alice",
			Email:  "alice@example.com",
			Age:    25,
			Status: "inactive",
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
			Status: "pending",
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

func read(db *gorm.DB) {
	// First: Get the first matching record, order by primary key
	var u User
	if err := db.Where("status = ?", "active").First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("user not found")
		} else {
			fmt.Println("failed to get the first active user, ", err)
		}
	}

	fmt.Println("the first active user: ", u)

	// Take: Take any matching record, order does not matter
	// slightly faster (no sort)
	var u1 User
	if err := db.Take(&u1).Error; err != nil {
		fmt.Println("failed to take a user")
	} else {
		fmt.Println("takes a user, ", u1)
	}

	var u2 User
	if err := db.Where("status = ?", "pending").Take(&u2).Error; err != nil {
		fmt.Println("failed to take a pending user")
	} else {
		fmt.Println("takes a pending user, ", u2)
	}

	// Find: Return all matching records
	var users []User
	if err := db.Where("status = ?", "active").Order("created_at desc").Find(&users).Error; err != nil {
		fmt.Println("failed to find all active users, ", err)
	} else {
		fmt.Println("active users, ", users)
	}

	var all []User
	if err := db.Find(&all).Error; err != nil {
		fmt.Println("failed to find all users, ", err)
	} else {
		fmt.Println("number of users, ", len(all))
	}

	// Scan: Execute a query, and copies result columns into a custom struct
	type UserSummary struct {
		Name   string
		Email  string
		Status string
	}

	var s []UserSummary
	if err := db.Model(&User{}).Select("name", "email", "status").Where("status = ?", "active").Scan(&s).Error; err != nil {
		fmt.Println("failed to scan users, ", err)
	} else {
		fmt.Println("active users: ", s)
	}

	// Count: Return the number of matching records
	var c int64
	if err := db.Model(&User{}).Where("status = ?", "active").Count(&c).Error; err != nil {
		fmt.Println("failed to count active users, ", err)
	} else {
		fmt.Println("the number of active users: ", c)
	}
}

func update(db *gorm.DB) {
	// Update: update a single field
	var u User
	if err := db.Where("email = ?", "bob@example.com").First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("user not found")
		} else {
			fmt.Println("failed to find the user, ", err)
		}
	} else {
		fmt.Println("the user before update: ", u)
		result := db.Model(&u).Update("status", "pending")
		if result.Error != nil {
			fmt.Println("failed to update user, ", result.Error)
		} else {
			fmt.Println("the user after update: ", u)
		}
	}

	// Update multiple records with model struct
	result := db.Model(&User{}).Where("status = ?", "inactive").Update("status", "pending")
	if result.Error != nil {
		fmt.Println("failed to update user, ", result.Error)
	} else {
		fmt.Printf("updated %d rows\n", result.RowsAffected)
	}

	// Updates: update multiple fields but ignore zero values
	var u1 User
	if err := db.Where("email = ?", "fiona@example.com").First(&u1).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("user not found")
		} else {
			fmt.Println("failed to find the user, ", err)
		}
	} else {
		fmt.Println("the user before update: ", u1)
		result = db.Model(&u1).Updates(User{Age: 50, Status: "inactive"})
		if result.Error != nil {
			fmt.Println("failed to update user, ", result.Error)
		} else {
			fmt.Println("the user after update: ", u1)
		}
	}

	// Save: replaces an entire record (including zero values)
	// - If primary key exists → UPDATE
	// - If primary key is zero → INSERT
	// - Updates all fields, including zero values
	u2 := User{
		ID:     1,
		Name:   "",
		Age:    0,
		Status: "inactive",
	}

	result = db.Save(&u2)
	if result.Error != nil {
		fmt.Println("failed to save user, ", result.Error)
	} else if result.RowsAffected == 0 {
		fmt.Println("failed to save user, user not found")
	} else {
		// verify the updates
		var u User
		if err := db.First(&u, 1).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Println("user not found")
			} else {
				fmt.Println("failed to find the user, ", err)
			}
		} else {
			fmt.Println("the saved user: ", u)
		}
	}
}

func delete() {

}
