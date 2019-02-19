package main

import (
	"fmt"
	"strings"

	"gitlab.com/itsMontoya/permissions"

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
