package jump

import (
	"github.com/Hatch1fy/jump/users"
)

// GetUsersList will get the current users list
func (j *Jump) GetUsersList() (us []*users.User, err error) {
	if err = j.usrs.ForEach(func(user *users.User) (err error) {
		us = append(us, user)
		return
	}); err != nil {
		us = nil
		return
	}

	return
}
