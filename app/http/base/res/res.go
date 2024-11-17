package res

import (
	"github.com/Superdanda/hade/app/http/base"
	"github.com/Superdanda/hade/framework/gin"
	"net/http"
)

func FailWithErr(ctx *gin.Context, err error) {
	ctx.ISetStatus(http.StatusInternalServerError).IJson(base.Fail(err.Error()))
}

func Fail(ctx *gin.Context) {
	ctx.ISetStatus(http.StatusInternalServerError).IJson(base.Fail("操作失败"))
}

func Success(ctx *gin.Context) {
	ctx.ISetStatus(http.StatusOK).IJson(base.SuccessWithOKMessage())
}

func SuccessWithData(ctx *gin.Context, data interface{}) {
	ctx.ISetStatus(http.StatusOK).IJson(base.Success(data))
}
