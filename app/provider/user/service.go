package user

import (
	"context"
	"encoding/json"
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
	//  来发布事件
	queueService := s.container.MustMake(contract.QueueKey).(contract.QueueService)
	return queueService.NewEventAndPublish(ctx, ChangeAmountTopic, NewChangeAmountEvent(userID, amount))
}

func (s *UserService) ChangeAmount(ctx context.Context, event contract.Event) error {
	amountEvent, err := contract.GetPayload[ChangeAmountEvent](event)
	if err != nil {
		return err
	}
	user, err := s.GetUser(ctx, amountEvent.UserID)
	if err != nil {
		return err
	}
	if user.Account == nil {
		user.Account = &Account{}
	}
	user.Account.Amount += amountEvent.Amount
	return s.repository.Save(ctx, user)
}

type ChangeAmountEvent struct {
	Topic  string `json:"topic"`
	UserID int64  `json:"user_id"`
	Amount int64  `json:"amount"`
}

func (e *ChangeAmountEvent) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ChangeAmountEvent) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

func NewChangeAmountEvent(userID int64, amount int64) *ChangeAmountEvent {
	return &ChangeAmountEvent{
		Topic:  ChangeAmountTopic,
		UserID: userID,
		Amount: amount,
	}
}

func NewUserService(params ...interface{}) (interface{}, error) {
	container := params[0].(framework.Container)
	userService := &UserService{container: container}
	//infrastructureService := container.MustMake(contract.InfrastructureKey).(contract.InfrastructureService)
	//ormRepository := infrastructureService.GetModuleOrmRepository(UserKey).(Repository)
	//userService.repository = ormRepository
	return userService, nil
}

func (s *UserService) Foo() string {
	return ""
}
