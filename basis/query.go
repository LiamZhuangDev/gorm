package basis

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func QueryTest() {
	db := setup("query.db")
	scopedTest(db)
	likeTest(db)
	groupTest(db)
	if users, err := searchUsersByEmail(db, "%@gmail.com", 1, 2); err != nil {
		panic(err)
	} else {
		fmt.Printf("%d users on current page: %v\n", len(users), users)
	}
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

func youngUsers(min, max, pageNum, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(ageBetween(min, max), paginate(pageNum, pageSize))
	}
}

// Use LIKE for pattern matching (SQL wildcard: %)
func likeTest(db *gorm.DB) {
	u := []User{}

	if err := db.Select("name", "email").Where("email LIKE ?", "%@gmail.com").Find(&u).Error; err != nil {
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

func searchUsersByEmail(db *gorm.DB, emailPattern string, pageNum, pageSize int) ([]User, error) {
	var users []User

	if err := db.Scopes(paginate(pageNum, pageSize)).Where("email LIKE ?", emailPattern).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}
