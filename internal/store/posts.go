package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	USERID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int64     `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentCount int64 `json:"comment_count"`
}

type PostStore struct {
	db *sql.DB
}

func (p *PostStore) Create(ctx context.Context, post *Post) error {
	query :=
		`INSERT INTO posts (content, title, user_id, tags)
	VALUES($1,$2,$3,$4) RETURNING id, created_at, updated_at`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := p.db.QueryRowContext(ctx, query, post.Content, post.Title, post.USERID, pq.Array(post.Tags)).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)

	return err
}

func (p *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `SELECT id, user_id, title, content, tags, created_at, updated_at, version FROM posts WHERE id = $1`
	post := &Post{}
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := p.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.USERID, &post.Title, &post.Content, pq.Array(&post.Tags), &post.CreatedAt, &post.UpdatedAt, &post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return post, err
}

func (p *PostStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return err
}

func (p *PostStore) Update(ctx context.Context, post *Post) error {
	query := `UPDATE posts SET title=$1, content=$2, updated_at=NOW(), version=version+1 WHERE id=$3 AND Version=$4 RETURNING version`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := p.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

func (p *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	query := `SELECT p.id, p.user_id, p.title, p.content, p.tags, p.created_at, p.updated_at, p.version,u.username, COUNT(c.id) as comments_count
	FROM posts p
	LEFT JOIN comments c ON p.id = c.post_id
	LEFT JOIN users u ON p.user_id = u.id
	JOIN followers f ON p.user_id = f.follower_id OR p.user_id = $1
	WHERE 
		f.user_id = $1 OR p.user_id = $1 AND
		(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
		(p.tags @> $5 OR $5 = '{}')
	GROUP BY p.id, u.username
	ORDER BY p.created_at ` + fq.Sort + `
	LIMIT $2 OFFSET $3`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	rows, err := p.db.QueryContext(ctx, query, userID, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	feed := []PostWithMetadata{}
	for rows.Next() {
		post := PostWithMetadata{}
		err := rows.Scan(&post.ID, &post.USERID, &post.Title, &post.Content, pq.Array(&post.Tags), &post.CreatedAt, &post.UpdatedAt, &post.Version, &post.User.Username, &post.CommentCount)
		if err != nil {
			return nil, err
		}
		feed = append(feed, post)
	}
	return feed, nil
}
