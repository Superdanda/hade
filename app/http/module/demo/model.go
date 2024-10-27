package demo

import (
	"database/sql"
	"github.com/Superdanda/hade/app/provider/demo"
	"github.com/Superdanda/hade/framework/contract"
	"time"
)

func init() {
}

func TypeRegister(typeRegister contract.TypeRegisterService) {
	typeRegister.Register("UserThree", demo.UserThree{})
	typeRegister.Register("UserTwo", demo.UserTwo{})
}

type UserModel struct {
	UserId int
	Name   string
	Age    int
}

// User is gorm model
type User struct {
	ID           uint
	Name         string
	Email        *string
	Age          uint8
	Birthday     *time.Time
	MemberNumber sql.NullString
	ActivatedAt  sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
