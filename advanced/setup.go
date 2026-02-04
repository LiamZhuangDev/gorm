package advanced

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var adminRole = Role{
	Name:        "admin",
	Description: "Administrator",
}

var userRole = Role{
	Name:        "user",
	Description: "Regular user",
}

func setup(dsn string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&User{}, &Profile{}, &Product{}, &Order{}, &OrderItem{}, &Role{}); err != nil {
		panic(err)
	}

	// Users and Roles is many-to-many relationship, when creating records with FullSaveAssociations enabled,
	// Same role name values may appear multiple times inthe same INSERT, but this field must be unique. So INSERT failed!
	// One solution is to create roles first, then reuse them later.
	if err := db.Create(&adminRole).Error; err != nil {
		panic(err)
	}
	if err := db.Create(&userRole).Error; err != nil {
		panic(err)
	}

	// Seed products first (independent entities, no foreign keys)
	// Products are referenced by order items, so they must exist before creating orders
	products := []Product{
		{
			Name:  "MacBook Pro 14",
			SKU:   "MBP-14-2024",
			Price: 1999.00,
		},
		{
			Name:  "iPhone 15 Pro",
			SKU:   "IPHONE-15-PRO",
			Price: 999.00,
		},
		{
			Name:  "AirPods Pro",
			SKU:   "AIRPODS-PRO",
			Price: 249.00,
		},
	}

	if err := db.Create(&products).Error; err != nil {
		panic(err)
	}

	user1 := User{
		Name:  "Alice",
		Email: "alice@example.com",
		Roles: []Role{adminRole, userRole},
		Profile: Profile{
			Nickname: "alice_w",
			Phone:    "3239000001",
			Address:  "123 Main St, CA",
		},
		Orders: []Order{
			{
				OrderNumber: "ORD-1001",
				Status:      "paid",
				TotalPrice:  2*products[2].Price + products[0].Price,
				Items: []OrderItem{
					{ProductID: products[2].ID, Quantity: 2, Price: products[2].Price},
					{ProductID: products[0].ID, Quantity: 1, Price: products[0].Price},
				},
			},
		},
	}

	user2 := User{
		Name:  "Bob",
		Email: "bob@example.com",
		Roles: []Role{userRole},
		Profile: Profile{
			Nickname: "bobby",
			Phone:    "3239000002",
			Address:  "456 Oak Ave, CA",
		},
		Orders: []Order{
			{
				OrderNumber: "ORD-1002",
				Status:      "delivered",
				TotalPrice:  3 * products[1].Price,
				Items: []OrderItem{
					{ProductID: products[1].ID, Quantity: 3, Price: products[1].Price},
				},
			},
			{
				OrderNumber: "ORD-1003",
				Status:      "shipped",
				TotalPrice:  4 * products[0].Price,
				Items: []OrderItem{
					{ProductID: products[0].ID, Quantity: 4, Price: products[0].Price},
				},
			},
		},
	}
	user3 := User{
		Name:  "Charlie",
		Email: "charlie@example.com",
		Roles: []Role{userRole},
		Profile: Profile{
			Nickname: "charlie_c",
			Phone:    "3239000003",
			Address:  "789 Pine Rd, CA",
		},
	}

	users := []User{user1, user2, user3}

	// Session with FullSaveAssociations: Ensures all nested associations are saved
	// Without this, GORM might skip zero-value associations
	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&users).Error; err != nil {
		panic(err)
	}

	return db
}
