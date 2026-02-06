// Optimistic Locking in GORM is a concurrency control mechanism that prevents lost updates
// when multiple transactions try to update the same row without using database locks (SELECT FOR UPDATE)
//
// "lost update" example, consider two transactions:
// T1: SELECT status FROM orders WHERE id = 1; -- sees "created"
// T2: SELECT status FROM orders WHERE id = 1; -- sees "created"
//
// T1: UPDATE orders SET status = 'paid' WHERE id = 1;
// T2: UPDATE orders SET status = 'cancelled' WHERE id = 1;
//
// Both UPDATEs succeed, but the last commit wins and one update is silently lost

package advanced

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/optimisticlock"
)

type OrderWithOptLock struct {
	ID          uint   `gorm:"primaryKey"`
	OrderNumber string `gorm:"uniqueIndex"`
	Amount      float64
	Status      string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	Version     optimisticlock.Version
}

// DO NOT use *Save* API in optimistic locking.
// Save behaves like:
// 1. try Update
// 2. if RowsAffected == 0
// 3. Assume the row does not exist and then perform UPSERT (INSERT ... ON CONFLICT DO UPDATE)
// This behavior is by design BUT breaks optimistic locking.
// So ALWAYS use Update/Updates in optimistic locking.
func OptimisticLockingTest() {
	// setup
	dsn := "db/opt_lock.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	// create order table if not exists
	if err := db.AutoMigrate(&OrderWithOptLock{}); err != nil {
		panic(err)
	}

	// insert an order
	order := OrderWithOptLock{
		OrderNumber: "ORD-888",
		Status:      "created",
	}
	if err := db.Create(&order).Error; err != nil {
		panic(err)
	}

	// run two transactions attempt to update the same order
	var wg sync.WaitGroup
	wg.Add(2)
	start := make(chan struct{})

	// Transaction A
	go func() {
		defer wg.Done()

		var o OrderWithOptLock
		db.First(&o, 1) // reads version = 1

		<-start // wait here

		// DO NOT use Save API
		// o.Status = "paid"
		// result := db.Save(&o)

		result := db.Model(&o).Update("status", "paid")

		if result.RowsAffected == 0 {
			fmt.Println("Tx A: conflict detected")
			return
		}
		fmt.Println("Tx A: update success")
	}()

	// Transaction B
	go func() {
		defer wg.Done()

		var o OrderWithOptLock
		db.First(&o, 1) // also reads version = 1

		<-start // wait here

		// DO NOT use Save API
		// o.Status = "cancelled"
		// result := db.Save(&o)

		result := db.Model(&o).Update("status", "cancelled")

		if result.RowsAffected == 0 {
			fmt.Println("Tx B: conflict detected")
			return
		}
		fmt.Println("Tx B: update success")
	}()

	time.Sleep(50 * time.Millisecond)
	close(start) // Unblocks ALL current and future receivers
	wg.Wait()
}
