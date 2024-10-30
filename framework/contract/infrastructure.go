package contract

const InfrastructureKey = "hade:infrastructure"

type InfrastructureService interface {
	// GetModuleOrmRepository 通过模块名称来获取 对应的基础设施 -仓储层实现类
	GetModuleOrmRepository(moduleName string) interface{}
	RegisterOrmRepository(moduleName string, repository interface{})
}
