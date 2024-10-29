package contract

import "context"

const RepositoryKey = "hade:repository"

type Repository[T any, ID comparable] interface {
	Save(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id ID) (*T, error)
	FindByField(ctx context.Context, fieldName string, value any) ([]*T, error)
	FindByIDs(ctx context.Context, ids []ID) ([]*T, error)
	FindByFieldIn(ctx context.Context, fieldName string, values []any) ([]*T, error)
}
