// 1. What is soft delete?
// soft delete means the row is NOT physically removed from the database.
// It is marked as deleted and normal queries automatically ignore it.
// In contrast, hard delete actually removes the row.
//
// 2. How soft delete works in GORM?
// GORM uses a special field: gorm.DeletedAt
//
// Instead of:
// DELETE FROM users WHERE id = 1;
//
// GORM executes:
// UPDATE users
// SET deleted_at = CURRENT_TIMESTAMP
// WHERE id = 1;
//
// 3. How normal queries work on soft deleted records?
// For db.Find(&users), SQL generated: SELECT * FROM users WHERE deleted_at IS NULL;
package advanced

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type OrderWithSoftDelete struct {
	ID          uint   `gorm:"primaryKey"`
	OrderNumber string `gorm:"uniqueIndex"`
	UserID      uint   // Foreign key to users table
	TotalPrice  float64
	Status      string
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func SoftDeleteTest() {
	dsn := "db/soft_delete.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&OrderWithSoftDelete{}); err != nil {
		panic(err)
	}

	order := OrderWithSoftDelete{
		UserID:      1,
		OrderNumber: "ORD-999",
		TotalPrice:  189.45,
		Status:      "created",
	}

	if err := db.Create(&order).Error; err != nil {
		panic(err)
	}

	// soft delete the order
	// add Unscoped for hard delete: db.Unscoped().Delete(..)
	if err := db.Delete(&order).Error; err != nil {
		panic(err)
	}

	// verify normal queries can no longer find the order
	var o1 OrderWithSoftDelete
	if err := db.First(&o1, order.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("Soft deleted records can no longer be found by using normal queries")
		} else {
			panic(err)
		}
	}

	// verify Unscoped queries can find the soft deleted orders
	var exists bool
	if err := db.Unscoped().
		Model(&OrderWithSoftDelete{}).
		Select("count(*) > 0").
		Where("id = ?", order.ID).
		Scan(&exists).Error; err != nil {
		panic(err)
	}

	if exists {
		fmt.Println("Soft deleted records can be found by unscoped queries")
	} else {
		panic("Soft deleted records should be found by unscoped quries")
	}
}
