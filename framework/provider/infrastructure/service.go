package infrastructure

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	_ "github.com/Superdanda/hade/framework/provider/repository"
)

type Service struct {
	container framework.Container
	contract.InfrastructureService
	ormRepositoryMap map[string]interface{}
}

func NewInfrastructureService(params ...interface{}) (interface{}, error) {
	return &Service{container: params[0].(framework.Container)}, nil
}

func (i *Service) GetModuleOrmRepository(moduleName string) interface{} {
	return i.ormRepositoryMap[moduleName]
}

func (i *Service) RegisterOrmRepository(moduleName string, repository interface{}) {
	i.ormRepositoryMap[moduleName] = repository
}
