package main

import "github.com/echo/hade/framework/gin"

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
	c.ISetOkStatus().IJson("ok SubjectListController")
}
