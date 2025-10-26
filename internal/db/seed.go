package db

import (
	"context"
	"log"
	"strconv"

	"github.com/Chandan185/Societal/internal/store"
)

func Seed(store store.Storage) {
	ctx := context.Background()
	users := generateUsers(100)
	for _, user := range users {
		if err := store.Users.Create(ctx, user); err != nil {
			log.Println("Error seeding user:", err)
			return
		}

	}
	posts := generatePosts(users, 200)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error seeding post:", err)
			return
		}
	}

	comments := generateComments(users, posts, 500)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("Error seeding comment:", err)
			return
		}
	}
	log.Println("Database seeding completed successfully.")
}

func generateUsers(n int) []*store.User {
	users := make([]*store.User, n)
	for i := 0; i < n; i++ {
		users[i] = &store.User{
			Username: "user" + strconv.Itoa(i),
			Email:    "user" + strconv.Itoa(i) + "@example.com",
			Password: "password123",
		}
	}
	return users
}

func generatePosts(users []*store.User, n int) []*store.Post {
	posts := make([]*store.Post, n)
	userCount := len(users)
	for i := 0; i < n; i++ {
		user := users[i%userCount]
		posts[i] = &store.Post{
			Title:   "Post Title " + strconv.Itoa(i),
			Content: "This is the content of post number " + strconv.Itoa(i),
			USERID:  user.ID,
			Tags:    []string{"tag1", "tag2"},
		}
	}
	return posts
}

func generateComments(users []*store.User, posts []*store.Post, n int) []*store.Comment {
	comments := make([]*store.Comment, n)
	userCount := len(users)
	postCount := len(posts)
	for i := 0; i < n; i++ {
		user := users[i%userCount]
		post := posts[i%postCount]
		comments[i] = &store.Comment{
			PostID:  post.ID,
			UserID:  user.ID,
			Content: "This is a comment number " + strconv.Itoa(i),
		}
	}
	return comments
}
