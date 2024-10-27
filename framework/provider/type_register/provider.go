package type_register

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type TypeRegisterProvider struct {
}

func (t TypeRegisterProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeTypeRegisterService
}

func (t TypeRegisterProvider) Boot(container framework.Container) error {
	return nil
}

func (t TypeRegisterProvider) IsDefer() bool {
	return false
}

func (t TypeRegisterProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (t TypeRegisterProvider) Name() string {
	return contract.TypeRegisterKey
}
