package cache

import (
	"context"

	"github.com/dunkykorZhik/social/internal/storage"
)

func NewMockStorage() Storage {
	return Storage{
		Users: &UserMockStorage{},
	}
}

type UserMockStorage struct {
}

func (u UserMockStorage) Get(ctx context.Context, id int64) (*storage.User, error) {
	return &storage.User{}, nil
}

func (u UserMockStorage) Set(ctx context.Context, user *storage.User) error {
	return nil
}

func (u UserMockStorage) Delete(ctx context.Context, userId int64) error {
	return nil

}
