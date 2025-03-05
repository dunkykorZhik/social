package storage

import (
	"context"
	"time"
)

func NewMockStorage() Storage {
	return Storage{
		Users: &UserMockStorage{},
	}
}

type UserMockStorage struct {
}

func (u *UserMockStorage) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}

func (u *UserMockStorage) Activate(ctx context.Context, token string) error {
	return nil
}

func (u *UserMockStorage) Delete(ctx context.Context, id int64) error {
	return nil
}

func (u *UserMockStorage) GetByID(ctx context.Context, userId int64) (*User, error) {

	return &User{}, nil
}

func (u *UserMockStorage) GetByEmail(ctx context.Context, email string) (*User, error) {

	return &User{}, nil
}

func (u *UserMockStorage) Follow(ctx context.Context, fId int64, userId int64) error {

	return nil

}

func (u *UserMockStorage) UnFollow(ctx context.Context, fId int64, userId int64) error {

	return nil

}

func (u *UserMockStorage) GetUserFeed(ctx context.Context, user_id int64, pagQ PaginateQuery) ([]PostForFeed, error) {

	return nil, nil
}
