package main

import (
	"context"
	"fmt"
	"github.com/Superdanda/hade/framework/gin"
	"net/http"
	"time"
)

func FooControllerHandler(ctx *gin.Context) error {
	finish := make(chan struct{}, 1)
	panicChan := make(chan interface{}, 1)

	durationCtx, cancel := context.WithTimeout(ctx.BaseContext(), 2*time.Second) // 这里记得当所有事情处理结束后调用 cancel，告知 durationCtx 的后续 Context 结束
	defer cancel()

	go func() {
		time.Sleep(1 * time.Second)
		ctx.ISetOkStatus().IJson(map[string]interface{}{"code": 0})
		finish <- struct{}{}
	}()

	select {
	case <-finish:
		fmt.Println("调用结束了")
	case <-durationCtx.Done():
		ctx.ISetStatus(http.StatusInternalServerError).IJson("time out")
	case <-panicChan:
		ctx.ISetStatus(http.StatusInternalServerError).IJson("panic")
	}
	return nil
}

func UserLoginController(c *gin.Context) {
	foo, _ := c.DefaultQueryString("foo", "def")
	// 等待10s才结束执行
	time.Sleep(10 * time.Second)
	// 输出结果
	c.ISetOkStatus().IJson("ok, UserLoginController: " + foo)
}
