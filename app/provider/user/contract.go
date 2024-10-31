package user

import (
	"context"
	"encoding/json"
	"time"
)

const UserKey = "user"

type Service interface {
	// GetUser 获取用户信息
	GetUser(ctx context.Context, userID int64) (*User, error)

	// SaveUser 保存用户信息
	SaveUser(ctx context.Context, user *User) error
}

type User struct {
	ID        int64     `gorm:"column:id;primary_key;auto_increment" json:"id"` // 代表用户id, 只有注册成功之后才有这个id，唯一表示一个用户
	UserName  string    `gorm:"column:username;type:varchar(255);comment:用户名;not null" json:"username"`
	NickName  string    `gorm:"column:username;type:varchar(255);comment:昵称;not null" json:"nickname"`
	Avatar    string    `gorm:"column:username;type:varchar(255);comment:头像" json:"avatar"`
	Password  string    `gorm:"column:password;type:varchar(255);comment:密码;not null" json:"password"`
	Email     string    `gorm:"column:email;type:varchar(255);comment:邮箱;not null" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;comment:创建时间;not null;<-:create" json:"createdAt"`
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}
