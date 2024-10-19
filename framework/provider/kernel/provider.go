package kernel

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/gin"
)

type HadeKernelProvider struct {
	HttpEngine *gin.Engine
}

func NewHadeKernelProvider(params ...interface{}) (interface{}, error) {
	return &HadeKernelProvider{}, nil
}

func (h HadeKernelProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeKernelService
}

func (h HadeKernelProvider) Boot(container framework.Container) error {
	if h.HttpEngine == nil {
		h.HttpEngine = gin.Default()
	}
	h.HttpEngine.SetContainer(container)
	return nil
}

func (h HadeKernelProvider) IsDefer() bool {
	return false
}

func (h HadeKernelProvider) Params(container framework.Container) []interface{} {
	return []interface{}{h.HttpEngine}
}

func (h HadeKernelProvider) Name() string {
	return contract.KernelKey
}
