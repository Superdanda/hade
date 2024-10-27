package database_connect

import "gorm.io/gorm"

const DatabaseConnectKey = "hade:database_connect"

type Service interface {
	LocalDatabaseConnect() *gorm.DB
	AliDataBaseConnect() *gorm.DB
}
