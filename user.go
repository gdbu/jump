package jump

import (
	"context"

	"github.com/gdbu/jump/users"
)

func (j *Jump) postUserCreateActions(ctx context.Context, userID string, groups []string) (apiKey string, err error) {
	// Ensure first group is the user group
	groups = append([]string{userID}, groups...)

	// Add groups to user
	if err = j.perm.AddGroup(ctx, userID, groups...); err != nil {
		return
	}

	if apiKey, err = j.api.New(userID, "primary"); err != nil {
		return
	}

	// Create a new resource key for the generated user ID
	resourceKey := NewResourceKey("user", userID)

	if err = j.SetPermission(ctx, resourceKey, userID, permRWD, permRWD); err != nil {
		return
	}

	return
}

// setLastLoggedInAt sets user last logged in at on the user struct
func (j *Jump) setLastLoggedInAt(userID string, timestamp int64) (err error) {
	err = j.usrs.UpdateLastLoggedInAt(userID, timestamp)
	return
}

// CreateUser will create a user and assign it's basic groups
// Note: It is advised that this function is used when creating users rather than directly calling j.Users().New()
func (j *Jump) CreateUser(ctx context.Context, email, password string, groups ...string) (userID, apiKey string, err error) {
	if userID, err = j.usrs.New(ctx, email, password); err != nil {
		return
	}

	apiKey, err = j.postUserCreateActions(ctx, userID, groups)
	return
}

// InsertUser will insert an existing user (no password hashing)
func (j *Jump) InsertUser(ctx context.Context, email, password string, groups ...string) (userID, apiKey string, err error) {
	if userID, err = j.usrs.Insert(ctx, email, password); err != nil {
		return
	}

	apiKey, err = j.postUserCreateActions(ctx, userID, groups)
	return
}

// GetUser will get a user by ID
func (j *Jump) GetUser(userID string) (user *users.User, err error) {
	return j.usrs.Get(userID)
}

// UpdateEmail will update a user's email address
func (j *Jump) UpdateEmail(ctx context.Context, userID, newEmail string) (err error) {
	return j.usrs.UpdateEmail(ctx, userID, newEmail)
}

// UpdatePassword is the update password handler
func (j *Jump) UpdatePassword(ctx context.Context, userID, newPassword string) (err error) {
	return j.usrs.UpdatePassword(ctx, userID, newPassword)
}

// EnableUser will enable a user
func (j *Jump) EnableUser(ctx context.Context, userID string) (err error) {
	if err = j.usrs.UpdateDisabled(ctx, userID, false); err != nil {
		return
	}

	return
}

// DisableUser will disable a user
func (j *Jump) DisableUser(ctx context.Context, userID string) (err error) {
	if err = j.usrs.UpdateDisabled(ctx, userID, true); err != nil {
		return
	}

	return j.sess.InvalidateUser(ctx, userID)
}
