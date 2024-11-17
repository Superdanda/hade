package user

import (
	"github.com/Superdanda/hade/app/provider/database_connect"
	tests "github.com/Superdanda/hade/test"

	"testing"
)

func TestLoadModel(test *testing.T) {
	container := tests.InitBaseContainer()
	databaseConnectService := container.MustMake(database_connect.DatabaseConnectKey).(database_connect.Service)

	db := databaseConnectService.DefaultDatabaseConnect()
	db.AutoMigrate(User{}, Account{})

}
