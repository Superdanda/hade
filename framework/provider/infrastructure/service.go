package infrastructure

import (
	"github.com/Superdanda/hade/framework"
	_ "github.com/Superdanda/hade/framework/provider/repository"
)

type Service struct {
	container        framework.Container
	ormRepositoryMap map[string]interface{}
}

func NewInfrastructureService(params ...interface{}) (interface{}, error) {
	infrastructureService := &Service{container: params[0].(framework.Container),
		ormRepositoryMap: make(map[string]interface{})}
	return infrastructureService, nil
}

func (i *Service) GetModuleOrmRepository(moduleName string) interface{} {
	return i.ormRepositoryMap[moduleName]
}

func (i *Service) RegisterOrmRepository(moduleName string, repository interface{}) {
	i.ormRepositoryMap[moduleName] = repository
}
