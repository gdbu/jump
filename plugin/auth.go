package main

import (
	"fmt"
	"net/http"

	"github.com/Hatch1fy/errors"
	"github.com/Hatch1fy/httpserve"
	"github.com/Hatch1fy/jump"
	"github.com/Hatch1fy/jump/users"
	core "github.com/Hatch1fy/service-core"
)

const (
	// ErrNoLoginFound is returned when an email address is provided that is not found within the system
	ErrNoLoginFound = errors.Error("no login was found for the provided email address")
)

// Login is the login handler
func Login(ctx *httpserve.Context) (res httpserve.Response) {
	var (
		login users.User
		err   error
	)

	if err = ctx.BindJSON(&login); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	var key, token string
	if login.ID, key, token, err = p.jump.Login(login.Email, login.Password); err != nil {
		if err == core.ErrEntryNotFound {
			err = ErrNoLoginFound
		}

		return httpserve.NewJSONResponse(400, err)
	}

	keyC := setCookie(ctx.Request.Host, jump.CookieKey, key)
	tokenC := setCookie(ctx.Request.Host, jump.CookieToken, token)

	http.SetCookie(ctx.Writer, &keyC)
	http.SetCookie(ctx.Writer, &tokenC)

	var user *users.User
	if user, err = p.jump.GetUser(login.ID); err != nil {
		err = fmt.Errorf("error getting user %s: %v", login.ID, err)
		return httpserve.NewJSONResponse(400, err)
	}

	return httpserve.NewJSONResponse(200, user)
}

// Logout is the logout handler
func Logout(ctx *httpserve.Context) (res httpserve.Response) {
	var err error
	userID := ctx.Get("userID")
	if len(userID) == 0 {
		return httpserve.NewJSONResponse(400, ErrAlreadyLoggedOut)
	}

	var key, token string
	if key, err = getCookieValue(ctx.Request, jump.CookieKey); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	if token, err = getCookieValue(ctx.Request, jump.CookieToken); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	if err = p.jump.Logout(key, token); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	keyC := unsetCookie(ctx.Request.URL.Host, jump.CookieKey, key)
	tokenC := unsetCookie(ctx.Request.URL.Host, jump.CookieToken, token)

	http.SetCookie(ctx.Writer, &keyC)
	http.SetCookie(ctx.Writer, &tokenC)

	return httpserve.NewNoContentResponse()
}
