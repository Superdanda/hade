package database_connect

import "gorm.io/gorm"

const DatabaseConnectKey = "hade:database_connect"

type Service interface {
	DefaultDatabaseConnect() *gorm.DB
	LocalDatabaseConnect() *gorm.DB
	AliDataBaseConnect() *gorm.DB
}
