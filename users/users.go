package users

import (
	"strings"

	core "github.com/gdbu/dbl"
	"github.com/hatchify/errors"
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
	ErrInvalidCredentials = errors.Error("invalid password")
	// ErrEmailExists is returned when a user is attempting to be created with an email already in use
	ErrEmailExists = errors.Error("email is already associated with a user")
	// ErrUserIsDisabled is returned when a user is disabled
	ErrUserIsDisabled = errors.Error("user is disabled")

	errBreak = errors.Error("jump break")
)

const (
	relationshipEmails = "emails"
)

var relationships = []string{relationshipEmails}

// New will return a new instance of users
func New(dir string) (up *Users, err error) {
	var u Users
	if u.c, err = core.New("users", dir, &User{}, relationships...); err != nil {
		return
	}

	up = &u
	return
}

// Users manages the users
type Users struct {
	c *core.Core
}

func (u *Users) new(txn *core.Transaction, user *User) (entryID string, err error) {
	if _, err = u.getByEmail(txn, user.Email); err == nil {
		err = ErrEmailExists
		return
	}

	entryID, err = txn.New(user)
	return
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
func (u *Users) getByEmail(txn *core.Transaction, email string) (up *User, err error) {
	var user User
	if err = txn.GetFirstByRelationship(relationshipEmails, email, &user); err != nil {
		return
	}

	up = &user
	return
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

func (u *Users) updateDisabled(txn *core.Transaction, id string, disabled bool) (err error) {
	err = u.edit(txn, id, func(user *User) (err error) {
		user.Disabled = disabled
		return
	})

	return
}

func (u *Users) updateLastLoggedInAt(txn *core.Transaction, id string, lastLoggedInAt int64) (err error) {
	err = u.edit(txn, id, func(user *User) (err error) {
		user.LastLoggedInAt = lastLoggedInAt
		return
	})

	return
}

// Match will return the matching email for the provided id and password
func (u *Users) match(txn *core.Transaction, id, password string) (email string, err error) {
	var match User
	if err = txn.Get(id, &match); err != nil {
		return
	}

	if !match.IsMatch(password) {
		err = ErrInvalidCredentials
		return
	}

	if match.Disabled {
		err = ErrUserIsDisabled
		return
	}

	email = match.Email
	return
}

// MatchEmail will return the matching user id for the provided email and password
func (u *Users) matchEmail(txn *core.Transaction, email, password string) (id string, err error) {
	// Ensure the comparing email is all lower case
	email = strings.ToLower(email)

	var match *User
	// Attempt to get match
	if match, err = u.getByEmail(txn, email); err != nil {
		return
	}

	if !match.IsMatch(password) {
		err = ErrInvalidCredentials
		return
	}

	if match.Disabled {
		err = ErrUserIsDisabled
		return
	}

	id = match.ID
	return
}

// New will create a new user
func (u *Users) New(email, password string) (entryID string, err error) {
	if len(email) == 0 {
		err = ErrInvalidEmail
		return
	}

	user := newUser(email, password)
	user.sanitize()

	if err = user.hashPassword(); err != nil {
		return
	}

	err = u.c.Transaction(func(txn *core.Transaction) (err error) {
		entryID, err = u.new(txn, &user)
		return
	})

	return
}

// Insert will insert an existing user
// Note: No password hashing will occur
func (u *Users) Insert(email, password string) (entryID string, err error) {
	if len(email) == 0 {
		err = ErrInvalidEmail
		return
	}

	user := newUser(email, password)
	user.sanitize()

	err = u.c.Transaction(func(txn *core.Transaction) (err error) {
		entryID, err = u.new(txn, &user)
		return
	})

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

// GetByEmail will get the user which matches the e,ail
func (u *Users) GetByEmail(email string) (user *User, err error) {
	if err = u.c.ReadTransaction(func(txn *core.Transaction) (err error) {
		if user, err = u.getByEmail(txn, email); err != nil {
			return
		}

		return
	}); err != nil {
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

	// Convert to lowercase
	email = strings.ToLower(email)

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

// UpdateDisabled will change the user's disabled state
func (u *Users) UpdateDisabled(id string, disabled bool) (err error) {
	if err = u.c.Transaction(func(txn *core.Transaction) (err error) {
		return u.updateDisabled(txn, id, disabled)
	}); err != nil {
		return
	}

	return
}

// UpdateLastLoggedInAt will change the user's last logged in at timestamp
func (u *Users) UpdateLastLoggedInAt(id string, lastLoggedInAt int64) (err error) {
	if err = u.c.Batch(func(txn *core.Transaction) (err error) {
		return u.updateLastLoggedInAt(txn, id, lastLoggedInAt)
	}); err != nil {
		return
	}

	return
}

// Match will return the matching email for the provided id and password
func (u *Users) Match(id, password string) (email string, err error) {
	err = u.c.ReadTransaction(func(txn *core.Transaction) (err error) {
		email, err = u.match(txn, id, password)
		return
	})

	return
}

// MatchEmail will return the matching user id for the provided email and password
func (u *Users) MatchEmail(email, password string) (id string, err error) {
	err = u.c.ReadTransaction(func(txn *core.Transaction) (err error) {
		id, err = u.matchEmail(txn, email, password)
		return
	})

	return
}

// Close will close the selected instance of users
func (u *Users) Close() (err error) {
	return u.c.Close()
}
