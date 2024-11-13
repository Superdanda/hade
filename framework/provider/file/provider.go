package file

import (
	"github.com/Superdanda/hade/framework"
)

type FileProvider struct{}

func (f FileProvider) Register(container framework.Container) framework.NewInstance {
	//TODO implement me
	panic("implement me")
}

func (f FileProvider) Boot(container framework.Container) error {
	//TODO implement me
	panic("implement me")
}

func (f FileProvider) IsDefer() bool {
	//TODO implement me
	panic("implement me")
}

func (f FileProvider) Params(container framework.Container) []interface{} {
	//TODO implement me
	panic("implement me")
}

func (f FileProvider) Name() string {
	//TODO implement me
	panic("implement me")
}
