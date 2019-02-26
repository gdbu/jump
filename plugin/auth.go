package main

import (
	"net/http"

	"github.com/Hatch1fy/httpserve"
	"github.com/Hatch1fy/jump"
	"github.com/Hatch1fy/jump/users"
)

// Login is the login handler
func Login(ctx *httpserve.Context) (res httpserve.Response) {
	var (
		user users.User
		err  error
	)

	if err = ctx.BindJSON(&user); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	var key, token string
	if key, token, err = p.jump.Login(user.Email, user.Password); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	keyC := setCookie(ctx.Request.URL.Host, jump.CookieKey, key)
	tokenC := setCookie(ctx.Request.URL.Host, jump.CookieToken, token)

	http.SetCookie(ctx.Writer, &keyC)
	http.SetCookie(ctx.Writer, &tokenC)
	return httpserve.NewNoContentResponse()
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
