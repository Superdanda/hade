package gin

import (
	"context"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

func (ctx *Context) BaseContext() context.Context {
	return ctx.Request.Context()
}

// Bind engine 实现 container 的绑定封装
func (engine *Engine) Bind(provider framework.ServiceProvider) error {
	return engine.container.Bind(provider)
}

// IsBind 关键字凭证是否已经绑定服务提供者
func (engine *Engine) IsBind(key string) bool {
	return engine.container.IsBind(key)
}

// GetContainer 获取服务提供者容器
func (engine *Engine) GetContainer() framework.Container {
	return engine.container
}

// Make context 实现 container 的几个封装
// 实现 make 的封装
func (ctx *Context) Make(key string) (interface{}, error) {
	return ctx.container.Make(key)
}

// MustMake 实现 mustMake 的封装
func (ctx *Context) MustMake(key string) interface{} {
	return ctx.container.MustMake(key)
}

// MustMakeLog 快速获取Log服务
func (ctx *Context) MustMakeLog() contract.Log {
	return ctx.container.MustMake(contract.LogKey).(contract.Log)
}

// MakeNew 实现 makenew 的封装
func (ctx *Context) MakeNew(key string, params []interface{}) (interface{}, error) {
	return ctx.container.MakeNew(key, params)
}

// SetContainer 设置服务提供者容器
func (engine *Engine) SetContainer(container framework.Container) {
	engine.container = container
}
