// In GORM, a transaction groups multiple database operations into a single all-or-nothing unit of work.
// If any step fails, everything is rolled back.
// If all steps succeed, the changes are committed.
//
// Why transactions matter?
// - Multipe DB operations must stay consistent
// - Partial updates would leave data corrupted
// For exampe, create an order and reduce inventory. If inventory update fails, the order must not exist.

package advanced

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

func TransactionTest() {
	dsn := "db/transaction.db"
	db := setup(dsn)

	AutoTransactionTest(db)
}

func AutoTransactionTest(db *gorm.DB) {
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Find the user
		var u User
		if err := tx.Preload("Orders").Preload("Orders.Items").Preload("Orders.Items.Product").First(&u, 1).Error; err != nil {
			fmt.Println("user not found")
			return err // roll back
		}

		before, _ := json.MarshalIndent(u, "", "  ")
		fmt.Println(string(before))

		// 2. Build the order
		items := []OrderItem{
			{ProductID: 1, Quantity: 1, Price: 999.00},
			{ProductID: 2, Quantity: 1, Price: 249.00},
		}
		order := Order{
			OrderNumber: "ORD-3001", // Change to ORD-1001 to trigger roll back
			UserID:      u.ID,       // FK to users table
			TotalPrice:  items[0].Price*float64(items[0].Quantity) + items[1].Price*float64(items[1].Quantity),
			Status:      "created",
			Items:       items,
		}
		if err := tx.Create(&order).Error; err != nil {
			fmt.Println("fails to create order", err)
			return err // roll back
		}

		// commit
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	var u User
	if err := db.Preload("Orders").Preload("Orders.Items").Preload("Orders.Items.Product").Find(&u, 1).Error; err != nil {
		panic(err)
	}

	after, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(after))
}

func ManualTransactionTest(db *gorm.DB) {

}

func StopPointTest(db *gorm.DB) {

}
