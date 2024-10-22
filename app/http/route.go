package http

import (
	"github.com/Superdanda/hade/app/http/module/demo"
	"github.com/Superdanda/hade/framework/gin"
)

func Routes(core *gin.Engine) {
	// /路径先去./dist目录下查找文件是否存在，找到使用文件服务提供服务
	//core.Use(static.Serve("/", static.LocalFile("./dist", false)))

	err := demo.Register(core)
	if err != nil {
		return
	}
}
