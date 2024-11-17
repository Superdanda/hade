package http

import (
	"github.com/Superdanda/hade/app/provider/user"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

func SubscribeEvent(container framework.Container) error {
	queueService := container.MustMake(contract.QueueKey).(contract.QueueService)
	userService := container.MustMake(user.UserKey).(user.Service)

	queueService.RegisterSubscribe(user.ChangeAmountTopic, func(event contract.Event) error {
		return userService.ChangeAmount(queueService.GetContext(), event.Payload())
	})

	return nil
}
