package id

import "github.com/Superdanda/hade/framework"

type HadeIDProvider struct {
}

func (h HadeIDProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeIDService
}

func (h HadeIDProvider) Boot(container framework.Container) error {
	return nil
}

func (h HadeIDProvider) IsDefer() bool {
	return false
}

func (h HadeIDProvider) Params(container framework.Container) []interface{} {
	return []interface{}{}
}

func (h HadeIDProvider) Name() string {
	return contract.IDKey
}