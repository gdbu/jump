package jump

import (
	"net/http"

	"github.com/vroomy/httpserve"
)

// NewSession will apply a session
func (j *Jump) NewSession(ctx *httpserve.Context, userID string) (err error) {
	var key, token string
	if key, token, err = j.sess.New(userID); err != nil {
		return
	}

	keyC := setCookie(ctx.GetRequest().Host, CookieKey, key)
	tokenC := setCookie(ctx.GetRequest().Host, CookieToken, token)

	http.SetCookie(ctx.GetWriter(), &keyC)
	http.SetCookie(ctx.GetWriter(), &tokenC)
	return
}
