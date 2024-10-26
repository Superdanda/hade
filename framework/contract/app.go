package contract

// AppKey 定义字符串凭证
const AppKey = "hade:app"

// App定义接口
type App interface {
	// Version 定义当前版本
	Version() string
	//BaseFolder 定义项目基础地址
	BaseFolder() string
	// ConfigFolder 定义了配置文件的路径
	ConfigFolder() string
	// LogFolder 定义了日志所在路径
	LogFolder() string
	// ProviderFolder 定义业务自己的服务提供者地址
	ProviderFolder() string
	// MiddlewareFolder 定义业务自己定义的中间件
	MiddlewareFolder() string
	// CommandFolder 定义业务定义的命令
	CommandFolder() string
	// RuntimeFolder 定义业务的运行中间态信息
	RuntimeFolder() string
	// TestFolder 存放测试所需要的信息
	TestFolder() string
	// StorageFolder 存放本地文件
	StorageFolder() string
	// HttpFolder 存放api
	HttpFolder() string
	// ConsoleFolder 存放命令行
	ConsoleFolder() string
	// AppID 表示当前这个app的唯一id, 可以用于分布式锁等
	AppId() string

	// HttpModuleFolder  表示业务层的模块路径
	HttpModuleFolder() string

	LoadAppConfig(mapString map[string]string)

	AppFolder() string
	// DeployFolder 部署文件夹
	DeployFolder() string
}
