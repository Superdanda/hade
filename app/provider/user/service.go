package user

import (
	"context"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type UserService struct {
	container  framework.Container
	repository Repository
}

func (s *UserService) GetUser(ctx context.Context, userID int64) (*User, error) {
	user, err := s.repository.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) SaveUser(ctx context.Context, user *User) error {
	err := s.repository.Save(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func NewUserService(params ...interface{}) (interface{}, error) {
	container := params[0].(framework.Container)
	infrastructureService := container.MustMake(contract.InfrastructureKey).(contract.InfrastructureService)
	ormRepository := infrastructureService.GetModuleOrmRepository(UserKey).(Repository)
	return &UserService{container: container, repository: ormRepository}, nil
}

func (s *UserService) Foo() string {
	return ""
}
