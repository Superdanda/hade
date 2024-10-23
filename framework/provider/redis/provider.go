package redis

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type RedisProvider struct {
}

func (r RedisProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeRedis
}

func (r RedisProvider) Boot(container framework.Container) error {
	return nil
}

func (r RedisProvider) IsDefer() bool {
	return true
}

func (r RedisProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (r RedisProvider) Name() string {
	return contract.RedisKey
}
