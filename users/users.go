package users

import (
	"context"
	"strings"

	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
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
)

const (
	relationshipEmails = "emails"
)

var relationships = []string{relationshipEmails}

// New will return a new instance of users
func New(opts mojura.Opts) (up *Users, err error) {
	opts.Name = "users"

	var u Users
	if u.m, err = mojura.New[User](opts, relationships...); err != nil {
		return
	}

	up = &u
	return
}

// Users manages the users
type Users struct {
	m *mojura.Mojura[User, *User]
}

func (u *Users) new(txn *mojura.Transaction[User, *User], user User) (created *User, err error) {
	if _, err = u.getByEmail(txn, user.Email); err == nil {
		err = ErrEmailExists
		return
	}

	return txn.New(user)
}

// getByEmail will return the matching user for the provided email
func (u *Users) getByEmail(txn *mojura.Transaction[User, *User], email string) (up *User, err error) {
	filter := filters.Match(relationshipEmails, email)
	opts := mojura.NewFilteringOpts(filter)
	return txn.GetFirst(opts)
}

func (u *Users) updateEmail(txn *mojura.Transaction[User, *User], id, email string) (err error) {
	if _, err = u.getByEmail(txn, email); err == nil {
		err = ErrEmailExists
		return
	}

	_, err = txn.Update(id, func(user *User) (err error) {
		user.Email = email
		return
	})

	return
}

func (u *Users) updatePassword(txn *mojura.Transaction[User, *User], id, password string) (err error) {
	_, err = txn.Update(id, func(user *User) (err error) {
		user.Password = password
		return user.hashPassword()
	})

	return
}

func (u *Users) updateVerified(txn *mojura.Transaction[User, *User], id string, verified bool) (err error) {
	_, err = txn.Update(id, func(user *User) (err error) {
		user.Verified = verified
		return
	})

	return
}

func (u *Users) updateDisabled(txn *mojura.Transaction[User, *User], id string, disabled bool) (err error) {
	_, err = txn.Update(id, func(user *User) (err error) {
		user.Disabled = disabled
		return
	})

	return
}

func (u *Users) updateLastLoggedInAt(txn *mojura.Transaction[User, *User], id string, lastLoggedInAt int64) (err error) {
	_, err = txn.Update(id, func(user *User) (err error) {
		user.LastLoggedInAt = lastLoggedInAt
		return
	})

	return
}

// Match will return the matching email for the provided id and password
func (u *Users) match(txn *mojura.Transaction[User, *User], id, password string) (email string, err error) {
	var match *User
	if match, err = txn.Get(id); err != nil {
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
func (u *Users) matchEmail(txn *mojura.Transaction[User, *User], email, password string) (id string, err error) {
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
func (u *Users) New(email, password string) (created *User, err error) {
	if len(email) == 0 {
		err = ErrInvalidEmail
		return
	}

	user := makeUser(email, password)
	user.sanitize()

	if err = user.hashPassword(); err != nil {
		return
	}

	err = u.m.Transaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
		created, err = u.new(txn, user)
		return
	})

	return
}

// Insert will insert an existing user
// Note: No password hashing will occur
func (u *Users) Insert(email, password string) (created *User, err error) {
	if len(email) == 0 {
		err = ErrInvalidEmail
		return
	}

	user := makeUser(email, password)
	user.sanitize()

	err = u.m.Transaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
		created, err = u.new(txn, user)
		return
	})

	return
}

// Get will get the user which matches the ID
func (u *Users) Get(id string) (user *User, err error) {
	if user, err = u.m.Get(id); err != nil {
		return
	}

	// Clear password
	user.Password = ""
	return
}

// GetByEmail will get the user which matches the e,ail
func (u *Users) GetByEmail(email string) (user *User, err error) {
	if err = u.m.ReadTransaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
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
	err = u.m.ForEach(func(_ string, user *User) (err error) {
		// Clear password
		user.Password = ""

		return fn(user)
	}, nil)

	return
}

// UpdateEmail will change the user's email
func (u *Users) UpdateEmail(id, email string) (err error) {
	if len(email) == 0 {
		return ErrInvalidEmail
	}

	// Convert to lowercase
	email = strings.ToLower(email)

	if err = u.m.Transaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
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

	if err = u.m.Transaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
		return u.updatePassword(txn, id, password)
	}); err != nil {
		return
	}

	return
}

// UpdateVerified will change the user's verified state
func (u *Users) UpdateVerified(id string, verified bool) (err error) {
	if err = u.m.Transaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
		return u.updateVerified(txn, id, verified)
	}); err != nil {
		return
	}

	return
}

// UpdateDisabled will change the user's disabled state
func (u *Users) UpdateDisabled(id string, disabled bool) (err error) {
	if err = u.m.Transaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
		return u.updateDisabled(txn, id, disabled)
	}); err != nil {
		return
	}

	return
}

// UpdateLastLoggedInAt will change the user's last logged in at timestamp
func (u *Users) UpdateLastLoggedInAt(id string, lastLoggedInAt int64) (err error) {
	if err = u.m.Batch(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
		return u.updateLastLoggedInAt(txn, id, lastLoggedInAt)
	}); err != nil {
		return
	}

	return
}

// Match will return the matching email for the provided id and password
func (u *Users) Match(id, password string) (email string, err error) {
	err = u.m.ReadTransaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
		email, err = u.match(txn, id, password)
		return
	})

	return
}

// MatchEmail will return the matching user id for the provided email and password
func (u *Users) MatchEmail(email, password string) (id string, err error) {
	err = u.m.ReadTransaction(context.Background(), func(txn *mojura.Transaction[User, *User]) (err error) {
		id, err = u.matchEmail(txn, email, password)
		return
	})

	return
}

// Close will close the selected instance of users
func (u *Users) Close() (err error) {
	return u.m.Close()
}
