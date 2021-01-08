package users

import (
	"strings"

	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
	"golang.org/x/crypto/bcrypt"
)

func newUser(email, password string) (u User) {
	u.Email = strings.ToLower(email)
	u.Password = password
	return
}

// User represents a user
type User struct {
	mojura.Entry

	Email    string `json:"email"`
	Password string `json:"password,omitempty"`

	Verified bool `json:"verified"`
	Disabled bool `json:"disabled"`

	LastLoggedInAt int64 `json:"lastLoggedInAt,omitempty"`
}

func (u *User) sanitize() {
	u.Email = strings.ToLower(u.Email)
}

// IsMatch returns if a provided password is a match for a user
func (u *User) IsMatch(password string) (match bool) {
	if len(u.Password) == 0 {
		return
	}

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
	if len(u.Password) == 0 {
		return
	}

	var hashed []byte
	password := []byte(u.Password)
	if hashed, err = bcrypt.GenerateFromPassword(password, 0); err != nil {
		return
	}

	u.Password = string(hashed)
	return
}

// Everything below are methods required for the mojura.Value interface

// GetRelationships will get the associated relationship IDs
func (u *User) GetRelationships() (r mojura.Relationships) {
	r.Append(u.Email)
	return
}
