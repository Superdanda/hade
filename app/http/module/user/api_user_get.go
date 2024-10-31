package user

import (
	"github.com/Superdanda/hade/app/http/result"
	"github.com/Superdanda/hade/app/provider/user"
	"github.com/Superdanda/hade/framework/gin"
	"github.com/spf13/cast"
	"net/http"
)

type UserGetParam struct {
	ID string `json:"id"`
}

func (api *UserApi) UserGet(c *gin.Context) {
	userService := c.MustMake(user.UserKey).(user.Service)
	param := &UserGetParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		c.ISetStatus(http.StatusBadRequest).IJson(result.Fail("参数错误"))
		return
	}

	userQuery, err := userService.GetUser(c, cast.ToInt64(param.ID))
	if err != nil {
		c.ISetStatus(http.StatusInternalServerError).IJson(result.Fail("网络开小差"))
		return
	}
	c.ISetOkStatus().IJson(result.Success(userQuery))
}
