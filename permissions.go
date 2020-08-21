package jump

import (
	"context"

	"github.com/gdbu/jump/permissions"
)

// SetPermission will give permissions to a provided group for a resourceKey
// Note: See NewResourceKey for more context
func (j *Jump) SetPermission(ctx context.Context, resourceKey, group string, actions, adminActions permissions.Action) (err error) {
	if err = j.perm.SetPermissions(ctx, resourceKey, group, actions); err != nil {
		return
	}

	if err = j.perm.SetPermissions(ctx, resourceKey, "admins", adminActions); err != nil {
		return
	}

	err = nil
	return
}

// UnsetPermission will remove permissions from a provided group for a resourceKey
// Note: See NewResourceKey for more context
func (j *Jump) UnsetPermission(ctx context.Context, resourceKey, group string) (err error) {
	return j.perm.UnsetPermissions(ctx, resourceKey, group)
}

// AddToGroup will add a user to a group
func (j *Jump) AddToGroup(ctx context.Context, userID, group string) (err error) {
	return j.perm.AddGroup(ctx, userID, group)
}

// RemoveFromGroup will remove a user from a group
func (j *Jump) RemoveFromGroup(ctx context.Context, userID, group string) (err error) {
	return j.perm.RemoveGroup(ctx, userID, group)
}
