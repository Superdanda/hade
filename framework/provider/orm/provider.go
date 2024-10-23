package orm

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type GormProvider struct {
}

func (g GormProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeGorm
}

func (g GormProvider) Boot(container framework.Container) error {
	return nil
}

func (g GormProvider) IsDefer() bool {
	return true
}

func (g GormProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (g GormProvider) Name() string {
	return contract.ORMKey
}
