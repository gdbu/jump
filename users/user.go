package users

import (
	"github.com/Hatch1fy/errors"
	"golang.org/x/crypto/bcrypt"
)

func newUser(email, password string) (u User) {
	u.Email = email
	u.Password = password
	return
}

// User represents a user
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// IsMatch returns if a provided password is a match for a user
func (u *User) IsMatch(password string) (match bool) {
	hashed := []byte(u.Password)
	pw := []byte(password)
	return bcrypt.CompareHashAndPassword(hashed, pw) == nil
}

// Validate will validate a user
func (u *User) Validate() (err error) {
	var errs errors.ErrorList
	if len(u.Email) == 0 {
		errs.Push(ErrInvalidEmail)
	}

	if len(u.Password) < 6 || len(u.Password) > 24 {
		errs.Push(ErrInvalidPassword)
	}

	return errs.Err()
}

func (u *User) hashPassword() (err error) {
	var hashed []byte
	password := []byte(u.Password)
	if hashed, err = bcrypt.GenerateFromPassword(password, 0); err != nil {
		return
	}

	u.Password = string(hashed)
	return
}
