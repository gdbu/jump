package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Hatch1fy/jump/permissions"

	"github.com/Hatch1fy/errors"
	"github.com/Hatch1fy/httpserve"
)

func newResourceKey(name, userID string) (resourceKey string) {
	if len(userID) == 0 {
		return name
	}

	return fmt.Sprintf("%s::%s", name, userID)
}

func getUserID(ctx *httpserve.Context) (userID string, err error) {
	if userID = ctx.Get("userID"); len(userID) == 0 {
		err = errors.Error("cannot assert permissions, userID is empty")
		return
	}

	return
}

func getAPIKey(ctx *httpserve.Context) (apiKey string) {
	q := ctx.Request.URL.Query()

	if apiKey = q.Get("apiKey"); len(apiKey) > 0 {
		return
	}

	var (
		vals []string
		ok   bool
	)

	if vals, ok = ctx.Request.Header["X-Api-Key"]; !ok {
		return
	}

	if len(vals) == 0 {
		return
	}

	apiKey = vals[0]
	return
}

func getPermissions(permsStr string) (a permissions.Action) {
	hasRead := strings.Index(permsStr, "r") > -1
	hasWrite := strings.Index(permsStr, "w") > -1
	hasDelete := strings.Index(permsStr, "d") > -1

	if hasRead {
		a |= permissions.ActionRead
	}

	if hasWrite {
		a |= permissions.ActionWrite
	}

	if hasDelete {
		a |= permissions.ActionDelete
	}

	return
}

func newCookie(host, name, value string, expires time.Time) (c http.Cookie) {
	c.Domain = host
	c.Name = name
	c.Value = value
	c.Expires = expires
	return
}

func setCookie(host, name, value string) (c http.Cookie) {
	return newCookie(host, name, value, time.Now().AddDate(1, 0, 0))
}

func unsetCookie(host, name, value string) (c http.Cookie) {
	return newCookie(host, name, value, time.Now().AddDate(-1, 0, 0))
}

func getCookieValue(req *http.Request, name string) (value string, err error) {
	var c *http.Cookie
	if c, err = req.Cookie(name); err != nil {
		return
	}

	value = c.Value
	return
}
