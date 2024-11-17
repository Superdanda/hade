package test

import (
	"github.com/Superdanda/hade/app/provider/database_connect"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/provider/app"
	"github.com/Superdanda/hade/framework/provider/cache"
	"github.com/Superdanda/hade/framework/provider/config"
	"github.com/Superdanda/hade/framework/provider/distributed"
	"github.com/Superdanda/hade/framework/provider/env"
	"github.com/Superdanda/hade/framework/provider/kafka"
	"github.com/Superdanda/hade/framework/provider/log"
	"github.com/Superdanda/hade/framework/provider/orm"
	"github.com/Superdanda/hade/framework/provider/queue"
	"github.com/Superdanda/hade/framework/provider/redis"
	"github.com/Superdanda/hade/framework/provider/repository"
	"github.com/Superdanda/hade/framework/provider/ssh"
	"github.com/Superdanda/hade/framework/provider/type_register"
)

const (
	BasePath = "C:\\Users\\a1033\\GolandProjects\\framework1"
)

func InitBaseContainer() framework.Container {
	// 初始化服务容器
	container := framework.NewHadeContainer()
	// 绑定App服务提供者
	container.Bind(&app.HadeAppProvider{BaseFolder: BasePath})
	// 后续初始化需要绑定的服务提供者...
	container.Bind(&env.HadeEnvProvider{})
	container.Bind(&distributed.LocalDistributedProvider{})
	container.Bind(&config.HadeConfigProvider{})
	//container.Bind(&id.HadeIDProvider{})
	//container.Bind(&trace.HadeTraceProvider{})
	container.Bind(&log.HadeLogServiceProvider{})
	container.Bind(&orm.GormProvider{})
	container.Bind(&redis.RedisProvider{})
	container.Bind(&cache.HadeCacheProvider{})
	container.Bind(&ssh.SSHProvider{})
	container.Bind(&type_register.TypeRegisterProvider{})
	//container.Bind(&infrastructure.InfrastructureProvider{})
	container.Bind(&repository.RepositoryProvider{})
	container.Bind(&kafka.KafkaProvider{})
	container.Bind(&queue.QueueProvider{})
	// 将HTTP引擎初始化,并且作为服务提供者绑定到服务容器中
	//if engine, err := http.NewHttpEngine(container); err == nil {
	//	container.Bind(&kernel.HadeKernelProvider{HttpEngine: engine})
	//}

	container.Bind(&database_connect.DatabaseConnectProvider{})
	return container
}
