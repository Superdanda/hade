package test

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/provider/app"
	"github.com/Superdanda/hade/framework/provider/env"
)

const (
	BasePath = "C:\\Users\\lulz1\\GolandProjects\\framework1"
)

func InitBaseContainer() framework.Container {
	// 初始化服务容器
	container := framework.NewHadeContainer()
	// 绑定App服务提供者
	container.Bind(&app.HadeAppProvider{BaseFolder: BasePath})
	// 后续初始化需要绑定的服务提供者...
	container.Bind(&env.HadeTestingEnvProvider{})
	return container
}
