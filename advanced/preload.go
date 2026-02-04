package advanced

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Preload in GORM is used for eager loading of associations.
// It lets you load related data in advance, instead of fetching it later one-by-one.
//
// #######################
// ## N+1 query problem ##
// #######################
//
// var users []User
// db.Find(&users) // 1 query for users
// for _, u := range users {
// 	db.Model(&u).Association("Profile").Find(&u.Profile) // N queries for profiles
// }
//
// ########################
// ## preload to rescue ##
// ########################
//
// var users []User
// db.Preload("Profile").Find(&users)
// SQL executed (conceptually)
// SELECT * From users; // 1 query for users
// SELECT * From profiles WHERE user_id IN (...); // 1 query for reading all related profiles for given users

func PreloadTest() {
	dsn := "db/preload.db"
	db := setup(dsn)

	preloadTest(db)
	conditionalPreloadTest(db)
	nestedPreloadTest(db)
	preloadAllTest(db)
	reversePreloadTest(db)
}

func preloadTest(db *gorm.DB) {
	var u User

	err := db.Preload("Roles").Preload("Profile").Preload("Orders").Where("email = ?", "alice@example.com").First(&u).Error
	if err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(b))
}

func conditionalPreloadTest(db *gorm.DB) {
	var u User

	err := db.Preload("Roles").Preload("Profile").Preload("Orders", "status = ?", "delivered").Where("email = ?", "bob@example.com").First(&u).Error
	if err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(b))
}

func nestedPreloadTest(db *gorm.DB) {
	var u User

	err := db.Preload("Roles").Preload("Profile").Preload("Orders").Preload("Orders.Items").Preload("Orders.Items.Product").Where("email = ?", "bob@example.com").First(&u).Error
	if err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(b))
}

// clause.Associations automatically preloads all associations of the model
// This is useful when you want to load all related data without specifying each association
// Note: This only preloads direct associations, NOT nested ones
func preloadAllTest(db *gorm.DB) {
	var u User

	err := db.Preload(clause.Associations).Where("email = ?", "bob@example.com").First(&u).Error
	if err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(b))
}

// Reversed preload for many-to-many relationship
// It works only when Role model has Users field.
func reversePreloadTest(db *gorm.DB) {
	var roles []Role

	// Find(&rolesWithUsers):
	// search roles table，retrieve all roles
	//
	// Preload("Users"):
	// For each role：
	// 		Search user_roles join table
	// 		Search users table
	// 		Populate role.Users
	if err := db.Preload("Users").Find(&roles).Error; err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(&roles, "", "  ")
	fmt.Println(string(b))
}
