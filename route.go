package main

import (
	"github.com/echo/hade/framework/gin"
)

func registerRouter(core *gin.Engine) {
	core.Use(Test2(), Test3())
	core.GET("/user/login", Test1(), UserLoginController)

	subjectGroup := core.Group("/subject")
	{
		parentGroup := subjectGroup.Group("/parent")
		parentGroup.Use(Test1())
		{
			parentGroup.DELETE("/:id", SubjectDelController)
			parentGroup.PUT("/:id", SubjectUpdateController)
			parentGroup.GET("/:id", SubjectUpdateController)
			parentGroup.GET("/list/all", SubjectListController)
		}
	}
}
