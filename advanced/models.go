package advanced

import "time"

// ERD (Entity-Relationship Diagram) arrows always point from the foreign key holder to the primary key owner.
// +-------------------+            +-------------------+
// |       users       |            |      profiles     |
// +-------------------+            +-------------------+
// | PK id             |            | PK id             |
// | name              |◄───────────| FK user_id (UNQ) |
// | email             |   1 : 1    | nickname          |
// | created_at        |            | phone             |
// | updated_at        |            | address           |
// +-------+-----------+            | created_at        |
//         ^                        | updated_at        |
//         |                        +-------------------+
//         |
//         | 1 : N
//         |
//         |
// +-------------------+            +-------------------+
// |      orders       |            |    order_items    |
// +-------------------+            +-------------------+
// | PK id             |◄───────────| PK id             |
// | order_no (UNQ)    |   1 : N    | FK order_id       |
// | FK user_id        |            | FK product_id     |
// | total_price       |            | quantity          |
// | status            |            | unit_price        |
// | created_at        |            | created_at        |
// | updated_at        |            | updated_at        |
// +-------------------+            +-------------------+
//                                          |
//                                          | N : 1
//                                          |
//                                          v
//                               +-------------------+
//                               |     products      |
//                               +-------------------+
//                               | PK id             |
//                               | name              |
//                               | sku (UNQ)         |
//                               | price             |
//                               | created_at        |
//                               | updated_at        |
//                               +-------------------+
//
// +-------------------+      M : N      +-------------------+
// |       users       |◄───────────────►|       roles       |
// +-------------------+   user_roles    +-------------------+
// | PK id             |                 | PK id             |
// | name              |                 | name (UNQ)        |
// | email             |                 | description       |
// +-------------------+                 | created_at        |
//         ^                             | updated_at        |
//         |                             +-------------------+
//         |                                          ^
//         |        +-------------------------+       |
//         |        |      user_roles         |       |
//         |        |-------------------------|       |
//         ---------| FK user_id              |       |
//                  | FK role_id              |--------
//                  +-------------------------+

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:64;not null"`
	Email     string    `gorm:"size:64;uniqueIndex;not null"`
	Profile   Profile   // has-One
	Orders    []Order   // has-many
	Roles     []Role    `gorm:"many2many:user_roles"` // Many-to-many, link via user_roles join table
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Profile struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex"` // Foreign key to users table, enforce one-to-one relationship
	Nickname  string    `gorm:"size:64"`
	Phone     string    `gorm:"uniqueIndex;not null"`
	Address   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Order struct {
	ID          uint        `gorm:"primaryKey"`
	OrderNumber string      `gorm:"uniqueIndex"`
	UserID      uint        // Foreign key to users table
	Items       []OrderItem // Has-Many
	TotalPrice  float64
	Status      string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey"`
	OrderID   uint    // Foreign key to orders table
	ProductID uint    // Foreign key to products table
	Product   Product // Belongs To: OrderItem belongs to one product
	Quantity  uint
	Price     float64
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Product struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	SKU       string `gorm:"uniqueIndex"`
	Price     float64
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// GORM never auto-loads reverse associations
// So Role.Users == null is correct behavior
// Prevents infinite loops & performance disasters
// We can hide reverse fields from JSON (json:"-")
type Role struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"uniqueIndex"`
	Description string    `gorm:"size:64"`
	Users       []User    `gorm:"many2many:user_roles" json:"-"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
