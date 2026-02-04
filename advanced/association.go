package advanced

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

func AssociationTest() {
	dsn := "db/association.db"
	db := setup(dsn)

	findTest(db)
	appendTest(db)
	updateForBelongsToTest(db)
	updatesForHasOneTest(db)
	replaceForHasManyTest(db)
	replaceForMany2ManyTest(db)
	deleteAssociationTest(db)
	clearAssociationTest(db)
	countAssiciationTest(db)
}

func findTest(db *gorm.DB) {
	var alice User
	if err := db.First(&alice, "name = ?", "Alice").Error; err != nil {
		panic(err)
	}

	var roles []Role
	if err := db.Model(&alice).Association("Roles").Find(&roles); err != nil {
		panic(err)
	}

	j, _ := json.MarshalIndent(roles, "", "  ")
	fmt.Println(string(j))
}

// Append admin role to user: this adds entries to the user_roles join table
func appendTest(db *gorm.DB) {
	var u User

	if err := db.Preload("Roles").Where("email = ?", "charlie@example.com").First(&u).Error; err != nil {
		panic(err)
	}

	before, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(before))

	if err := db.Model(&u).Where("email = ?", "charlie@example.com").Association("Roles").Append(&adminRole); err != nil {
		panic(err)
	}

	after, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(after))
}

// Remove existing associated records / links and add only the new ones you pass in
// Relationship	   Association Action
// HasOne	       Update foreign key to point to new record
// BelongsTo	   Update FK on current table
// HasMany	       Clear old children, attach new ones
// Many2Many	   Delete join-table rows, insert new ones

// For a 1-on-1 (HasOne / BelongsTo) relationship, Updates should be used instead of Replace in almost all real-world cases.
// Replace tries to insert a new profile before removing the old one.
// So for a brief moment, two profiles pointed to the same user, which violates the one-to-one constraint.
func updatesForHasOneTest(db *gorm.DB) {
	var u User

	if err := db.Preload("Profile").First(&u, 1).Error; err != nil {
		panic(err)
	}

	before, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(before))

	newProfile := Profile{
		Phone:   "9990001234",
		Address: "548 Village DR., CA",
	}

	if err := db.Model(&u.Profile).Updates(newProfile).Error; err != nil {
		panic(err)
	}

	after, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(after))
}

func updateForBelongsToTest(db *gorm.DB) {
	// take any order belongs to Alice
	var bob User
	if err := db.Where("email = ?", "bob@example.com").First(&bob).Error; err != nil {
		panic(err)
	}

	var order Order
	if err := db.Take(&order, "user_id = ?", bob.ID).Error; err != nil {
		panic(err)
	}

	before, _ := json.MarshalIndent(order, "", "  ")
	fmt.Println(string(before))

	// replace the user the order belongs to
	var charlie User
	if err := db.Where("email = ?", "charlie@example.com").First(&charlie).Error; err != nil {
		panic(err)
	}

	// This throws Unsupported relations: User
	// because Order model doesn't have User field
	// if err := db.Model(&order).Association("User").Replace(&charlie); err != nil {
	// 	panic(err)
	// }

	if err := db.Model(&order).Update("user_id", charlie.ID).Error; err != nil {
		panic(err)
	}

	// Reload order to verify
	var updated Order
	if err := db.First(&updated, order.ID).Error; err != nil {
		panic(err)
	}

	after, _ := json.MarshalIndent(updated, "", "  ")
	fmt.Println(string(after))
}

func replaceForHasManyTest(db *gorm.DB) {
	var alice User
	db.Preload("Orders").First(&alice, "name = ?", "Alice")

	before, _ := json.MarshalIndent(alice, "", "  ")
	fmt.Println(string(before))

	// New orders that should become Alice's ONLY orders
	newOrders := []Order{
		{OrderNumber: "ORD-2001"},
		{OrderNumber: "ORD-2002"},
	}

	if err := db.Model(&alice).Association("Orders").Replace(&newOrders); err != nil {
		panic(err)
	}

	// Reload to verify
	after, _ := json.MarshalIndent(alice, "", "  ")
	fmt.Println(string(after))
}

func replaceForMany2ManyTest(db *gorm.DB) {
	var charlie User
	if err := db.Preload("Roles").Where("email = ?", "charlie@example.com").First(&charlie).Error; err != nil {
		panic(err)
	}

	before, _ := json.MarshalIndent(charlie, "", "  ")
	fmt.Println(string(before))

	var admin, user Role
	db.First(&admin, "name = ?", "admin")
	db.First(&user, "name = ?", "user")
	if err := db.Model(&charlie).Association("Roles").Replace(&admin, &user); err != nil {
		panic(err)
	}

	// Reload to verify
	after, _ := json.MarshalIndent(charlie, "", "  ")
	fmt.Println(string(after))
}

func deleteAssociationTest(db *gorm.DB) {
	var alice, bob User
	db.Preload("Orders").First(&alice, "name = ?", "Alice")
	db.Preload("Orders").First(&bob, "name = ?", "Bob")

	before, _ := json.MarshalIndent(alice, "", "  ")
	fmt.Println(string(before))

	before2, _ := json.MarshalIndent(bob, "", "  ")
	fmt.Println(string(before2))

	var orders []Order
	db.Where("order_number IN ?", []string{"ORD-1001", "ORD-1002"}).Find(&orders)

	if err := db.Model(&alice).Association("Orders").Delete(&orders); err != nil {
		panic(err)
	}

	if err := db.Model(&bob).Association("Orders").Delete(&orders); err != nil {
		panic(err)
	}

	after, _ := json.MarshalIndent(alice, "", "  ")
	fmt.Println(string(after))

	after2, _ := json.MarshalIndent(bob, "", "  ")
	fmt.Println(string(after2))
}

// Orders still exist, but they are no longer linked with Users
func clearAssociationTest(db *gorm.DB) {
	var u User
	db.Preload("Orders").First(&u, 1)

	before, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(before))

	if err := db.Model(&u).Association("Orders").Clear(); err != nil {
		panic(err)
	}

	// Reload to verify
	after, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(after))
}

func countAssiciationTest(db *gorm.DB) {
	var u User
	db.Preload("Orders").First(&u, 1)

	fmt.Printf("The count of Orders associations: %d\n", db.Model(&u).Association("Orders").Count())
}
