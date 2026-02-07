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
	PostWithTagsTest(db)
	SoftDeleteCommentTest(db)
	HardDeleteCommentTest(db)
}

// ===== Tags =====
var Tags = []Tag{
	{Name: "Go"},
	{Name: "Database"},
	{Name: "GORM"},
	{Name: "Backend"},
	{Name: "Cloud"},
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

	if err := db.Create(&Tags).Error; err != nil {
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
			selected[j] = Tags[r.Intn(len(Tags))]
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

func PublishPostWithTags(
	db *gorm.DB,
	userID uint,
	subject, content string,
	tagIDs []uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// find tags
		var tags []Tag
		if len(tagIDs) > 0 {
			if err := tx.Find(&tags, "id IN ?", tagIDs).Error; err != nil {
				return err // roll back
			}
		}

		// create a post
		p := Post{
			Subject: subject,
			Content: content,
			UserID:  userID,
		}
		if err := tx.Create(&p).Error; err != nil {
			return err // roll back
		}

		// attach tags to post
		if err := tx.Model(&p).Association("Tags").Append(tags); err != nil {
			return err // roll back
		}

		// or create with association
		// p1 := Post{
		// 	Subject: subject,
		// 	Content: content,
		// 	UserID:  userID,
		// 	Tags:    tags, // <-- attach tags here! It won't overwrite existing tags because the post is just created and no old tags exist!
		// }
		// if err := tx.Create(&p1).Error; err != nil {
		// 	return err // roll back
		// }

		// Commit
		return nil
	})
}

func PostWithTagsTest(db *gorm.DB) {
	userID := 1
	var u User
	if err := db.Preload("Posts").Preload("Posts.Tags").First(&u, userID).Error; err != nil {
		panic(err)
	}

	// print user before publish a post
	b, _ := json.MarshalIndent(&u, "", "  ")
	fmt.Println(string(b))

	// randomly select 1-3 tags
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(3) + 1
	tagIDs := make([]uint, n)
	for i := range n {
		tagIDs[i] = uint(r.Intn(len(Tags)) + 1)
	}

	// publish
	if err := PublishPostWithTags(db, uint(userID), "This is a new Post", "Just wanna say hello web3!", tagIDs); err != nil {
		panic(err)
	}

	// reload to verify
	var u1 User
	if err := db.Preload("Posts").Preload("Posts.Tags").Find(&u1, userID).Error; err != nil {
		panic(err)
	}

	b, _ = json.MarshalIndent(&u1, "", "  ")
	fmt.Println(string(b))
}

func SoftDeleteComment(db *gorm.DB, commentID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var comment Comment
		if err := tx.First(&comment, commentID).Error; err != nil {
			return err // roll back
		}

		if err := tx.Delete(&comment).Error; err != nil {
			return err // roll back
		}

		return nil // commit
	})
}

func SoftDeleteCommentTest(db *gorm.DB) {
	commentID := 1
	var c Comment
	if err := db.First(&c, commentID).Error; err != nil {
		panic(err)
	}

	// load the linked post
	var p Post
	if err := db.Preload("Comments").Find(&p, c.PostID).Error; err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(&p, "", "  ")
	fmt.Println(string(b))

	if err := SoftDeleteComment(db, uint(commentID)); err != nil {
		panic(err)
	}

	// reload to verify
	var p1 Post
	if err := db.Preload("Comments").First(&p1, c.PostID).Error; err != nil {
		panic(err)
	}

	b, _ = json.MarshalIndent(&p1, "", "  ")
	fmt.Println(string(b))

	fmt.Println("Use Unscoped to include soft-deleted comments")

	var p2 Post
	if err := db.Unscoped().Preload("Comments").First(&p2, c.PostID).Error; err != nil {
		panic(err)
	}

	b, _ = json.MarshalIndent(&p2, "", "  ")
	fmt.Println(string(b))
}

func HardDeleteComment(db *gorm.DB, commentID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var comment Comment
		if err := tx.First(&comment, commentID).Error; err != nil {
			return err // roll back
		}

		if err := tx.Unscoped().Delete(&comment).Error; err != nil { // use Unscoped for hard delete
			return err // roll back
		}

		return nil // commit
	})
}

func HardDeleteCommentTest(db *gorm.DB) {
	commentID := 1
	var c Comment
	if err := db.First(&c, commentID).Error; err != nil {
		panic(err)
	}

	// load the linked post
	var p Post
	if err := db.Preload("Comments").Find(&p, c.PostID).Error; err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(&p, "", "  ")
	fmt.Println(string(b))

	if err := HardDeleteComment(db, uint(commentID)); err != nil {
		panic(err)
	}

	// reload to verify
	var p1 Post
	if err := db.Unscoped().Preload("Comments").First(&p1, c.PostID).Error; err != nil {
		panic(err)
	}

	b, _ = json.MarshalIndent(&p1, "", "  ")
	fmt.Println(string(b))
}
