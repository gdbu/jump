package main

import (
	"github.com/Hatch1fy/httpserve"
	"github.com/Hatch1fy/jump/users"
)

// SetUserMW is will check user permissions. Expects the following arguments:
//	- redirectOnFail (e.g. false)
func SetUserMW(args ...string) (h httpserve.Handler, err error) {
	var redirectOnFail bool
	switch len(args) {
	case 1:
		redirectOnFail = args[0] == "true"
	case 0:

	default:
		err = ErrInvalidSetUserArguments
		return
	}

	h = p.jump.NewSetUserIDMW(redirectOnFail)
	return
}

// CheckPermissionsMW is will check user permissions. Expects the following arguments:
//	- groupName (e.g. users)
//	- paramKey (e.g. userID)
func CheckPermissionsMW(args ...string) (h httpserve.Handler, err error) {
	if len(args) != 2 {
		err = ErrInvalidCheckPermissionsArguments
		return
	}

	resourceName := args[0]
	paramKey := args[1]
	h = p.jump.NewCheckPermissionsMW(resourceName, paramKey)
	return
}

// GrantPermissionsMW is will grant user permissions. Expects the following arguments:
//	- groupName (e.g. users)
//	- paramKey (e.g. userID)
func GrantPermissionsMW(args ...string) (h httpserve.Handler, err error) {
	if len(args) != 3 {
		err = ErrInvalidGrantPermissionsArguments
		return
	}

	resourceName := args[0]
	actions := getPermissions(args[1])
	adminActions := getPermissions(args[2])
	h = p.jump.NewGrantPermissionsMW(resourceName, actions, adminActions)
	return
}

// CreateUser is a handler for creating a new user
func CreateUser(ctx *httpserve.Context) (res httpserve.Response) {
	var (
		user users.User
		err  error
	)

	if err = ctx.BindJSON(&user); err != nil {
		httpserve.NewJSONResponse(400, err)
	}

	var resp CreateUserResponse
	if resp.UserID, resp.APIKey, err = p.jump.CreateUser(user.Email, user.Password, "users"); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	return httpserve.NewJSONResponse(200, resp)
}

// CreateUserResponse is returned after a user is created
type CreateUserResponse struct {
	UserID string `json:"userID"`
	APIKey string `json:"apiKey"`
}
