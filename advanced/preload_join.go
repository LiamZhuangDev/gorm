package advanced

// | Aspect                  | `Preload`                    | `Joins`                         |
// | ----------------------- | ---------------------------- | ------------------------------- |
// | Purpose                 | Load related data            | Filter / query using relations  |
// | SQL                     | Multiple queries             | Single SQL with JOIN            |
// | Populates struct fields | ✅ Yes                       | ❌ No (unless selected manually)|
// | Avoids N+1              | ✅ Yes                       | ⚠️ Depends                      |
// | Best for                | API responses, object graphs | Search, filters, conditions     |
// | Supports has-many       | ✅                           | ⚠️ Duplicates rows              |
// | Supports many-to-many   | ✅                           | ⚠️ Complex                      |
//
// Preload - "I want the related data"
//
// var users []User
// db.Preload("Orders").Preload("Roles").Find(&users)
//
// What happens
// - Query users
// - Query orders WHERE user_id IN (...)
// - Query roles via join table
//
// SELECT * FROM users;
// SELECT * FROM orders WHERE user_id IN (...);
// SELECT users.*, user_roles.* FROM roles JOIN user_roles ...
//
// users[0].Orders // populated
// users[0].Roles  // populated
//
// Joins - "I want to filter using related tables"
//
// var users []User
// db.Joins("JOIN orders ON orders.user_id = users.id").Where("orders.status = ?", "delivered").Find(&users)
//
// SELECT users.*
// FROM users
// JOIN orders ON orders.user_id = users.id
// WHERE orders.status = 'delivered';
//
// users[0].Orders // empty, Joins does NOT populate associations
//
// Use Joins to FILTER, Preload to LOAD, if you want both, use both together.
//
// Users who have delivered orders, and load those orders:
// var users []User
// db.Joins("JOIN orders ON orders.user_id = users.id").
//     Where("orders.status = ?", "delivered").
// 	   Preload("Orders", "status = ?", "delivered").
// 	   Find(&users)
//
// Joins create duplicate records for has-many associations
//
// db.Joins("Orders").Find(&users)
//
// SELECT *
// FROM users
// LEFT JOIN orders ON orders.user_id = users.id;
//
// If user has 3 orders -> user row appears 3 times
// You must use db.Distinct("users.id")
//
// Preload is recommended for many-to-many relationship
// db.Preload("Roles").Find(&users)
// vs
// db.Joins("JOIN user_roles ur ON ur.user_id = users.id").
// 	   Joins("JOIN roles r ON r.id = ur.role_id").
// 	   Where("r.name = ?", "admin").
// 	   Find(&users)
//
// Final Takeaway: Preload builds object graphs and Joins shapes result sets.
