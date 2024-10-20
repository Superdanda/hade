package config

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"path/filepath"
)

type HadeConfigProvider struct{}

// Register registe a new function for make a services instance
func (provider *HadeConfigProvider) Register(c framework.Container) framework.NewInstance {
	return NewHadeConfigService
}

// Boot will called when the services instantiate
func (provider *HadeConfigProvider) Boot(c framework.Container) error {
	return nil
}

// IsDefer define whether the services instantiate when first make or register
func (provider *HadeConfigProvider) IsDefer() bool {
	return false
}

// Params define the necessary params for NewInstance
func (provider *HadeConfigProvider) Params(c framework.Container) []interface{} {
	appService := c.MustMake(contract.AppKey).(contract.App)
	envService := c.MustMake(contract.EnvKey).(contract.Env)
	env := envService.AppEnv()
	// 配置文件夹地址
	configFolder := appService.ConfigFolder()
	envFolder := filepath.Join(configFolder, env)
	return []interface{}{c, envFolder, envService.All()}
}

// / Name define the name for this services
func (provider *HadeConfigProvider) Name() string {
	return contract.ConfigKey
}
