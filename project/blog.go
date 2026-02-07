package project

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func BlogTest() {
	dsn := "db/blog.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&User{}, &Post{}, &Tag{}, &Comment{}); err != nil {
		panic(err)
	}

	if err := SeedBlogData(db); err != nil {
		panic(err)
	}

	GetUserLatestPostsTest(db)
	CountPostCommentsTest(db)
}

func SeedBlogData(db *gorm.DB) error {
	// ===== Users =====
	users := make([]User, 5)
	for i := range 5 {
		users[i] = User{
			Name:  fmt.Sprintf("user_%d", i+1),
			Email: fmt.Sprintf("user_%d@test.com", i+1),
		}
	}
	if err := db.Create(&users).Error; err != nil {
		return err
	}

	// ===== Tags =====
	tags := []Tag{
		{Name: "Go"},
		{Name: "Database"},
		{Name: "GORM"},
		{Name: "Backend"},
		{Name: "Cloud"},
	}
	if err := db.Create(&tags).Error; err != nil {
		return err
	}

	// ===== Posts =====
	posts := make([]Post, 20)
	for i := range 20 {
		posts[i] = Post{
			Subject: fmt.Sprintf("Post Subject %d", i+1),
			Content: fmt.Sprintf("This is the content of post %d", i+1),
			UserID:  users[i%len(users)].ID, // 平均分给 5 个用户
		}
	}
	if err := db.Create(&posts).Error; err != nil {
		return err
	}

	// ===== Post <-> Tags (many2many) =====
	// 每篇文章随机挂 1~3 个标签
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range posts {
		n := r.Intn(3) + 1
		selected := make([]Tag, n)
		for j := range n {
			selected[j] = tags[r.Intn(len(tags))]
		}

		if err := db.Model(&posts[i]).Association("Tags").Append(selected); err != nil {
			return err
		}
	}

	// ===== Comments =====
	comments := make([]Comment, 10)
	for i := range 10 {
		comments[i] = Comment{
			PostID:  posts[r.Intn(len(posts))].ID,
			UserID:  users[r.Intn(len(users))].ID,
			Content: fmt.Sprintf("Comment %d content", i+1),
		}
	}
	if err := db.Create(&comments).Error; err != nil {
		return err
	}

	return nil
}

func GetUserLatestPosts(db *gorm.DB, userID uint, number int) ([]Post, error) {
	var posts []Post
	if err := db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(number).
		Preload("Tags").
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func GetUserLatestPostsTest(db *gorm.DB) {
	userID := uint(1)
	numOfPosts := 10
	posts, err := GetUserLatestPosts(db, userID, numOfPosts)

	if err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(posts, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

type PostWithCommentCount struct {
	Post
	CommentCount int64
}

func CountPostComments(db *gorm.DB) ([]PostWithCommentCount, error) {
	var result []PostWithCommentCount

	if err := db.Model(&Post{}).
		Select("posts.*, COUNT(comments.id) AS comment_count").
		Joins("LEFT JOIN comments ON comments.post_id = posts.id").
		Group("posts.id").
		Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func CountPostCommentsTest(db *gorm.DB) {
	result, err := CountPostComments(db)

	if err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
