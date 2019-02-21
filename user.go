package jump

import (
	"github.com/Hatch1fy/jump/users"
)

// CreateUser will create a user and assign it's basic groups
// Note: It is advised that this function is used when creating users rather than directly calling j.Users().New()
func (j *Jump) CreateUser(email, password string, groups ...string) (userID, apiKey string, err error) {
	if userID, err = j.usrs.New(email, password); err != nil {
		return
	}

	// Ensure first group is the user group
	groups = append([]string{userID}, groups...)

	// Add groups to user
	if err = j.perm.AddGroup(userID, groups...); err != nil {
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

// GetUser will get a user by ID
func (j *Jump) GetUser(userID string) (user *users.User, err error) {
	return j.usrs.Get(userID)
}

// UpdateEmail will update a user's email address
func (j *Jump) UpdateEmail(userID, newEmail string) (err error) {
	return j.usrs.ChangeEmail(userID, newEmail)
}

// UpdatePassword is the update password handler
func (j *Jump) UpdatePassword(userID, newPassword string) (err error) {
	return j.usrs.ChangePassword(userID, newPassword)
}
