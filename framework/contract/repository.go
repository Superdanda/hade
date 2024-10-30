package contract

import (
	"context"
	"github.com/Superdanda/hade/framework"
)

const RepositoryKey = "hade:repository"

type RepositoryService interface {
	GetGenericRepositoryByKey(key string) interface{}
	GetGenericRepositoryMap() map[string]interface{}
	GetContainer() framework.Container
}

type GenericRepository[T any, ID comparable] interface {
	Save(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id ID) (*T, error)
	FindByField(ctx context.Context, fieldName string, value string) ([]*T, error)
	FindByIDs(ctx context.Context, ids []ID) ([]*T, error)
	FindByFieldIn(ctx context.Context, fieldName string, values []string) ([]*T, error)
}

type OrmRepository[T any, ID comparable] interface {
	SaveToDB(entity *T) error
	FindByIDFromDB(id ID) (*T, error)
	FindByIDsFromDB(ids []ID) ([]*T, error)
	GetPrimaryKey(entity *T) ID
	GetBaseField() string
	GetFieldQueryFunc(fieldName string) (func(value string) ([]*T, error), bool)
	GetFieldInQueryFunc(fieldName string) (func(values []string) ([]*T, error), bool)
	GetFieldValueFunc(fieldName string) (func(entity *T) string, bool)
}
