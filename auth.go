package jump

import (
	"context"
	"net/http"
	"time"

	"github.com/gdbu/jump/sso"
	"github.com/gdbu/jump/users"
	"github.com/hatchify/errors"
	"github.com/vroomy/common"
)

const (
	// ErrAlreadyLoggedOut is returned when a logout is attempted for a user whom has already logged out of the system.
	ErrAlreadyLoggedOut = errors.Error("already logged out")
)

// Login will attempt to login with a provided email and password combo
// If successful, a key/token pair will be returned to represent the session pair
func (j *Jump) Login(ctx common.Context, email, password string) (userID string, err error) {
	if userID, err = j.usrs.MatchEmail(email, password); err != nil {
		return
	}

	err = j.NewSession(ctx, userID)
	err = j.setLastLoggedInAt(userID, time.Now().Unix())
	return
}

// NewSSO will create a new SSO session
func (j *Jump) NewSSO(ctx context.Context, email string) (loginCode string, err error) {
	var u *users.User
	if u, err = j.usrs.GetByEmail(email); err != nil {
		return
	}

	var e *sso.Entry
	if e, err = j.sso.New(ctx, u.ID); err != nil {
		return
	}

	loginCode = e.LoginCode.String()
	return
}

// SSOLogin will attempt to login with a provided login code
// If successful, a key/token pair will be returned to represent the session pair
func (j *Jump) SSOLogin(ctx common.Context, loginCode string) (err error) {
	var userID string
	if userID, err = j.sso.Login(ctx.Request().Context(), loginCode); err != nil {
		return
	}

	err = j.NewSession(ctx, userID)
	err = j.setLastLoggedInAt(userID, time.Now().Unix())
	return
}

// SSOMultiLogin will attempt to login with a provided login code
// If successful, a key/token pair will be returned to represent the session pair
// Note: Instead of the login code being instantly destroyed, it now has a 30 second TTL
// after usage.
func (j *Jump) SSOMultiLogin(ctx common.Context, loginCode string, ttl time.Duration) (err error) {
	var userID string
	if userID, err = j.sso.MultiLogin(ctx.Request().Context(), loginCode, ttl); err != nil {
		return
	}

	err = j.NewSession(ctx, userID)
	err = j.setLastLoggedInAt(userID, time.Now().Unix())
	return
}

// Logout is the logout handler
func (j *Jump) Logout(ctx common.Context) (err error) {
	userID := ctx.Get("userID")
	if len(userID) == 0 {
		return ErrAlreadyLoggedOut
	}

	var key, token string
	if key, err = getCookieValue(ctx.Request(), CookieKey); err != nil {
		return
	}

	if token, err = getCookieValue(ctx.Request(), CookieToken); err != nil {
		return
	}

	if err = j.sess.Remove(key, token); err != nil {
		return
	}

	keyC := unsetCookie(ctx.Request().URL.Host, CookieKey, key)
	tokenC := unsetCookie(ctx.Request().URL.Host, CookieToken, token)

	http.SetCookie(ctx.Writer(), &keyC)
	http.SetCookie(ctx.Writer(), &tokenC)
	return
}
