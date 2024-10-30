package user

import "github.com/Superdanda/hade/app/provider/user"

func ConvertUserToDTO(user *user.User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{}
}
