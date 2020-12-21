package jump

import "github.com/gdbu/jump/permissions"

// SetPermission will give permissions to a provided group for a resourceKey
// Note: See NewResourceKey for more context
func (j *Jump) SetPermission(resourceKey, group string, actions, adminActions permissions.Action) (err error) {
	if err = j.perm.SetPermissions(resourceKey, group, actions); err != nil {
		return
	}

	if err = j.perm.SetPermissions(resourceKey, "admins", adminActions); err != nil {
		return
	}

	err = nil
	return
}

// UnsetPermission will remove permissions from a provided group for a resourceKey
// Note: See NewResourceKey for more context
func (j *Jump) UnsetPermission(resourceKey, group string) (err error) {
	return j.perm.UnsetPermissions(resourceKey, group)
}

// AddToGroup will add a user to a group
func (j *Jump) AddToGroup(userID, group string) (err error) {
	_, err = j.grps.AddGroups(userID, group)
	return
}

// RemoveFromGroup will remove a user from a group
func (j *Jump) RemoveFromGroup(userID, group string) (err error) {
	_, err = j.grps.RemoveGroups(userID, group)
	return
}
