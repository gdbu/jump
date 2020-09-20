package jump

import (
	"context"
	"net/http"

	vroomy "github.com/vroomy/common"
)

// NewSession will apply a session
func (j *Jump) NewSession(ctx vroomy.Context, userID string) (err error) {
	var key, token string
	// TODO: use httpserve.Context here (needs PR's on httpserver Context type, out of scope currently)
	if key, token, err = j.sess.New(context.Background(), userID); err != nil {
		return
	}

	keyC := setCookie(ctx.GetRequest().Host, CookieKey, key)
	tokenC := setCookie(ctx.GetRequest().Host, CookieToken, token)

	http.SetCookie(ctx.GetWriter(), &keyC)
	http.SetCookie(ctx.GetWriter(), &tokenC)
	return
}
