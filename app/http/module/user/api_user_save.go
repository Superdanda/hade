package user

import (
	"github.com/Superdanda/hade/app/http/result"
	"github.com/Superdanda/hade/app/provider/database_connect"
	"github.com/Superdanda/hade/app/provider/user"
	"github.com/Superdanda/hade/framework/gin"
	"net/http"
)

type UserSaveParam struct {
	ID       int64  `gorm:"column:id;primary_key;auto_increment" json:"id"` // 代表用户id, 只有注册成功之后才有这个id，唯一表示一个用户
	UserName string `gorm:"column:username;type:varchar(255);comment:用户名;not null" json:"username"`
	NickName string `gorm:"column:username;type:varchar(255);comment:昵称;not null" json:"nickname"`
	Email    string `gorm:"column:email;type:varchar(255);comment:邮箱;not null" json:"email"`
}

func (u *UserSaveParam) conventUser() *user.User {
	return &user.User{
		ID:       u.ID,
		UserName: u.UserName,
		NickName: u.NickName,
		Email:    u.Email,
	}
}

func (api *UserApi) UserSave(c *gin.Context) {

	service := c.MustMake(database_connect.DatabaseConnectKey).(database_connect.Service)
	service.DefaultDatabaseConnect().AutoMigrate(&user.User{})

	userService := c.MustMake(user.UserKey).(user.Service)
	param := &UserSaveParam{}
	err := c.ShouldBindJSON(param)
	if err != nil {
		c.ISetStatus(http.StatusBadRequest).IJson(result.Fail("参数错误"))
	}
	conventUser := param.conventUser()
	err = userService.SaveUser(c, conventUser)
	if err != nil {
		c.ISetStatus(http.StatusInternalServerError).IJson(result.Fail("网络开小差"))
	}
	c.ISetOkStatus().IJson(result.SuccessWithOKMessage())
}
