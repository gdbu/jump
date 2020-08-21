package jump

import (
	"context"
	"net/http"

	"github.com/Hatch1fy/httpserve"
)

// NewSession will apply a session
func (j *Jump) NewSession(ctx *httpserve.Context, userID string) (err error) {
	var key, token string
	// TODO: use httpserve.Context here (needs PR's on httpserver Context type, out of scope currently)
	if key, token, err = j.sess.New(context.Background(), userID); err != nil {
		return
	}

	keyC := setCookie(ctx.Request.Host, CookieKey, key)
	tokenC := setCookie(ctx.Request.Host, CookieToken, token)

	http.SetCookie(ctx.Writer, &keyC)
	http.SetCookie(ctx.Writer, &tokenC)
	return
}
