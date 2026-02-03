package basis

import (
	"fmt"

	"gorm.io/gorm"
)

func QueryTest() {
	db := setup("db/query.db")
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
