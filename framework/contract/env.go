package contract

const (
	EnvProduction  = "production"
	EnvTesting     = "testing"
	EnvDevelopment = "development"
	EnvKey         = "hade:env"
	EnvAppKey      = "APP_ENV"
)

type Env interface {
	// AppEnv 获取当前的环境，建议分为development/testing/production
	AppEnv() string
	// IsExist 判断一个环境变量是否有被设置
	IsExist(string) bool
	// Get 获取某个环境变量，如果没有设置，返回""
	Get(string) string
	// All 获取所有的环境变量，.env和运行环境变量融合后结果
	All() map[string]string
}
