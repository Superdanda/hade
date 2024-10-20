package demo

import (
	"fmt"
	"github.com/Superdanda/hade/framework"
)

// Key Demo 服务的 key
const Key = "hade:demo"

// Service Demo 服务的接口
type Service interface {
	GetFoo() Foo
}

// Foo Demo 服务接口定义的一个数据结构
type Foo struct {
	Name string
}

// DemoService serviceProvider 实现
type DemoService struct {
	Service
	c framework.Container
}

// GetFoo 实现接口
func (s *DemoService) GetFoo() Foo {
	return Foo{Name: "i am foo"}
}

func NewDemoService(params ...interface{}) (interface{}, error) {
	c := params[0].(framework.Container)
	fmt.Println("new demo services")
	return &DemoService{c: c}, nil
}
