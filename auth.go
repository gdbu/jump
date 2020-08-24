package jump

import (
	"net/http"
	"time"

	"github.com/Hatch1fy/httpserve"
	"github.com/hatchify/errors"
)

const (
	// ErrAlreadyLoggedOut is returned when a logout is attempted for a user whom has already logged out of the system.
	ErrAlreadyLoggedOut = errors.Error("already logged out")
)

// Login will attempt to login with a provided email and password combo
// If successful, a key/token pair will be returned to represent the session pair
func (j *Jump) Login(ctx *httpserve.Context, email, password string) (userID string, err error) {
	if userID, err = j.usrs.MatchEmail(email, password); err != nil {
		return
	}

	err = j.NewSession(ctx, userID)
	err = j.setLastLoggedInAt(userID, time.Now().Unix())
	return
}

// Logout is the logout handler
func (j *Jump) Logout(ctx *httpserve.Context) (err error) {
	userID := ctx.Get("userID")
	if len(userID) == 0 {
		return ErrAlreadyLoggedOut
	}

	var key, token string
	if key, err = getCookieValue(ctx.Request, CookieKey); err != nil {
		return
	}

	if token, err = getCookieValue(ctx.Request, CookieToken); err != nil {
		return
	}

	if err = j.sess.Remove(key, token); err != nil {
		return
	}

	keyC := unsetCookie(ctx.Request.URL.Host, CookieKey, key)
	tokenC := unsetCookie(ctx.Request.URL.Host, CookieToken, token)

	http.SetCookie(ctx.Writer, &keyC)
	http.SetCookie(ctx.Writer, &tokenC)
	return
}
