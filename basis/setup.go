package basis

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setup(dsn string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to open %s, %v\n", dsn, err))
	}

	if err = db.AutoMigrate(&User{}); err != nil {
		panic(fmt.Sprintf("failed to auto migrate, %v\n", err))
	}

	now := time.Now()
	users := []User{
		{
			Name:        "Alice",
			Email:       "alice@example.com",
			Phone:       "3239085547",
			Age:         25,
			Status:      "inactive",
			LastLoginAt: now.AddDate(0, 0, -10).Add(-5 * time.Hour).Add(-49 * time.Minute).Add(-10 * time.Second),
		},
		{
			Name:        "Bob",
			Email:       "bob@gmail.com",
			Phone:       "4239085657",
			Age:         30,
			Status:      "active",
			LastLoginAt: now.AddDate(0, 0, -20).Add(-6 * time.Hour).Add(-9 * time.Minute).Add(-34 * time.Second),
		},
		{
			Name:        "Charlie",
			Email:       "charlie@gmail.com",
			Phone:       "9099085547",
			Age:         28,
			Status:      "pending",
			LastLoginAt: now.AddDate(0, 0, -1).Add(-19 * time.Hour).Add(-9 * time.Minute).Add(-1 * time.Second),
		},
		{
			Name:        "Diana",
			Email:       "diana@example.com",
			Phone:       "6230085547",
			Age:         35,
			Status:      "active",
			LastLoginAt: now.AddDate(0, 0, -7).Add(-3 * time.Hour).Add(-19 * time.Minute).Add(-56 * time.Second),
		},
		{
			Name:        "Ethan",
			Email:       "ethan@example.com",
			Phone:       "2134905547",
			Age:         22,
			Status:      "pending",
			LastLoginAt: now.AddDate(0, 0, -7).Add(-51 * time.Hour).Add(-8 * time.Minute).Add(-15 * time.Second),
		},
		{
			Name:        "Fiona",
			Email:       "fiona@gmail.com",
			Phone:       "2134985547",
			Age:         40,
			Status:      "active",
			LastLoginAt: now,
		},
	}

	if err := db.Create(&users).Error; err != nil {
		panic(fmt.Sprintf("failed to inserts users, %v\n", err))
	} else {
		fmt.Printf("created %d users\n", len(users))
	}

	return db
}
