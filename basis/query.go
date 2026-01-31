package basis

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func QueryTest() {
	db := setup()
	scopedTest(db)
}

func setup() *gorm.DB {
	dsn := "query.db"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to open %s, %v\n", dsn, err))
	}

	if err = db.AutoMigrate(&User{}); err != nil {
		panic(fmt.Sprintf("failed to auto migrate, %v\n", err))
	}

	// clear the users table before crud
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		panic(fmt.Sprintf("failed to clear users table, %v\n", err))
	}

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
		{
			Name:   "Fiona",
			Email:  "fiona@example.com",
			Age:    40,
			Status: "active",
		},
	}

	if err := db.Create(&users).Error; err != nil {
		panic(fmt.Sprintf("failed to inserts users, %v\n", err))
	} else {
		fmt.Printf("created %d users\n", len(users))
	}

	return db
}

func scopedTest(db *gorm.DB) {
	paged := []User{}
	if err := db.Scopes(paginate(1, 2)).Where("status = ?", "active").Order("created_at desc").Find(&paged).Error; err != nil {
		panic(err)
	}
	if len(paged) != 2 {
		panic(fmt.Sprintf("the expected number of users should be 2, but actual value is %d\n", len(paged)))
	} else {
		fmt.Println("the number of users on the current page is", len(paged))
	}
}

func paginate(pageNum, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// normalize input
		if pageNum <= 0 {
			pageNum = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		// calculate offset
		offset := (pageNum - 1) * pageSize

		// skip ${offset} records and return at most ${pageSize} records
		return db.Offset(offset).Limit(pageSize)
	}
}
