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

func (s *UserService) AddAmount(ctx context.Context, userID int64, amount int64) error {
	// 使用kafka 来发布事件
	return nil
}

//func (s *UserService) ChangeAmount(container  framework.Container, ctx *gin.Context) error {
//	queueService := s.container.MustMake(contract.QueueKey).(contract.QueueService)
//	queueService.SubscribeEvent(ctx,ChangeAmountTopic, )
//}

const ChangeAmountTopic = "UserAmountChange"

type ChangeAmountEvent struct {
	UserID int64 `json:"user_id"`
	Amount int64 `json:"amount"`
}

func NewUserService(params ...interface{}) (interface{}, error) {
	container := params[0].(framework.Container)
	userService := &UserService{container: container}
	infrastructureService := container.MustMake(contract.InfrastructureKey).(contract.InfrastructureService)
	ormRepository := infrastructureService.GetModuleOrmRepository(UserKey).(Repository)
	userService.repository = ormRepository
	return userService, nil
}

func (s *UserService) Foo() string {
	return ""
}
