package jump

import (
	"fmt"
	"net/http"

	"github.com/gdbu/scribe"
	"github.com/mojura/mojura"

	"github.com/hatchify/errors"

	"github.com/gdbu/jump/apikeys"
	"github.com/gdbu/jump/events"
	"github.com/gdbu/jump/groups"
	"github.com/gdbu/jump/permissions"
	"github.com/gdbu/jump/sessions"
	"github.com/gdbu/jump/sso"
	"github.com/gdbu/jump/users"
)

const (
	// ErrUserIDIsEmpty is returned when user id is expected but not found within a request context
	ErrUserIDIsEmpty = errors.Error("userID is empty")
	// ErrResourceIDIsEmpty is returned when resource id is expected but not found withing a permissions hook
	ErrResourceIDIsEmpty = errors.Error("resourceID is empty")
)

const (
	permRWD = permissions.ActionRead | permissions.ActionWrite | permissions.ActionDelete
)

const (
	// CookieKey is the jump HTTP key
	CookieKey = "jump_key"
	// CookieToken is the jump HTTP token
	CookieToken = "jump_token"
)

// New will return a new instance of Jump
func New(opts mojura.Opts) (jp *Jump, err error) {
	var j Jump
	j.out = scribe.New("Jump")
	j.evts = events.New()
	if j.perm, err = permissions.New(opts); err != nil {
		return
	}

	if j.sess, err = sessions.New(opts); err != nil {
		return
	}

	if j.api, err = apikeys.New(opts); err != nil {
		return
	}

	if j.usrs, err = users.New(opts, j.evts); err != nil {
		return
	}

	if j.grps, err = groups.New(opts); err != nil {
		return
	}

	if j.sso, err = sso.New(opts); err != nil {
		return
	}

	j.perm.SetGroups(j.grps)
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
	grps *groups.Groups
	sso  *sso.Controller
	evts *events.Controller
}

func (j *Jump) getUserIDFromAPIKey(apiKey string) (userID string, err error) {
	var a *apikeys.APIKey
	if a, err = j.api.Get(apiKey); err != nil {
		err = fmt.Errorf("error getting api key information: %v", err)
		return
	}

	var u *users.User
	if u, err = j.usrs.Get(a.UserID); err != nil {
		err = fmt.Errorf("error getting user \"%s\": %v", a.UserID, err)
		return
	}

	if u.Disabled {
		err = users.ErrUserIsDisabled
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

	var sess *sessions.Session
	if sess, err = j.sess.Get(key.Value, token.Value); err != nil {
		return
	}

	userID = sess.UserID
	return
}

// Events will return the underlying Events controller
func (j *Jump) Events() *events.Controller {
	return j.evts
}

// Permissions will return the underlying permissions
func (j *Jump) Permissions() *permissions.Permissions {
	return j.perm
}

// Users will return the underlying users
func (j *Jump) Users() *users.Users {
	return j.usrs
}

// Groups will return the underlying groups
func (j *Jump) Groups() *groups.Groups {
	return j.grps
}

// Sessions will return the underlying sessions
func (j *Jump) Sessions() *sessions.Sessions {
	return j.sess
}

// APIKeys will return the underlying apikeys
func (j *Jump) APIKeys() *apikeys.APIKeys {
	return j.api
}

// SSO will return the underlying sso
func (j *Jump) SSO() *sso.Controller {
	return j.sso
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
