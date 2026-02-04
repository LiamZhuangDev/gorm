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

	// autoTransactionTest(db)
	manualTransactionTest(db)
}

func autoTransaction(db *gorm.DB, u *User) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// build the order
		items := []OrderItem{
			{ProductID: 1, Quantity: 1, Price: 999.00},
			{ProductID: 2, Quantity: 1, Price: 249.00},
		}
		order := Order{
			OrderNumber: "ORD-3001", // change to ORD-1001 to trigger roll back
			UserID:      u.ID,       // capture parameters in closure, and FK to users table
			TotalPrice:  items[0].Price*float64(items[0].Quantity) + items[1].Price*float64(items[1].Quantity),
			Status:      "created",
			Items:       items,
		}

		if err := tx.Create(&order).Error; err != nil {
			fmt.Println("fails to create order", err)
			return err // roll back
		}

		return nil
	})
}

func autoTransactionTest(db *gorm.DB) {
	// find the user
	userID := uint(1)
	var u User
	if err := db.
		Preload("Orders").
		Preload("Orders.Items").
		Preload("Orders.Items.Product").
		First(&u, userID).Error; err != nil {
		panic(err)
	}

	// take a snapshot before the tansaction
	before, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(before))

	// start the transaction
	if err := autoTransaction(db, &u); err != nil {
		fmt.Println("fails to create order", err)
	}

	// reload to verify
	var refreshed User
	if err := db.
		Preload("Orders").
		Preload("Orders.Items").
		Preload("Orders.Items.Product").
		First(&refreshed, userID).Error; err != nil {
		panic(err)
	}

	after, _ := json.MarshalIndent(&refreshed, "", "  ")
	fmt.Println(string(after))
}

func manualTransactionTest(db *gorm.DB) {
	// start a transaction
	tx := db.Begin()
	if tx.Error != nil {
		panic(tx.Error)
	}

	// protect against panics
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// find the user
	userID := uint(1)
	var u User
	if err := db.
		Preload("Orders").
		Preload("Orders.Items").
		Preload("Orders.Items.Product").
		First(&u, userID).Error; err != nil {
		fmt.Println("user not found", err)
		tx.Rollback()
	}

	// take a snapshot before change
	before, _ := json.MarshalIndent(u, "", "  ")
	fmt.Println(string(before))

	// build the order
	items := []OrderItem{
		{ProductID: 1, Quantity: 1, Price: 999.00},
		{ProductID: 2, Quantity: 1, Price: 249.00},
	}
	order := Order{
		OrderNumber: "ORD-3001", // change to ORD-1001 to trigger roll back
		UserID:      u.ID,       // capture parameters in closure, and FK to users table
		TotalPrice:  items[0].Price*float64(items[0].Quantity) + items[1].Price*float64(items[1].Quantity),
		Status:      "created",
		Items:       items,
	}

	if err := tx.Create(&order).Error; err != nil {
		fmt.Println("fails to create order", err)
		tx.Rollback()
	}

	// commit the transaction
	if err := tx.Commit().Error; err != nil {
		fmt.Println("fails to commit", err)
		tx.Rollback()
	}

	// reload to verify
	var refreshed User
	if err := db.
		Preload("Orders").
		Preload("Orders.Items").
		Preload("Orders.Items.Product").
		First(&refreshed, userID).Error; err != nil {
		panic(err)
	}

	after, _ := json.MarshalIndent(&refreshed, "", "  ")
	fmt.Println(string(after))
}

func savePointTest(db *gorm.DB) {

}

func nestedTransactionsTest(db *gorm.DB) {

}
