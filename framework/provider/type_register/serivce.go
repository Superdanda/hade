package type_register

import (
	"reflect"
	"sync"
)

type HadeTypeRegisterService struct {
	mu    sync.RWMutex
	types map[string]reflect.Type
}

func NewHadeTypeRegisterService(params ...interface{}) (interface{}, error) {
	return &HadeTypeRegisterService{
		types: make(map[string]reflect.Type),
	}, nil
}

func (r *HadeTypeRegisterService) Register(typeName string, instance interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[typeName] = reflect.TypeOf(instance)
}

func (r *HadeTypeRegisterService) GetType(typeName string) (reflect.Type, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	typ, exists := r.types[typeName]
	return typ, exists
}
