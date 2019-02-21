package jump

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

// TODO Get user

// TODO Update user email

// TODO Update user password
