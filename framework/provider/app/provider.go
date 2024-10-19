package app

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type HadeAppProvider struct {
	BaseFolder string
}

func (h HadeAppProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeApp
}

func (h HadeAppProvider) Boot(container framework.Container) error {
	return nil
}

func (h HadeAppProvider) IsDefer() bool {
	return false
}

func (h HadeAppProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container, h.BaseFolder}
}

func (h HadeAppProvider) Name() string {
	return contract.AppKey
}
