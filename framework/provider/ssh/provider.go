package ssh

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

// SSHProvider 提供App的具体实现方法
type SSHProvider struct {
}

func (S *SSHProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeSSH
}

func (S *SSHProvider) Boot(container framework.Container) error {
	return nil
}

func (S *SSHProvider) IsDefer() bool {
	return true
}

func (S *SSHProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (S *SSHProvider) Name() string {
	return contract.SSHKey
}
