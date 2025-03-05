package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/go-redis/redis/v8"
)

const UserExpTime = time.Minute

type UserStorage struct {
	rdb *redis.Client
}

func (u UserStorage) Get(ctx context.Context, id int64) (*storage.User, error) {
	cacheKey := fmt.Sprintf("user-%d", id)
	data, err := u.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user storage.User
	if data != "" {
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			return nil, err
		}

	}

	return &user, nil
}

func (u UserStorage) Set(ctx context.Context, user *storage.User) error {
	cacheKey := fmt.Sprintf("user-%d", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return u.rdb.SetEX(ctx, cacheKey, json, UserExpTime).Err()
}

func (u UserStorage) Delete(ctx context.Context, userId int64) error {
	cacheKey := fmt.Sprintf("user-%d", userId)
	return u.rdb.Del(ctx, cacheKey).Err()

}
