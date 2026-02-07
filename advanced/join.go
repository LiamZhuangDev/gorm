package advanced

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 1. A JOIN combines rows from two tables based on a related column.
// Every join has exactly two parts: LEFT TABLE + MATCH CONDITION + RIGHT TABLE
//
// 2. JOIN type semantics:
// INNER JOIN (the default, most common)
// Only rows that match on both sides survive.
//
// LEFT JOIN (very important)
// Keep all rows from the left table, even if there is no match.
//
// RIGHT JOIN (mirror image)
// Same as LEFT JOIN, but flipped.
//
// FULL OUTER JOIN
// Keep all rows from both tables.
//
// CROSS JOIN
// Every row joins with every row. Rarely useful, often dangerous.
//
// 3. JOIN vs Preload
// | Aspect                  | `Preload`                    | `Joins`                         |
// | ----------------------- | ---------------------------- | ------------------------------- |
// | Purpose                 | Load related data            | Filter / query using relations  |
// | SQL                     | Multiple queries             | Single SQL with JOIN            |
// | Populates struct fields | ✅ Yes                       | ❌ No (unless selected manually)|
// | Avoids N+1              | ✅ Yes                       | ⚠️ Depends                      |
// | Best for                | API responses, object graphs | Search, filters, conditions     |
// | Supports has-many       | ✅                           | ⚠️ Duplicates rows              |
// | Supports many-to-many   | ✅                           | ⚠️ Complex                      |
//
// Preload - "I want the related data"
//
// var users []User
// db.Preload("Orders").Preload("Roles").Find(&users)
//
// What happens
// - Query users
// - Query orders WHERE user_id IN (...)
// - Query roles via join table
//
// SELECT * FROM users;
// SELECT * FROM orders WHERE user_id IN (...);
// SELECT users.*, user_roles.* FROM roles JOIN user_roles ...
//
// users[0].Orders // populated
// users[0].Roles  // populated
//
// Joins - "I want to filter using related tables"
//
// var users []User
// db.Joins("JOIN orders ON orders.user_id = users.id").Where("orders.status = ?", "delivered").Find(&users)
//
// SELECT users.*
// FROM users
// JOIN orders ON orders.user_id = users.id
// WHERE orders.status = 'delivered';
//
// users[0].Orders // empty, Joins does NOT populate associations
//
// Use Joins to FILTER, Preload to LOAD, if you want both, use both together.
//
// Users who have delivered orders, and load those orders:
// var users []User
// db.Joins("JOIN orders ON orders.user_id = users.id").
//     Where("orders.status = ?", "delivered").
// 	   Preload("Orders", "status = ?", "delivered").
// 	   Find(&users)
//
// Joins create duplicate records for has-many associations
// db.Joins("Orders").Find(&users)
//
// SELECT *
// FROM users
// LEFT JOIN orders ON orders.user_id = users.id;
//
// For each row in users, match every row in orders where orders.user_id = users.id and output one row per match.
// If user has 3 orders -> user row appears 3 times
// You must use db.Distinct("users.id")
//
// Preload is recommended for many-to-many relationship
// db.Preload("Roles").Find(&users)
// vs
// db.Joins("JOIN user_roles ur ON ur.user_id = users.id").
// 	   Joins("JOIN roles r ON r.id = ur.role_id").
// 	   Where("r.name = ?", "admin").
// 	   Find(&users)
//
// Final Takeaway: Preload builds object graphs and Joins shapes result sets.

type User4JoinDemo struct {
	ID     uint
	Name   string
	Orders []Order4JoinDemo `gorm:"foreignKey:UserID;references:ID"`
}

