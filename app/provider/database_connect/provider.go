package database_connect

import (
	"github.com/Superdanda/hade/framework"
)

type DatabaseConnectProvider struct {
	framework.ServiceProvider

	c framework.Container
}

func (sp *DatabaseConnectProvider) Name() string {
	return DatabaseConnectKey
}

func (sp *DatabaseConnectProvider) Register(c framework.Container) framework.NewInstance {
	return NewDatabaseConnectService
}

func (sp *DatabaseConnectProvider) IsDefer() bool {
	return false
}

func (sp *DatabaseConnectProvider) Params(c framework.Container) []interface{} {
	return []interface{}{c}
}

func (sp *DatabaseConnectProvider) Boot(c framework.Container) error {
	return nil
}
