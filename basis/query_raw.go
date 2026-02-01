package basis

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func RawQueryTest() {
	db := setup("raw.db")

	rawGroupTest(db)
	rawUpdateTest(db)
	rawCountTest(db)
}

func rawGroupTest(db *gorm.DB) {
	type StatusSummary struct {
		Status string
		Total  int64
		AvgAge float64
	}

	ss := []StatusSummary{}
	start := time.Now().AddDate(0, 0, -60)
	end := time.Now()
	if err := db.Raw(`
		SELECT status, COUNT(*) AS total, AVG(age) AS avg_age
		FROM USERS
		WHERE created_at BETWEEN ? AND ?
		GROUP BY status
	`, start, end).Scan(&ss).Error; err != nil {
		panic(err)
	}
	fmt.Println("group users by status and age,", ss)
}

func rawUpdateTest(db *gorm.DB) {
	result := db.Exec("UPDATE users SET status = ? WHERE last_login_at < ?", "inactive", time.Now().AddDate(0, 0, -30))
	if result.Error != nil {
		panic(result.Error)
	}
	fmt.Println("rows affected:", result.RowsAffected)
}

func rawCountTest(db *gorm.DB) {
	var c int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM users
		WHERE status = ?
	`, "inactive").Scan(&c).Error; err != nil {
		panic(err)
	}
	fmt.Println("the number of inactive users is", c)
}
