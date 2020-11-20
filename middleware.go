package jump

import (
	"fmt"
	"net/url"

	"github.com/gdbu/jump/permissions"
	"github.com/hatchify/errors"

	"github.com/vroomy/httpserve"
)

var loginURL = url.URL{Path: "/login"}

// NewGrantPermissionsMW will create a new permissions middleware which will grant permissions to a new owner
// Note: The user-group for the current user will be assigned the actions for the resourceName + resourceID (set in context storage).
// If the resourceID is not available in the context, and error will be logged
func (j *Jump) NewGrantPermissionsMW(resourceName string, actions, adminActions permissions.Action) httpserve.Handler {
	return func(ctx *httpserve.Context) (res httpserve.Response) {
		var (
			userID string
			err    error
		)

		if userID, err = getUserID(ctx); err != nil {
			j.out.Errorf("Error getting user id while setting permissions: %v", err)
			return
		}

		hook := j.newPermissionHook(userID, resourceName, actions, adminActions)
		ctx.AddHook(hook)
		return
	}
}

// NewSetUserIDMW will set the user id of the currently logged in user
func (j *Jump) NewSetUserIDMW(redirectOnFail bool) (fn func(ctx *httpserve.Context) (res httpserve.Response)) {
	return func(ctx *httpserve.Context) (res httpserve.Response) {
		var (
			userID string
			err    error
		)

		if apiKey := getAPIKey(ctx); len(apiKey) > 0 {
			if userID, err = j.getUserIDFromAPIKey(apiKey); err != nil {
				err = fmt.Errorf("error getting user ID from API key: %v", err)
			}
		} else {
			if userID, err = j.getUserIDFromSession(ctx.GetRequest()); err != nil {
				err = fmt.Errorf("error getting user ID from session key: %v", err)
			}
		}

		if err != nil {
			if !redirectOnFail {
				return httpserve.NewJSONResponse(401, err)
			}

			return httpserve.NewRedirectResponse(302, getRedirectURL(ctx))
		}

		ctx.Put("userID", userID)
		return
	}
}

// NewCheckPermissionsMW will check the user to ensure they have permissions to view a particular resource
func (j *Jump) NewCheckPermissionsMW(resourceName, paramKey string) httpserve.Handler {
	return func(ctx *httpserve.Context) (res httpserve.Response) {
		userID := ctx.Get("userID")
		if len(userID) == 0 {
			return httpserve.NewJSONResponse(401, errors.Error("cannot assert permissions, user ID is empty"))
		}

		var action permissions.Action
		switch ctx.GetRequest().Method {
		case "GET", "OPTIONS":
			action = permissions.ActionRead
		case "PUT", "POST", "PATCH":
			action = permissions.ActionWrite
		case "DELETE":
			action = permissions.ActionDelete

		default:
			err := fmt.Errorf("cannot assert permissions, unsupported method: %s", ctx.GetRequest().Method)
			return httpserve.NewJSONResponse(500, err)
		}

		var resourceID string
		if len(paramKey) > 0 {
			resourceID = NewResourceKey(resourceName, ctx.Param(paramKey))
		} else {
			resourceID = resourceName
		}

		if !j.perm.Can(userID, resourceID, action) {
			return ctx.NewJSONResponse(403, errors.Error("forbidden"))
		}

		return
	}
}
