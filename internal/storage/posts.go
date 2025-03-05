package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	Version   int       `json:"version"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comment"`
	User      User      `json:"user"`
}

type PostForFeed struct {
	Post         Post
	CommentCount int `json:"comment_count"`
}

type PostStorage struct {
	db *sql.DB
}

func (p *PostStorage) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (title, content, user_id, tags) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at;`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := p.db.QueryRowContext(ctx, query, post.Title, post.Content, post.UserID,
		pq.Array(post.Tags)).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostStorage) GetByID(ctx context.Context, postID int64) (*Post, error) {
	var post Post
	query := `SELECT id, title, user_id, content,  tags, created_at, updated_at FROM posts WHERE id = $1;`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := p.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.Title,
		&post.UserID,
		&post.Content,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}

	}

	return &post, nil
}

func (p *PostStorage) Delete(ctx context.Context, postId int64) error {
	query := `DELETE FROM posts WHERE id=$1;`
	//ToDo : delete comments
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	res, err := p.db.ExecContext(ctx, query, postId)
	if err != nil {
		return err
	}
	rowsCount, err := res.RowsAffected()
	if err != nil {

		if rowsCount == 0 {
			return ErrNotFound
		} else if rowsCount != 1 {
			return ErrTooMuchChanged
		}
		return err
	}
	return nil

}

func (p *PostStorage) Update(ctx context.Context, post *Post) error {
	query := `UPDATE posts
		SET title = $1, content = $2, version = version + 1
		WHERE id = $3 AND version=$4 
		RETURNING version`
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
