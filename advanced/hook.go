// 1. In GORM, a hook is a lifecycle callback, a method you define on a model that
// GORM automatically calls before or after certain database operations.
//
// +--------------------------------------------------+
// |Operations| Hooks        | when it runs           |
// +----------+--------------+------------------------+
// | Create   | BeforeCreate | Before INSERT          |
// |          | AfterCreate  | After INSERT           |
// |--------------------------------------------------|
// | Find     | AfterFind    | After SELECT           |
// |--------------------------------------------------|
// | Update   | BeforeUpdate | Before UPDATE          |
// |          | AfterUpdate  | After UPDATE           |
// |--------------------------------------------------|
// | Save     | BeforeSave   | Before CREATE or UPDATE|
// |          | AfterSave    | After CREATE or UPDATE |
// |--------------------------------------------------|
// | Delete   | BeforeDelete | Before DELETE          |
// |          | AfterDelete  | After DELETE           |
// +--------------------------------------------------+
//
// 2. Why hooks are useful? Hooks are good for *cross-cutting concerns:
// - Auto-set timestamps
// - Validate fields
// - Normalize data
// - Soft delete logic
// - Audit logs
// - Generate IDs
//
// *Cross-cutting concerns are software functionalities—such as logging, security, caching, and error handling—that are needed across multiple, unrelated modules of an application.
// Unlike core business logic, these concerns "cut across" the entire system, leading to code scattering and tangling if not managed properly.
//
// 3. Hooks run inside the same transaction as the operation. If a hook returns an error, the operation is rolled back.
//
// 4. GORM automatically starts a transaction for you when you call operations like Create, Save, Update, or Delete.
// For example, for db.Create(&user), GORM does something below:
// db.Transaction(func(tx *gorm.DB) error {
//     // call BeforeSave
//     // call BeforeCreate
//     // INSERT
//     // call AfterCreate
//     // call AfterSave
//     return nil
// })

package advanced

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.Email = strings.ToLower(u.Email)
	return nil
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	var sum float64

	for _, it := range o.Items {
		sum += it.Price * float64(it.Quantity)
	}

	if sum <= 0 {
		return errors.New("invalid amount")
	}

	o.TotalPrice = sum

	return nil
}

// func (o *Order) BeforeCreate(tx *gorm.DB) error {
// 	var sum float64

// 	if err := tx.Model(&OrderItem{}).
// 		Where("order_id = ?", o.ID).
// 		Select("COALESCE(SUM(price * quantity), 0)").
// 		Scan(&sum).Error; err != nil {
// 		return err
// 	}

// 	if o.TotalPrice != sum {
// 		return fmt.Errorf("order total mismatch: total=%.2f, expected=%.2f", o.TotalPrice, sum)
// 	}

// 	return nil
// }

func HookTest() {
	dsn := "db/hook.db"
	db := setup(dsn)

	u := User{
		Name:  "Frank",
		Email: "FRANK@EXAMPLE.COM",
		Roles: []Role{adminRole, userRole},
		Profile: Profile{
			Nickname: "frank_z",
			Phone:    "5639000001",
			Address:  "903 Main St, CA",
		},
		Orders: []Order{
			{
				OrderNumber: "ORD-1045",
				Status:      "created",
				Items: []OrderItem{
					{ProductID: products[2].ID, Quantity: 2, Price: products[2].Price},
					{ProductID: products[0].ID, Quantity: 1, Price: products[0].Price},
				},
			},
		},
	}

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&u).Error; err != nil {
		panic(err)
	}

	// reload to verify
	if err := db.Preload("Orders").First(&u).Error; err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(&u, "", "  ")
	fmt.Println(string(b))
}
