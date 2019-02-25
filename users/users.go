package users

import (
	"encoding/json"
	"path/filepath"

	"github.com/Hatch1fy/errors"
	"github.com/boltdb/bolt"
	"gitlab.com/itsMontoya/dbutils"
)

const (
	// ErrNotInitialized is returned when a service has not been properly initialized
	ErrNotInitialized = errors.Error("service not initialized")
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

var (
	bktKey = []byte("users")
)

// New will return a new instance of users
func New(dir string) (up *Users, err error) {
	var u Users
	if err = u.init(dir); err != nil {
		return
	}

	up = &u
	return
}

// Users manages the users
type Users struct {
	db  *bolt.DB
	dbu *dbutils.DBUtils
}

func (u *Users) init(dir string) (err error) {
	filename := filepath.Join(dir, "users.bdb")
	if u.db, err = bolt.Open(filename, 0744, nil); err != nil {
		return
	}

	u.dbu = dbutils.New(8)

	if err = u.db.Update(func(txn *bolt.Tx) (err error) {
		if err = u.dbu.Init(txn); err != nil {
			return
		}

		if _, err = txn.CreateBucketIfNotExists(bktKey); err != nil {
			return
		}

		return
	}); err != nil {
		return
	}

	return
}

// get will return the matching user for the provided id
func (u *Users) get(txn *bolt.Tx, id []byte) (user *User, err error) {
	var bkt *bolt.Bucket
	if bkt = txn.Bucket(bktKey); bkt == nil {
		err = ErrNotInitialized
		return
	}

	var bs []byte
	if bs = bkt.Get(id); len(bs) == 0 {
		err = ErrUserNotFound
		return
	}

	var uu User
	if err = json.Unmarshal(bs, &uu); err != nil {
		return
	}

	user = &uu
	return
}

// getByEmail will return the matching user for the provided email
func (u *Users) getByEmail(txn *bolt.Tx, email []byte) (user *User, err error) {
	var bkt *bolt.Bucket
	if bkt = txn.Bucket(bktKey); bkt == nil {
		err = ErrNotInitialized
		return
	}

	emailStr := string(email)

	err = bkt.ForEach(func(email, bs []byte) (err error) {
		var uu User
		if err = json.Unmarshal(bs, &uu); err != nil {
			return
		}

		if uu.Email != emailStr {
			return
		}

		user = &uu
		return errBreak
	})

	switch err {
	case errBreak:
		err = nil
		return
	case nil:
		err = ErrUserNotFound

	default:
		return
	}

	return
}

func (u *Users) getNextID(txn *bolt.Tx) (id string, err error) {
	var idBytes []byte
	if idBytes, err = u.dbu.Next(txn, bktKey); err != nil {
		return
	}

	id = string(idBytes)
	return
}

func (u *Users) put(txn *bolt.Tx, id []byte, user *User) (err error) {
	var bkt *bolt.Bucket
	if bkt = txn.Bucket(bktKey); bkt == nil {
		err = ErrNotInitialized
		return
	}

	user.ID = string(id)

	var bs []byte
	if bs, err = json.Marshal(user); err != nil {
		return
	}

	return bkt.Put(id, bs)
}

func (u *Users) create(txn *bolt.Tx, user *User) (id string, err error) {
	if _, err = u.getByEmail(txn, []byte(user.Email)); err == nil {
		err = ErrEmailExists
		return
	}

	if user.ID, err = u.getNextID(txn); err != nil {
		return
	}

	if err = user.hashPassword(); err != nil {
		return
	}

	if err = u.put(txn, []byte(user.ID), user); err != nil {
		return
	}

	id = user.ID
	return
}

// New will create a new user
func (u *Users) New(email, password string) (id string, err error) {
	if len(email) == 0 {
		err = ErrInvalidEmail
		return
	}

	user := newUser(email, password)

	if err = u.db.Update(func(txn *bolt.Tx) (err error) {
		id, err = u.create(txn, &user)
		return
	}); err != nil {
		return
	}

	return
}

// Get will get the user which matches the ID
func (u *Users) Get(id string) (user *User, err error) {
	if err = u.db.View(func(txn *bolt.Tx) (err error) {
		user, err = u.get(txn, []byte(id))
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
	if err = u.db.View(func(txn *bolt.Tx) (err error) {
		var bkt *bolt.Bucket
		if bkt = txn.Bucket(bktKey); bkt == nil {
			err = ErrNotInitialized
			return
		}

		return bkt.ForEach(func(userID, userBytes []byte) (err error) {
			var user User
			if err = json.Unmarshal(userBytes, &user); err != nil {
				return
			}

			// Clear password
			user.Password = ""

			return fn(&user)
		})
	}); err != nil {
		return
	}

	return
}

// edit will edit the user which matches the ID
func (u *Users) edit(txn *bolt.Tx, id []byte, fn func(*User) error) (err error) {
	var user *User
	if user, err = u.get(txn, id); err != nil {
		return
	}

	if err = fn(user); err != nil {
		return
	}

	return u.put(txn, id, user)
}

// ChangeEmail will change the user's email
func (u *Users) ChangeEmail(id, email string) (err error) {
	if len(email) == 0 {
		return ErrInvalidEmail
	}

	if err = u.db.Update(func(txn *bolt.Tx) (err error) {
		if _, err = u.getByEmail(txn, []byte(email)); err == nil {
			err = ErrEmailExists
			return
		}

		return u.edit(txn, []byte(id), func(user *User) (err error) {
			user.Email = email
			return
		})
	}); err != nil {
		return
	}

	return
}

// ChangePassword will change the user's password
func (u *Users) ChangePassword(id, password string) (err error) {
	if len(password) == 0 {
		return ErrInvalidPassword
	}

	if err = u.db.Update(func(txn *bolt.Tx) (err error) {
		return u.edit(txn, []byte(id), func(user *User) (err error) {
			user.Password = password
			return user.hashPassword()
		})
	}); err != nil {
		return
	}

	return
}

// Match will return the matching user ID for the provided email and password
func (u *Users) Match(id, password string) (email string, err error) {
	if err = u.db.View(func(txn *bolt.Tx) (err error) {
		var orig *User
		if orig, err = u.get(txn, []byte(id)); err != nil {
			return
		}

		if !orig.IsMatch(password) {
			err = ErrInvalidCredentials
			return
		}

		id = orig.ID
		return
	}); err != nil {
		return
	}

	return
}

// MatchEmail will return the matching user id for the provided email and password
func (u *Users) MatchEmail(email, password string) (id string, err error) {
	if err = u.db.View(func(txn *bolt.Tx) (err error) {
		var orig *User
		if orig, err = u.getByEmail(txn, []byte(email)); err != nil {
			return
		}

		if !orig.IsMatch(password) {
			err = ErrInvalidCredentials
			return
		}

		id = orig.ID
		return
	}); err != nil {
		return
	}

	return
}
