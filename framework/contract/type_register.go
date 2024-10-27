package contract

import "reflect"

const TypeRegisterKey = "hade:typeRegister"

type TypeRegisterService interface {
	Register(typeName string, instance interface{})
	GetType(typeName string) (reflect.Type, bool)
}
