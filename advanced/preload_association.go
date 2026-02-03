package advanced

import "fmt"

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

	var u User
	err := db.Preload("Profile").Preload("Orders").Preload("Orders.Items").Preload("Orders.Items.Product").Where("email = ?", "bob@example.com").First(&u).Error
	if err != nil {
		panic(err)
	}

	fmt.Println("user:", u)
}
