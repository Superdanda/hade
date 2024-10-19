package http

import (
	"github.com/Superdanda/hade/framework/gin"
)

func Routes(core *gin.Engine) {
	core.Static("/dist/", "./dist/")
}
