package main

import "framework1/framework"

func registerRouter(core *framework.Core) {
	core.Get("/user/login", UserLoginController)

	subjectGroup := core.Group("/subject")
	{
		subjectGroup.Delete("/:id", SubjectDelController)
		subjectGroup.Put("/:id", SubjectUpdateController)
		subjectGroup.Get("/:id", SubjectUpdateController)
		subjectGroup.Get("/list/all", SubjectListController)
	}
}
