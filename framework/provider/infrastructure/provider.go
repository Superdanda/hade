package infrastructure

import "github.com/Superdanda/hade/framework"

type InfrastructureProvider struct {
}

const InfrastructureKey = "hade:infrastructure"

func (i *InfrastructureProvider) Register(container framework.Container) framework.NewInstance {
	return NewInfrastructureService
}

func (i *InfrastructureProvider) Boot(container framework.Container) error {
	return nil
}

func (i *InfrastructureProvider) IsDefer() bool {
	return false
}

func (i *InfrastructureProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (i *InfrastructureProvider) Name() string {
	return InfrastructureKey
}
