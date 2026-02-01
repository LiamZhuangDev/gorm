package basis

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func QueryTest() {
	db := setup("query.db")
	scopedTest(db)
	likeTest(db)
	groupTest(db)
	// More examples:
	// db.Where("status IN ?", []string{"active", "pending"}).Find(&users)
	// db.Where("status = ? AND age > ?", "active", 25).Find(&users)
}

func setup(dsn string) *gorm.DB {
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
			Email:  "fiona@gmail.com",
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

// In GORM, Scopes are a way to reuse and compose query conditions cleanly.
// - Reuse common query logic
// - Compose filters dynamically
// - only build queries, not execute them (call First, Take, Update, Delete, etc to execute)
func scopedTest(db *gorm.DB) {
	paged := []User{}
	if err := db.Scopes(active(), ageBetween(30, 50), paginate(1, 2), orderByCreated(true)).Find(&paged).Error; err != nil {
		panic(err)
	}
	if len(paged) != 2 {
		panic(fmt.Sprintf("the expected number of users should be 2, but actual value is %d\n", len(paged)))
	} else {
		fmt.Println("the number of users on the current page is", len(paged))
	}
}

// Status scope
func active() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", "active")
	}
}

// Age scope
func ageBetween(min, max int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("age BETWEEN ? AND ?", min, max)
	}
}

// Paginate scope
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

// Sorting scope
func orderByCreated(desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("created_at DESC")
		}
		return db.Order("created_at ASC")
	}
}

// Use LIKE for pattern matching (SQL wildcard: %)
func likeTest(db *gorm.DB) {
	u := []User{}

	if err := db.Where("email LIKE ?", "%@gmail.com").Select("name", "email").Find(&u).Error; err != nil {
		panic(err)
	}
	fmt.Println("users found:", u)
}

// Aggregation
// Group is required when using aggregate functions like COUNT()
func groupTest(db *gorm.DB) {
	type StatusCount struct {
		Status string
		Total  int64
	}
	sc := []StatusCount{}
	if err := db.Model(&User{}).Select("status, COUNT(*) as total").Group("status").Order("total DESC").Scan(&sc).Error; err != nil {
		panic(err)
	}
	fmt.Println("group users by status,", sc)
}
