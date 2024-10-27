package http

import (
	"github.com/Superdanda/hade/app/http/module/demo"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

func TypeRegister(container framework.Container) {
	typeRegister := container.MustMake(contract.TypeRegisterKey).(contract.TypeRegisterService)
	demo.TypeRegister(typeRegister)
}
