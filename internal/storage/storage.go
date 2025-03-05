package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrTooMuchChanged    = errors.New("the request changed more than expected")
	QueryTimeoutDuration = time.Second * 5
	ErrConflict          = errors.New("conflict between resources")
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
	}
	Users interface {
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error

		GetByID(context.Context, int64) (*User, error)
		GetByEmail(context.Context, string) (*User, error)

		Follow(context.Context, int64, int64) error
		UnFollow(context.Context, int64, int64) error

		GetUserFeed(context.Context, int64, PaginateQuery) ([]PostForFeed, error)
	}
	Comments interface {
		GetByPostID(context.Context, int64) ([]Comment, error)
	}
	Roles interface {
		GetByName(context.Context, string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStorage{db},
		Users:    &UserStorage{db},
		Comments: &CommentStorage{db},
		Roles:    &RoleStorage{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()

	// Read more about transaction, commit and rollback
}
