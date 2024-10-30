package infrastructure

import (
	"context"
	"github.com/Superdanda/hade/app/provider/database_connect"
	userModule "github.com/Superdanda/hade/app/provider/user"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"gorm.io/gorm"
)

type UserRepository struct {
	container framework.Container
	db        *gorm.DB
	contract.OrmRepository[userModule.User, int64]
	userModule.Repository
}

func NewUserRepository(container framework.Container) contract.OrmRepository[userModule.User, int64] {
	connectService := container.MustMake(database_connect.DatabaseConnectKey).(database_connect.Service)
	connect := connectService.DefaultDatabaseConnect()
	userOrmService := &UserRepository{container: container, db: connect}
	infrastructureService := container.MustMake(contract.InfrastructureKey).(contract.InfrastructureService)
	infrastructureService.RegisterOrmRepository(userModule.UserKey, userOrmService)

	//repository.RegisterRepository[userModule.User, int64](userModule.UserKey,)
	return userOrmService
}

func (u *UserRepository) SaveToDB(entity *userModule.User) error {
	u.db.Save(entity)
	return nil
}

func (u *UserRepository) FindByIDFromDB(id int64) (*userModule.User, error) {
	user := &userModule.User{}
	u.db.Find(user, id)
	return user, nil
}

func (u *UserRepository) FindByID64sFromDB(ids []int64) ([]*userModule.User, error) {
	var users []*userModule.User
	// 使用 GORM 的 Where 方法查询用户 ID 在给定 ID 列表中的记录
	if err := u.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err // 如果查询出错，返回错误
	}
	return users, nil // 返回查询结果和 nil 错误
}

func (u *UserRepository) GetPrimaryKey(entity *userModule.User) int64 {
	return entity.ID
}

func (u *UserRepository) GetBaseField() string {
	return userModule.UserKey
}

func (u *UserRepository) GetFieldQueryFunc(fieldName string) (func(value string) ([]*userModule.User, error), bool) {
	switch fieldName {
	case "Email":
		return func(value string) ([]*userModule.User, error) {
			var users []*userModule.User
			// 执行查询，匹配 Email 字段
			if err := u.db.Where("email = ?", value).Find(&users).Error; err != nil {
				return nil, err
			}
			return users, nil
		}, true
	case "UserName":
		return func(value string) ([]*userModule.User, error) {
			var users []*userModule.User
			// 执行查询，匹配 UserName 字段
			if err := u.db.Where("user_name = ?", value).Find(&users).Error; err != nil {
				return nil, err
			}
			return users, nil
		}, true
	default:
		// 如果传入的字段名不支持，返回 nil 和 false
		return nil, false
	}
}

func (u *UserRepository) GetFieldInQueryFunc(fieldName string) (func(values []string) ([]*userModule.User, error), bool) {
	switch fieldName {
	case "Email":
		return func(values []string) ([]*userModule.User, error) {
			var users []*userModule.User
			// 批量查询 Email 字段匹配的用户
			if err := u.db.Where("email IN ?", values).Find(&users).Error; err != nil {
				return nil, err
			}
			return users, nil
		}, true

	case "UserName":
		return func(values []string) ([]*userModule.User, error) {
			var users []*userModule.User
			// 批量查询 UserName 字段匹配的用户
			if err := u.db.Where("user_name IN ?", values).Find(&users).Error; err != nil {
				return nil, err
			}
			return users, nil
		}, true

	default:
		return nil, false // 不支持的字段返回 false
	}
}

func (u *UserRepository) GetFieldValueFunc(fieldName string) (func(entity *userModule.User) string, bool) {
	switch fieldName {
	case "Email":
		return func(entity *userModule.User) string {
			return entity.Email
		}, true

	case "UserName":
		return func(entity *userModule.User) string {
			return entity.UserName
		}, true

	default:
		return nil, false // 不支持的字段返回 false
	}
}

func (u *UserRepository) Save(ctx context.Context, user *userModule.User) error {
	repository := u.container.MustMake(contract.RepositoryKey).(contract.Repository[userModule.User, int64])
	if err := repository.Save(ctx, userModule.UserKey, user); err != nil {
		return err
	}
	return nil
}

func (u *UserRepository) FindById(ctx context.Context, id int64) (*userModule.User, error) {
	repository := u.container.MustMake(contract.RepositoryKey).(contract.Repository[userModule.User, int64])
	byID, err := repository.FindByID(ctx, userModule.UserKey, id)
	if err != nil {
		return nil, err
	}
	return byID, nil
}
