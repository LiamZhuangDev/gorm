// Audit fields in GORM are metadata fields used to track when and by whom
// a record was created, updated or deleted.

package advanced

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Audit struct {
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	CreatedBy uint
	UpdatedBy uint
	DeletedBy *uint // NULL means not deleted
}

type OrderWithAudit struct {
	ID     uint
	Status string
	Audit
}

type ctxKeyUserID struct{} // typed context key to prevent collision

func AuditTest() {
	dsn := "db/audit.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&OrderWithAudit{}); err != nil {
		panic(err)
	}

	AuditHooksTest(db)
	AuditCallbackTest(db)
}

func AuditHooksTest(db *gorm.DB) {
	// insert an order
	o := OrderWithAudit{
		Status: "created",
	}
	ctx := context.WithValue(context.Background(), ctxKeyUserID{}, uint(42))
	if err := db.WithContext(ctx).Save(&o).Error; err != nil {
		panic(err)
	}

	// reload to verify the audit fields
	var o1 OrderWithAudit
	if err := db.Where("ID = ?", o.ID).First(&o1).Error; err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(&o1, "", "  ")
	fmt.Println(string(b))
}

func (o *OrderWithAudit) BeforeCreate(tx *gorm.DB) error {
	if uid, ok := tx.Statement.Context.Value(ctxKeyUserID{}).(uint); ok {
		o.CreatedBy = uid
		o.UpdatedBy = uid
	}
	return nil
}

func (o *OrderWithAudit) BeforeUpdate(tx *gorm.DB) error {
	if uid, ok := tx.Statement.Context.Value("userID").(uint); ok {
		o.UpdatedBy = uid
	}
	return nil
}

// Simplified Create pipeline:
// BeforeCreate(model hook)
// gorm:before_create <-- callback registered before this step
// gorm:create
// gorm:after_create
// AfterCreate(model hook)
func AuditCallbackTest(db *gorm.DB) {
	// instead of per-model hooks, we can use global GORM callback
	db.Callback().Create().Before("gorm:create").
		Register("audit:create", func(tx *gorm.DB) {
			if uid, ok := tx.Statement.Context.Value(ctxKeyUserID{}).(uint); ok {
				tx.Statement.SetColumn("CreatedBy", uid)
				tx.Statement.SetColumn("UpdatedBy", uid)
			}
		})

	db.Callback().Update().Before("gorm:update").
		Register("audit:update", func(tx *gorm.DB) {
			if uid, ok := tx.Statement.Context.Value(ctxKeyUserID{}).(uint); ok {
				tx.Statement.SetColumn("UpdatedBy", uid)
			}
		})

	// insert an order
	o := OrderWithAudit{
		Status: "created",
	}

	ctx := context.WithValue(context.Background(), ctxKeyUserID{}, uint(42))
	if err := db.WithContext(ctx).Save(&o).Error; err != nil {
		panic(err)
	}

	// reload to verify
	var o1 OrderWithAudit
	if err := db.Where("ID = ?", o.ID).First(&o1).Error; err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(&o1, "", "  ")
	fmt.Println(string(b))
}
