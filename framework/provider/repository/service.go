package repository

import (
	"context"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/pkg/errors"
)

type HadeRepositoryService[T any, ID comparable] struct {
	container    framework.Container
	cacheService contract.CacheService
}

func NewHadeRepositoryService[T any, ID comparable](params ...interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, errors.New("insufficient parameters")
	}
	container, ok := params[0].(framework.Container)
	if !ok {
		return nil, errors.New("invalid container parameter")
	}
	cacheService, ok := params[1].(contract.CacheService)
	if !ok {
		return nil, errors.New("invalid cacheService parameter")
	}
	return &HadeRepositoryService[T, ID]{
		container:    container,
		cacheService: cacheService,
	}, nil
}

func (h *HadeRepositoryService[T, ID]) Save(ctx context.Context, entity *T) error {
	//TODO implement me
	panic("implement me")
}

func (h HadeRepositoryService[T, ID]) FindByID(ctx context.Context, id ID) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (h HadeRepositoryService[T, ID]) FindByField(ctx context.Context, fieldName string, value any) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

func (h HadeRepositoryService[T, ID]) FindByIDs(ctx context.Context, ids []ID) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

func (h HadeRepositoryService[T, ID]) FindByFieldIn(ctx context.Context, fieldName string, values []any) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

type HadeCacheRepository[T any, ID comparable] struct {
	cacheService contract.CacheService
}
