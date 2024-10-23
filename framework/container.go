package framework

import (
	"errors"
	"fmt"
	"sync"
)

type Container interface {
	// Bind 绑定一个服务提供者，如果关键字凭证已经存在，会进行替换操作，返回 error
	Bind(provider ServiceProvider) error

	// IsBind 关键字凭证是否已经绑定服务提供者
	IsBind(key string) bool

	// Make 根据关键字凭证获取一个服务，
	Make(key string) (interface{}, error)

	// MustMake 根据关键字凭证获取一个服务，如果这个关键字凭证未绑定服务提供者，那么会 panic。
	//所以在使用这个接口的时候请保证服务容器已经为这个关键字凭证绑定了服务提供者。
	MustMake(key string) interface{}

	// MakeNew 根据关键字凭证获取一个服务，只是这个服务并不是单例模式的
	//它是根据服务提供者注册的启动函数和传递的 params 参数实例化出来的
	//这个函数在需要为不同参数启动不同实例的时候非常有用
	MakeNew(key string, params []interface{}) (interface{}, error)

	//NameList 返回所有提供服务者的字符串凭证
	NameList() []string
}

type HadeContainer struct {
	Container
	// providers 存储注册的服务提供者，key 为字符串凭证
	providers map[string]ServiceProvider
	// instance 存储具体的实例，key 为字符串凭证
	instances map[string]interface{}
	// lock 用于锁住对容器的变更操作
	lock sync.RWMutex
}

func NewHadeContainer() *HadeContainer {
	return &HadeContainer{
		providers: make(map[string]ServiceProvider),
		instances: make(map[string]interface{}),
		lock:      sync.RWMutex{},
	}
}

// PrintProviders 输出服务容器中注册的关键字
func (h *HadeContainer) PrintProviders() []string {
	ret := []string{}
	for _, provider := range h.providers {
		name := provider.Name()

		line := fmt.Sprint(name)
		ret = append(ret, line)
	}
	return ret
}

func (h *HadeContainer) Bind(provider ServiceProvider) error {
	key := provider.Name()
	h.providers[key] = provider
	if !provider.IsDefer() {
		_, err := h.newInstance(key, provider, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *HadeContainer) IsBind(key string) bool {
	_, ok := h.instances[key]
	return ok
}

func (h *HadeContainer) Make(key string) (interface{}, error) {
	return h.make(key, nil, false)
}

func (h *HadeContainer) MustMake(key string) interface{} {
	instance, err := h.make(key, nil, false)
	if err != nil {
		panic(err)
	}
	return instance
}

func (h *HadeContainer) MakeNew(key string, params []interface{}) (interface{}, error) {
	return h.make(key, params, true)
}

func (h *HadeContainer) NameList() []string {
	ret := []string{}
	for _, provider := range h.providers {
		name := provider.Name()
		ret = append(ret, name)
	}
	return ret
}

func (h *HadeContainer) make(key string, params []interface{}, forceNew bool) (interface{}, error) {
	provider, err := h.findServiceProvider(key)
	if err != nil {
		return nil, err
	}

	//强制重新实例化
	if forceNew {
		return h.newInstance(key, provider, params)
	}

	// 不需要强制重新实例化，如果容器中已经实例化了，那么就直接使用容器中的实例
	if instance, ok := h.instances[key]; ok {
		return instance, nil
	}

	// 容器中还未实例化，则进行一次实例化
	instance, err := h.newInstance(key, provider, params)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (h *HadeContainer) findServiceProvider(key string) (ServiceProvider, error) {
	provider := h.providers[key]
	if provider == nil {
		return nil, errors.New("no such provider: " + key)
	}
	return provider, nil
}

func (h *HadeContainer) newInstance(key string, provider ServiceProvider, params []interface{}) (interface{}, error) {
	//因为要对容器进行更改，先使用读写锁避免并发操作
	h.lock.Lock()
	defer h.lock.Unlock()
	// force new a
	if err := provider.Boot(h); err != nil {
		return nil, err
	}
	if params == nil {
		params = provider.Params(h)
	}
	register := provider.Register(h)
	instance, err := register(params...)
	if err != nil {
		return nil, err
	}
	h.instances[key] = instance
	return instance, nil
}
