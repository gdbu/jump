package users

import (
	"github.com/Hatch1fy/errors"
	core "github.com/Hatch1fy/service-core"
)

const (
	// ErrInvalidEmail is returned when an empty email is provided
	ErrInvalidEmail = errors.Error("invalid email, cannot be empty")
	// ErrInvalidPassword is returned when an invalid password is provided
	//	- Password must be at least 6 characters
	//	- Password must not be greater than 24 characters
	ErrInvalidPassword = errors.Error("invalid password, must have a length of at least six and a length less than twenty-four")
	// ErrUserNotFound is returned when a user was not found
	ErrUserNotFound = errors.Error("user not found")
	// ErrInvalidCredentials is returned when a non-matching email/password combo was provided
	ErrInvalidCredentials = errors.Error("invalid credentials")
	// ErrEmailExists is returned when a user is attempting to be created with an email already in use
	ErrEmailExists = errors.Error("email is already associated with a user")

	errBreak = errors.Error("jump break")
)

const (
	lookupEmails = "emails"
)

// New will return a new instance of users
func New(dir string) (up *Users, err error) {
	var u Users
	if u.c, err = core.New("users", dir, &User{}); err != nil {
		return
	}

	up = &u
	return
}

// Users manages the users
type Users struct {
	c *core.Core
}

func (u *Users) getWithFn(id string, fn func(string, core.Value) error) (user *User, err error) {
	var usr User
	if err = fn(id, &usr); err != nil {
		return
	}

	user = &usr
	return
}

// getByEmail will return the matching user for the provided email
func (u *Users) getByEmail(txn *core.Transaction, email string) (user *User, err error) {
	var ids []string
	if ids, err = txn.GetLookup(lookupEmails, email); err != nil {
		return
	}

	if len(ids) == 0 {
		err = ErrUserNotFound
		return
	}

	id := ids[0]

	return u.getWithFn(id, txn.Get)
}

// edit will edit the user which matches the ID
func (u *Users) edit(txn *core.Transaction, id string, fn func(*User) error) (err error) {
	var user *User
	if user, err = u.getWithFn(id, txn.Get); err != nil {
		return
	}

	if err = fn(user); err != nil {
		return
	}

	return txn.Edit(id, user)
}

func (u *Users) updateEmail(txn *core.Transaction, id, email string) (err error) {
	if _, err = u.getByEmail(txn, email); err == nil {
		err = ErrEmailExists
		return
	}

	err = u.edit(txn, id, func(user *User) (err error) {
		user.Email = email
		return
	})

	return
}

func (u *Users) updatePassword(txn *core.Transaction, id, password string) (err error) {
	err = u.edit(txn, id, func(user *User) (err error) {
		user.Password = password
		return user.hashPassword()
	})

	return
}

// New will create a new user
func (u *Users) New(email, password string) (id string, err error) {
	if len(email) == 0 {
		err = ErrInvalidEmail
		return
	}

	user := newUser(email, password)

	if err = user.hashPassword(); err != nil {
		return
	}

	if id, err = u.c.New(&user); err != nil {
		return
	}

	if err = u.c.SetLookup(lookupEmails, email, id); err != nil {
		id = ""
		return
	}

	return
}

// Get will get the user which matches the ID
func (u *Users) Get(id string) (user *User, err error) {
	if user, err = u.getWithFn(id, u.c.Get); err != nil {
		return
	}

	// Clear password
	user.Password = ""
	return
}

// ForEach will iterate through all users in the database
func (u *Users) ForEach(fn func(*User) error) (err error) {
	err = u.c.ForEach(func(userID string, val core.Value) (err error) {
		user := val.(*User)

		// Clear password
		user.Password = ""

		return fn(user)
	})

	return
}

// UpdateEmail will change the user's email
func (u *Users) UpdateEmail(id, email string) (err error) {
	if len(email) == 0 {
		return ErrInvalidEmail
	}

	if err = u.c.Transaction(func(txn *core.Transaction) (err error) {
		return u.updateEmail(txn, id, email)
	}); err != nil {
		return
	}

	return
}

// UpdatePassword will change the user's password
func (u *Users) UpdatePassword(id, password string) (err error) {
	if len(password) == 0 {
		return ErrInvalidPassword
	}

	if err = u.c.Transaction(func(txn *core.Transaction) (err error) {
		return u.updatePassword(txn, id, password)
	}); err != nil {
		return
	}

	return
}

// Match will return the matching email for the provided id and password
func (u *Users) Match(id, password string) (email string, err error) {
	var orig User
	if err = u.c.Get(id, &orig); err != nil {
		return
	}

	if !orig.IsMatch(password) {
		err = ErrInvalidCredentials
		return
	}

	email = orig.Email
	return
}

// MatchEmail will return the matching user id for the provided email and password
func (u *Users) MatchEmail(email, password string) (id string, err error) {
	var orig *User
	if err = u.c.Transaction(func(txn *core.Transaction) (err error) {
		orig, err = u.getByEmail(txn, email)
		return
	}); err != nil {
		return
	}

	if !orig.IsMatch(password) {
		err = ErrInvalidCredentials
		return
	}

	id = orig.ID
	return
}

// Close will close the selected instance of users
func (u *Users) Close() (err error) {
	return u.c.Close()
}
