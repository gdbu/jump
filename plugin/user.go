package main

import (
	"github.com/Hatch1fy/httpserve"
	"github.com/Hatch1fy/jump/users"
)

// GetUserID will get the ID of the currently logged in user
func GetUserID(ctx *httpserve.Context) (res httpserve.Response) {
	return httpserve.NewJSONResponse(200, ctx.Get("userID"))
}

// GetUser will get a user by ID
func GetUser(ctx *httpserve.Context) (res httpserve.Response) {
	var (
		user *users.User
		err  error
	)

	if user, err = p.jump.GetUser(ctx.Param("userID")); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	return httpserve.NewJSONResponse(200, user)
}

// UpdateEmail will update a user's email address
func UpdateEmail(ctx *httpserve.Context) (res httpserve.Response) {
	var (
		user users.User
		err  error
	)

	if err = ctx.BindJSON(&user); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	userID := ctx.Param("userID")

	if err = p.jump.UpdateEmail(userID, user.Email); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	return httpserve.NewNoContentResponse()
}

// UpdatePassword is the update password handler
func UpdatePassword(ctx *httpserve.Context) (res httpserve.Response) {
	var (
		user users.User
		err  error
	)

	if err = ctx.BindJSON(&user); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	userID := ctx.Param("userID")

	if err = p.jump.UpdatePassword(userID, user.Password); err != nil {
		return httpserve.NewJSONResponse(400, err)
	}

	return httpserve.NewNoContentResponse()
}
