package jump

import (
	"fmt"
	"net/http"

	"github.com/hatchify/scribe"

	"github.com/Hatch1fy/errors"

	"github.com/Hatch1fy/jump/apikeys"
	"github.com/Hatch1fy/jump/permissions"
	"github.com/Hatch1fy/jump/sessions"
	"github.com/Hatch1fy/jump/users"
)

const (
	// ErrUserIDIsEmpty is returned when user id is expected but not found within a request context
	ErrUserIDIsEmpty = errors.Error("userID is empty")
	// ErrResourceIDIsEmpty is returned when resource id is expected but not found withing a permissions hook
	ErrResourceIDIsEmpty = errors.Error("resourceID is empty")
)

const (
	permR   = permissions.ActionRead
	permRW  = permissions.ActionRead | permissions.ActionWrite
	permRWD = permissions.ActionRead | permissions.ActionWrite | permissions.ActionDelete
	permWD  = permissions.ActionWrite | permissions.ActionDelete
)

const (
	// CookieKey is the jump HTTP key
	CookieKey = "jump_key"
	// CookieToken is the jump HTTP token
	CookieToken = "jump_token"
)

// New will return a new instance of Jump
func New(dir string) (jp *Jump, err error) {
	var j Jump
	j.out = scribe.New("Jump")
	if j.perm, err = permissions.New(dir); err != nil {
		return
	}

	if j.sess, err = sessions.New(dir); err != nil {
		return
	}

	if j.api, err = apikeys.New(dir); err != nil {
		return
	}

	if j.usrs, err = users.New(dir); err != nil {
		return
	}

	jp = &j
	return
}

// Jump manages the basic ancillary components of a web service
type Jump struct {
	out *scribe.Scribe

	perm *permissions.Permissions
	sess *sessions.Sessions
	api  *apikeys.APIKeys
	usrs *users.Users
}

func (j *Jump) getUserIDFromAPIKey(apiKey string) (userID string, err error) {
	var a *apikeys.APIKey
	if a, err = j.api.Get(apiKey); err != nil {
		err = fmt.Errorf("error getting api key information: %v", err)
		return
	}

	userID = a.UserID
	return
}

func (j *Jump) getUserIDFromSession(req *http.Request) (userID string, err error) {
	var key *http.Cookie
	if key, err = req.Cookie(CookieKey); err != nil {
		return
	}

	var token *http.Cookie
	if token, err = req.Cookie(CookieToken); err != nil {
		return
	}

	return j.sess.Get(key.Value, token.Value)
}

// Permissions will return the underlying permissions
func (j *Jump) Permissions() *permissions.Permissions {
	return j.perm
}

// Users will return the underlying users
func (j *Jump) Users() *users.Users {
	return j.usrs
}

// Close will close jump
func (j *Jump) Close() (err error) {
	var errs errors.ErrorList
	errs.Push(j.usrs.Close())
	errs.Push(j.sess.Close())
	errs.Push(j.api.Close())
	errs.Push(j.perm.Close())
	return errs.Err()
}
