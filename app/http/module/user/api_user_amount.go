package user

import (
	"github.com/Superdanda/hade/app/provider/user"
	"github.com/Superdanda/hade/app/utils"
	"github.com/Superdanda/hade/framework/gin"
	"net/http"
)

type ChangeAmountParam struct {
	UserId int64 `json:"user_id"`
	Amount int64 `json:"amount"`
}

// ChangeAmount 更改金额
// @Summary 更改金额
// @Description 更改金额
// @ID ChangeAmount
// @Tags ChangeAmount
// @Accept json
// @Produce json
// @Param ChangeAmountParam body ChangeAmountParam true "查询详情请求参数"
// @Success 200 {object} base.Result "返回成功的流程定义数据"
// @Failure 500 {object} base.Result "操作失败"
// @Router /user/amount [post]
func (api *UserApi) ChangeAmount(context *gin.Context) {

	param := utils.QuickBind[ChangeAmountParam](context)

	userService := context.MustMake(user.UserKey).(user.Service)
	err := userService.AddAmount(context, param.UserId, param.Amount)
	if err != nil {
		return
	}
	context.JSON(http.StatusOK, gin.H{"code": 0, "data": nil})
}
