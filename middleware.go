package jump

import (
	"fmt"
	"net/url"

	"github.com/gdbu/jump/permissions"
	"github.com/hatchify/errors"

	"github.com/vroomy/common"
)

var loginURL = url.URL{Path: "/login"}

// NewGrantPermissionsMW will create a new permissions middleware which will grant permissions to a new owner
// Note: The user-group for the current user will be assigned the actions for the resourceName + resourceID (set in context storage).
// If the resourceID is not available in the context, and error will be logged
func (j *Jump) NewGrantPermissionsMW(resourceName string, actions, adminActions permissions.Action) common.Handler {
	return func(ctx common.Context) {
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
	}
}

// NewSetUserIDMW will set the user id of the currently logged in user
func (j *Jump) NewSetUserIDMW(redirectOnFail, allowNonLoggedIn bool) (fn common.Handler) {
	return func(ctx common.Context) {
		var (
			userID string
			err    error
		)

		userID, err = j.getUserIDFromRequest(ctx)
		switch {
		case err == nil:
			// No error occurred, set user ID
			ctx.Put("userID", userID)
		case allowNonLoggedIn:
			// Error occurred, but we are allowing non logged in access, do nothing
		case redirectOnFail:
			// Error occurred and redirect on fail flag is set, redirect
			ctx.Redirect(302, getRedirectURL(ctx))
		default:
			// Error occurred, write as JSON error
			ctx.WriteJSON(401, err)
		}
	}
}

// NewCheckPermissionsMW will check the user to ensure they have permissions to view a particular resource
func (j *Jump) NewCheckPermissionsMW(resourceName, paramKey string) common.Handler {
	return func(ctx common.Context) {
		userID := ctx.Get("userID")
		if len(userID) == 0 {
			ctx.WriteJSON(401, errors.Error("cannot assert permissions, user ID is empty"))
			return
		}

		var action permissions.Action
		switch ctx.Request().Method {
		case "GET", "OPTIONS":
			action = permissions.ActionRead
		case "PUT", "POST", "PATCH":
			action = permissions.ActionWrite
		case "DELETE":
			action = permissions.ActionDelete

		default:
			err := fmt.Errorf("cannot assert permissions, unsupported method: %s", ctx.Request().Method)
			ctx.WriteJSON(500, err)
			return
		}

		var resourceID string
		if len(paramKey) > 0 {
			resourceID = NewResourceKey(resourceName, ctx.Param(paramKey))
		} else {
			resourceID = resourceName
		}

		if !j.perm.Can(userID, resourceID, action) {
			ctx.WriteJSON(403, errors.Error("forbidden"))
			return
		}
	}
}

func (j *Jump) getUserIDFromRequest(ctx common.Context) (userID string, err error) {
	if apiKey := getAPIKey(ctx); len(apiKey) > 0 {
		if userID, err = j.getUserIDFromAPIKey(apiKey); err != nil {
			err = fmt.Errorf("error getting user ID from API key: %v", err)
			return
		}
	}

	if userID, err = j.getUserIDFromSession(ctx.Request()); err != nil {
		err = fmt.Errorf("error getting user ID from session key: %v", err)
		return
	}

	return
}
