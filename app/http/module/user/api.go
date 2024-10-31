package user

import (
	"github.com/Superdanda/hade/app/provider/user"
	"github.com/Superdanda/hade/framework/gin"
)

type UserApi struct{}

// 注册路由
func RegisterRoutes(r *gin.Engine) error {

	api := UserApi{}

	if !r.IsBind(user.UserKey) {
		r.Bind(&user.UserProvider{})
	}

	Group := r.Group("/")
	{

		userGroup := Group.Group("/user")
		{

			userGroup.POST("/login", api.UserLogin)
			userGroup.POST("/get", api.UserGet)
			userGroup.POST("/save", api.UserSave)
		}
	}

	return nil
}
