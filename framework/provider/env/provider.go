package env

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type HadeEnvProvider struct {
	Folder string
}

func (h HadeEnvProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeEnv
}

func (h HadeEnvProvider) Boot(container framework.Container) error {
	app := container.MustMake(contract.AppKey).(contract.App)
	h.Folder = app.BaseFolder()
	return nil
}

func (h HadeEnvProvider) IsDefer() bool {
	return false
}

func (h HadeEnvProvider) Params(container framework.Container) []interface{} {
	return []interface{}{h.Folder}
}

func (h HadeEnvProvider) Name() string {
	return contract.EnvKey
}
