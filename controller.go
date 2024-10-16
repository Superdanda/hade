package main

import (
	"context"
	"fmt"
	"framework1/framework"
	"time"
)

func FooControllerHandler(ctx *framework.Context) error {
	finish := make(chan struct{}, 1)
	panicChan := make(chan interface{}, 1)

	durationCtx, cancel := context.WithTimeout(ctx.BaseContext(), 2*time.Second) // 这里记得当所有事情处理结束后调用 cancel，告知 durationCtx 的后续 Context 结束
	defer cancel()

	go func() {
		time.Sleep(1 * time.Second)
		ctx.Json(200, map[string]interface{}{"code": 0})
		finish <- struct{}{}
	}()

	select {
	case <-finish:
		fmt.Println("调用结束了")
	case <-durationCtx.Done():
		ctx.Json(500, "time out")
	case <-panicChan:
		ctx.Json(500, "panic")
	}
	return nil
}

func UserLoginController(c *framework.Context) error {
	c.Json(200, "ok UserLoginController")
	return nil
}

func SubjectDelController(c *framework.Context) error {
	c.Json(200, "ok SubjectDelController")
	return nil
}

func SubjectUpdateController(c *framework.Context) error {
	c.Json(200, "ok SubjectUpdateController")
	return nil
}

func SubjectGetController(c *framework.Context) error {
	c.Json(200, "ok SubjectGetController")
	return nil
}

func SubjectListController(c *framework.Context) error {
	c.Json(200, "ok SubjectListController")
	return nil
}
