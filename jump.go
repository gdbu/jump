package jump

import (
	"fmt"
	"net/http"

	"github.com/missionMeteora/journaler"

	"github.com/Hatch1fy/errors"
	"github.com/Hatch1fy/httpserve"
	"github.com/Hatch1fy/jump/users"

	"gitlab.com/itsMontoya/apikeys"
	"gitlab.com/itsMontoya/permissions"
	"gitlab.com/itsMontoya/sessions"
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
	j.out = journaler.New("Jump")
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
	out *journaler.Journaler

	perm *permissions.Permissions
	sess *sessions.Sessions
	api  *apikeys.APIKeys
	usrs *users.Users
}

// setPermission will give permissions to a provided group for a resourceKey
func (j *Jump) setPermission(resourceKey, group string, actions, adminActions permissions.Action) (err error) {
	if err = j.perm.SetPermissions(resourceKey, group, actions); err != nil && err != permissions.ErrPermissionsUnchanged {
		return
	}

	if err = j.perm.SetPermissions(resourceKey, "admins", adminActions); err != nil && err != permissions.ErrPermissionsUnchanged {
		return
	}

	err = nil
	return
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

func (j *Jump) newPermissionHook(userID, resourceName string, actions, adminActions permissions.Action) (hook httpserve.Hook) {
	return func(statusCode int, storage httpserve.Storage) {
		if statusCode >= 400 {
			return
		}

		var resourceID string
		if resourceID = storage["resourceID"]; len(resourceID) == 0 {
			j.out.Error("Error setting permissions: %v", ErrResourceIDIsEmpty)
			return
		}

		// Create resource key from resource name and resource id
		resourceKey := newResourceKey(resourceName, resourceID)

		var err error
		if err = j.setPermission(resourceKey, userID, actions, adminActions); err != nil {
			j.out.Error("Error setting permissons for %s / %s: %v", userID, resourceName, err)
		}

		return
	}
}

// SetPermission will give permissions to a provided group for a resourceName:resourceID
func (j *Jump) SetPermission(resourceName, resourceID, group string, actions, adminActions permissions.Action) (err error) {
	resourceKey := newResourceKey(resourceName, resourceID)
	return j.setPermission(resourceKey, group, actions, adminActions)
}

// CreateUser will create a user and assign it's basic groups
// Note: It is advised that this function is used when creating users rather than directly calling j.Users().New()
func (j *Jump) CreateUser(email, password string, groups ...string) (userID, apiKey string, err error) {
	if userID, err = j.usrs.New(email, password); err != nil {
		return
	}

	// Ensure first group is the user group
	groups = append([]string{userID}, groups...)

	// Add groups to user
	if err = j.perm.AddGroup(userID, groups...); err != nil {
		return
	}

	if apiKey, err = j.api.New(userID, "primary"); err != nil {
		return
	}

	if err = j.SetPermission("user", userID, userID, permRWD, permRWD); err != nil {
		return
	}

	return
}

// Users will return the users controller
func (j *Jump) Users() *users.Users {
	return j.usrs
}

// APIKeys will return the internal api keys controller
func (j *Jump) APIKeys() *apikeys.APIKeys {
	return j.api
}

// Permissions will return the intenral jump permissions
func (j *Jump) Permissions() *permissions.Permissions {
	return j.perm
}

// NewSession will generate a new session for a given user ID
func (j *Jump) NewSession(userID string) (key, token string, err error) {
	return j.sess.New(userID)
}

// Login will attempt to login with a provided email and password combo
// If successful, a key/token pair will be returned to represent the session pair
func (j *Jump) Login(email, password string) (key, token string, err error) {
	var userID string
	if userID, err = j.usrs.MatchEmail(email, password); err != nil {
		return
	}

	return j.NewSession(userID)
}
