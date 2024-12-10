package user

import (
	"context"
	"encoding/json"
	"github.com/Superdanda/hade/framework/contract"
	"time"
)

const UserKey = "user"

const ChangeAmountTopic = "UserAmountChange"

type Service interface {
	// GetUser 获取用户信息
	GetUser(ctx context.Context, userID int64) (*User, error)

	// SaveUser 保存用户信息
	SaveUser(ctx context.Context, user *User) error

	// AddAmount 金额变化
	AddAmount(ctx context.Context, userID int64, amount int64) error

	// ChangeAmount 订阅ChangeAmountTopic事件
	ChangeAmount(ctx context.Context, event contract.Event) error
}

type User struct {
	ID        int64     `gorm:"column:id;primary_key;auto_increment" json:"id"` // 代表用户id, 只有注册成功之后才有这个id，唯一表示一个用户
	UserName  string    `gorm:"column:username;type:varchar(255);comment:用户名;not null" json:"username"`
	NickName  string    `gorm:"column:username;type:varchar(255);comment:昵称;not null" json:"nickname"`
	Avatar    string    `gorm:"column:username;type:varchar(255);comment:头像" json:"avatar"`
	Password  string    `gorm:"column:password;type:varchar(255);comment:密码;not null" json:"password"`
	Email     string    `gorm:"column:email;type:varchar(255);comment:邮箱;not null" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;comment:创建时间;not null;<-:create" json:"createdAt"`
	Account   *Account  `gorm:"foreignkey:UserId;constraint:OnDelete:CASCADE" json:"account"`
}

type Account struct {
	ID     int64 `gorm:"column:id;primary_key;auto_increment" json:"id"`
	UserId int64 `gorm:"column:user_id;primary_key;auto_increment" json:"userId"`
	Amount int64 `gorm:"column:amount;type:bigint" json:"amount"`
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}
