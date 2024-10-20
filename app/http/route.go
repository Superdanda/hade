package http

import (
	"github.com/Superdanda/hade/app/http/module/demo"
	"github.com/Superdanda/hade/framework/gin"
)

func Routes(core *gin.Engine) {
	core.Static("/dist/", "./dist/")

	err := demo.Register(core)
	if err != nil {
		return
	}
}