type Order4JoinDemo struct {
	ID        uint
	UserID    uint
	Status    string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (User4JoinDemo) TableName() string {
	return "users"
}

func (Order4JoinDemo) TableName() string {
	return "orders"
}

func JoinTest() {
	// Seed test data
	o1 := Order4JoinDemo{
		Status: "created",
	}

	o2 := Order4JoinDemo{
		Status: "paid",
	}

	o3 := Order4JoinDemo{
		Status: "delivered",
	}

	o4 := Order4JoinDemo{
		Status: "shipped",
	}

	o5 := Order4JoinDemo{
		Status: "paid",
	}

	u1 := User4JoinDemo{
		Name:   "Alice",
		Orders: []Order4JoinDemo{o1, o4},
	}

	u2 := User4JoinDemo{
		Name:   "Frank",
		Orders: []Order4JoinDemo{o2, o5},
	}

	u3 := User4JoinDemo{
		Name:   "Charlie",
		Orders: []Order4JoinDemo{o3},
	}

	u4 := User4JoinDemo{
		Name:   "Jordan",
		Orders: []Order4JoinDemo{},
	}

	users := []User4JoinDemo{u1, u2, u3, u4}

	// insert test data
	dsn := "db/join.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&User4JoinDemo{}, &Order4JoinDemo{}); err != nil {
		panic(err)
	}

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Create(users).Error; err != nil {
		panic(err)
	}

	innerJoinTest(db)
	leftJoinTest(db)
}

func innerJoinTest(db *gorm.DB) {
	type Result struct {
		UserID  uint
		Name    string
		OrderID uint
		Status  string
	}

	var result []Result
	db.Table("users").
		Select("users.id as user_id, users.name, orders.id as order_id, orders.status").
		Joins("JOIN orders ON users.id = orders.user_id").
		Scan(&result)

	b, _ := json.MarshalIndent(&result, "", "  ")
	fmt.Println(string(b))
}

func leftJoinTest(db *gorm.DB) {
	join4FilteringTest(db)
	join4AggregationTest(db)
	join4SortingTest(db)
}

// LEFT JOIN is using for filtering parent by *child* condition
//
// "Orders" vs "orders"
// db.Joins("Orders") means join the association defined by the Orders field on User model
// GORM figures out:
// - related model: Order
// - table name: orders
// - foreign key: orders.user_id
// Then build the LEFT JOIN automatically:
// SELECT users.*, orders.*
// FROM users
// LEFT JOIN orders ON orders.user_id = users.id AND orders.status = "paid";
//
// Uppercase "Orders" is for GORM world and lowercase "orders" must be used in raw SQL.
func join4FilteringTest(db *gorm.DB) {
	var users []User4JoinDemo
	if err := db.Preload("Orders").Model(&User4JoinDemo{}).
		Select("users.*").
		Joins("JOIN orders ON orders.user_id = users.id").
		Where("orders.status = ?", "paid").
		Distinct(). // Distinct("users.id") rewrites the SELECT list to only users.id, change to no args to keep the existing list. Or we can use Group("users.id")
		Find(&users).Error; err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(&users, "", "  ")
	fmt.Println(string(b))
}

func join4AggregationTest(db *gorm.DB) {
	type Result struct {
		UserID     uint
		Name       string
		OrderCount int64
	}
	var result []Result
	db.Model(&User4JoinDemo{}).
		Joins("LEFT JOIN orders ON orders.user_id = users.id").
		Select("users.id as user_id, users.name, COUNT(orders.id) as order_count").
		Group("users.id").
		Scan(&result)

	b, _ := json.MarshalIndent(&result, "", "  ")
	fmt.Println(string(b))
}

func join4SortingTest(db *gorm.DB) {
	var users []User4JoinDemo
	db.Model(&User4JoinDemo{}).
		Preload("Orders").
		Joins("LEFT JOIN orders ON orders.user_id = users.id"). // Use JOIN (INNER JOIN) if you only want users that have at least one order.
		Select("users.*").
		Order("orders.created_at DESC").
		Find(&users)

	b, _ := json.MarshalIndent(&users, "", "  ")
	fmt.Println(string(b))
}
