package jump

import (
	"github.com/gdbu/jump/users"
)

func (j *Jump) postUserCreateActions(userID string, groups []string) (apiKey string, err error) {
	// Ensure first group is the user group
	groups = append([]string{userID}, groups...)

	// Add groups to user
	if _, err = j.grps.AddGroups(userID, groups...); err != nil {
		return
	}

	if apiKey, err = j.api.New(userID, "primary"); err != nil {
		return
	}

	// Create a new resource key for the generated user ID
	resourceKey := NewResourceKey("user", userID)

	if err = j.SetPermission(resourceKey, userID, permRWD, permRWD); err != nil {
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
func (j *Jump) CreateUser(email, password string, groups ...string) (userID, apiKey string, err error) {
	if userID, err = j.usrs.New(email, password); err != nil {
		return
	}

	apiKey, err = j.postUserCreateActions(userID, groups)
	return
}

// InsertUser will insert an existing user (no password hashing)
func (j *Jump) InsertUser(email, password string, groups ...string) (userID, apiKey string, err error) {
	if userID, err = j.usrs.Insert(email, password); err != nil {
		return
	}

	apiKey, err = j.postUserCreateActions(userID, groups)
	return
}

// GetUser will get a user by ID
func (j *Jump) GetUser(userID string) (user *users.User, err error) {
	return j.usrs.Get(userID)
}

// UpdateEmail will update a user's email address
func (j *Jump) UpdateEmail(userID, newEmail string) (err error) {
	return j.usrs.UpdateEmail(userID, newEmail)
}

// UpdatePassword is the update password handler
func (j *Jump) UpdatePassword(userID, newPassword string) (err error) {
	return j.usrs.UpdatePassword(userID, newPassword)
}

// EnableUser will enable a user
func (j *Jump) EnableUser(userID string) (err error) {
	if err = j.usrs.UpdateDisabled(userID, false); err != nil {
		return
	}

	return
}

// DisableUser will disable a user
func (j *Jump) DisableUser(userID string) (err error) {
	if err = j.usrs.UpdateDisabled(userID, true); err != nil {
		return
	}

	return j.sess.InvalidateUser(userID)
}
