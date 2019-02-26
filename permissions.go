package jump

import "gitlab.com/itsMontoya/permissions"

// SetPermission will give permissions to a provided group for a resourceKey
// Note: See NewResourceKey for more context
func (j *Jump) SetPermission(resourceKey, group string, actions, adminActions permissions.Action) (err error) {
	if err = j.perm.SetPermissions(resourceKey, group, actions); err != nil && err != permissions.ErrPermissionsUnchanged {
		return
	}

	if err = j.perm.SetPermissions(resourceKey, "admins", adminActions); err != nil && err != permissions.ErrPermissionsUnchanged {
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
