package cache

import (
	"context"

	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*storage.User, error)
		Set(context.Context, *storage.User) error
		Delete(context.Context, int64) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStorage{rdb: rdb},
	}

}
