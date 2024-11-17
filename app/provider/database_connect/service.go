package database_connect

import (
	"fmt"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/provider/orm"
	"gorm.io/gorm"
)

type DatabaseConnectService struct {
	container framework.Container
}

func (d DatabaseConnectService) LocalDatabaseConnect() *gorm.DB {
	return getDatabaseConnectByYaml("database.local", d)
}

func (d DatabaseConnectService) AliDataBaseConnect() *gorm.DB {
	return getDatabaseConnectByYaml("database.ali", d)
}

func (d DatabaseConnectService) DefaultDatabaseConnect() *gorm.DB {
	return getDatabaseConnectByYaml("database.local", d)
}

func NewDatabaseConnectService(params ...interface{}) (interface{}, error) {
	container := params[0].(framework.Container)
	return &DatabaseConnectService{container: container}, nil
}

func getDatabaseConnectByYaml(yamlPath string, d DatabaseConnectService) *gorm.DB {
	ormService := d.container.MustMake(contract.ORMKey).(contract.ORMService)
	db, err := ormService.GetDB(orm.WithConfigPath(yamlPath))
	if err != nil {
		fmt.Println(yamlPath + "数据库连接失败，请检查配置")
		return nil
	}
	return db
}
