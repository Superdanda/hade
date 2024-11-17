package http

import (
	"github.com/Superdanda/hade/app/infrastructure"
	"github.com/Superdanda/hade/app/provider/database_connect"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/gin"
)

// NewHttpEngine 创建了一个绑定了路由的Web引擎
func NewHttpEngine(container framework.Container) (*gin.Engine, error) {
	// 设置为Release，为的是默认在启动中不输出调试信息
	gin.SetMode(gin.ReleaseMode)
	// 默认启动一个Web引擎
	r := gin.Default()
	r.SetContainer(container)

	// 返回绑定路由后的Web引擎

	// 对业务模型进行注册，通过注册名获取业务模型类型信息
	TypeRegister(container)

	//绑定服务
	container.Bind(&database_connect.DatabaseConnectProvider{})

	//注册 infrastructure 包的实例
	infrastructure.NewOrmRepositoryAndRegister(container)

	// 业务绑定路由操作
	Routes(r)

	//注册消息队列事件
	SubscribeEvent(container)
	return r, nil
}
