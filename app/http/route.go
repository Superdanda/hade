package http

import (
	"github.com/Superdanda/hade/app/http/module/demo"
	"github.com/Superdanda/hade/app/http/module/user"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/gin"
	ginSwagger "github.com/Superdanda/hade/framework/middleware/gin-swagger"
	"github.com/Superdanda/hade/framework/middleware/gin-swagger/swaggerFiles"
	"github.com/Superdanda/hade/framework/middleware/static"
)

func Routes(core *gin.Engine) {
	container := core.GetContainer()
	configService := container.MustMake(contract.ConfigKey).(contract.Config)

	// 如果配置了swagger，则显示swagger的中间件
	if configService.GetBool("app.swagger") == true {
		core.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// /路径先去./dist目录下查找文件是否存在，找到使用文件服务提供服务
	core.Use(static.Serve("/", static.LocalFile("./dist", false)))

	err := demo.Register(core)

	err = user.RegisterRoutes(core)

	if err != nil {
		return
	}
}
