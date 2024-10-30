package user

import "context"

type Repository interface {
	Save(ctx context.Context, user *User) error
	FindById(ctx context.Context, id int64) (*User, error)
}
