package repository

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type RepositoryProvider struct{}

func (r RepositoryProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeRepositoryService
}

func (r RepositoryProvider) Boot(container framework.Container) error {
	return nil
}

func (r RepositoryProvider) IsDefer() bool {
	return true
}

func (r RepositoryProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (r RepositoryProvider) Name() string {
	return contract.RepositoryKey
}
