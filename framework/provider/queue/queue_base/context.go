package queue_base

import (
	"context"
	"sync"
	"time"

	"github.com/Superdanda/hade/framework"
)

// Context 是消息队列或异步执行任务的上下文封装
type Context struct {
	Ctx       context.Context     // 标准上下文（用于 timeout、cancel）
	Container framework.Container // DI 容器（用于解析 service、config 等）
	Identity  AuthIdentity        // 当前登录用户信息（可选）
	Metadata  map[string]any      // 请求级元数据（结构化日志/链路 ID 等）
	mu        sync.RWMutex        // 锁保护 Metadata
}

func NewContext(ctx context.Context, container framework.Container, identity AuthIdentity) *Context {
	return &Context{
		Ctx:       ctx,
		Container: container,
		Identity:  identity,
		Metadata:  make(map[string]any),
	}
}

// WithValue 设置 Metadata 中的字段（线程安全）
func (c *Context) WithValue(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Metadata[key] = value
}

// Value 获取 Metadata 或嵌套 context 中的字段（标准 context.Context 接口）
func (c *Context) Value(key any) any {
	if strKey, ok := key.(string); ok {
		c.mu.RLock()
		defer c.mu.RUnlock()
		if val, ok := c.Metadata[strKey]; ok {
			return val
		}
	}
	return c.Ctx.Value(key)
}

func (c *Context) Deadline() (time.Time, bool) {
	return c.Ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}

func (c *Context) Err() error {
	return c.Ctx.Err()
}
