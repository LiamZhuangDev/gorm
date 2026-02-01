package main

import (
	basis "gorm/basis"
)

func main() {
	basis.CrudTest()
	basis.QueryTest()
	basis.RawQueryTest()
}

// import (
// 	"context"
// 	"fmt"

// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// )

// type Product struct {
// 	gorm.Model
// 	Code  string
// 	Price uint
// }

// func main() {
// 	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
// 	if err != nil {
// 		panic("failed to connect database")
// 	}

// 	ctx := context.Background()

// 	// Migrate the schema
// 	db.AutoMigrate(&Product{})

// 	// Create
// 	// gorm.G[T](db) returns a generic query context bound to the T model. It wraps *gorm.DB and provides a type-safe CRUD API.
// 	err = gorm.G[Product](db).Create(ctx, &Product{Code: "D42", Price: 100})
// 	if err != nil {
// 		fmt.Println("Failed to create/insert product")
// 	}

// 	// Read
// 	product, err := gorm.G[Product](db).Where("id = ?", 1).First(ctx) // find product with integer primary key
// 	fmt.Println("Found product: ", product)
// 	products, err := gorm.G[Product](db).Where("code = ?", "D42").Find(ctx) // find product with code D42
// 	fmt.Println("Found products: ", products)

// 	// Update - update product's price to 200
// 	rows, err := gorm.G[Product](db).Where("id = ?", product.ID).Update(ctx, "Price", 200)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Raws updated: ", rows)

// 	// Update - update multiple fields
// 	rows, err = gorm.G[Product](db).Where("id = ?", product.ID).Updates(ctx, Product{Code: "D42", Price: 100})
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Raws updated: ", rows)

// 	// Delete - delete product
// 	rows, err = gorm.G[Product](db).Where("id = ?", product.ID).Delete(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Raws updated: ", rows)
// }
