package main

import (
	"github.com/Superdanda/hade/app/provider/demo"
	"github.com/Superdanda/hade/framework/gin"
)

func SubjectDelController(c *gin.Context) {
	c.ISetOkStatus().IJson("ok SubjectDelController")
}

func SubjectUpdateController(c *gin.Context) {
	c.ISetOkStatus().IJson("ok SubjectUpdateController")
}

func SubjectGetController(c *gin.Context) {
	c.ISetOkStatus().IJson("ok SubjectGetController")
}

func SubjectListController(c *gin.Context) {
	demoService := c.MustMake(demo.Key).(demo.Service)
	foo := demoService.GetFoo()
	c.ISetOkStatus().IJson(foo)
}
