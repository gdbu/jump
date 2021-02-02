package jump

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gdbu/jump/permissions"
	"github.com/hatchify/errors"
	"github.com/vroomy/common"
)

const (
	// PermR is a read-only alias
	PermR = permissions.ActionRead
	// PermRW is a read/write alias
	PermRW = permissions.ActionRead | permissions.ActionWrite
	// PermRWD is a read/write/delete alias
	PermRWD = permissions.ActionRead | permissions.ActionWrite | permissions.ActionDelete
)

// NewResourceKey will return a new resource key from a given resource name and resource ID
// Note: Providing an empty resourceID will treat the resource as a grouping resource (No ID association)
func NewResourceKey(resourceName, resourceID string) (resourceKey string) {
	if len(resourceID) == 0 {
		return resourceName
	}

	return fmt.Sprintf("%s::%s", resourceName, resourceID)
}

func getUserID(ctx common.Context) (userID string, err error) {
	if userID = ctx.Get("userID"); len(userID) == 0 {
		err = errors.Error("cannot assert permissions, userID is empty")
		return
	}

	return
}

func getAPIKey(ctx common.Context) (apiKey string) {
	q := ctx.Request().URL.Query()

	if apiKey = q.Get("apiKey"); len(apiKey) > 0 {
		return
	}

	var (
		vals []string
		ok   bool
	)

	if vals, ok = ctx.Request().Header["X-Api-Key"]; !ok {
		return
	}

	if len(vals) == 0 {
		return
	}

	apiKey = vals[0]
	return
}

func newCookie(host, name, value string, expires time.Time) (c http.Cookie) {
	c.Domain = host
	c.Name = name
	c.Value = value
	c.Expires = expires
	c.Path = "/"
	return
}

func setCookie(host, name, value string) (c http.Cookie) {
	return newCookie(host, name, value, time.Now().AddDate(0, 0, 7))
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

func getRedirectURL(ctx common.Context) (redirectURL string) {
	u := loginURL
	q := u.Query()
	q.Set("redirect", ctx.Request().URL.String())
	u.RawQuery = q.Encode()
	return u.String()
}
