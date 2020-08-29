package jump

import (
	"github.com/gdbu/jump/permissions"
)

func (j *Jump) newPermissionHook(userID, resourceName string, actions, adminActions permissions.Action) (hook func(int, map[string]string)) {
	return func(statusCode int, storage map[string]string) {
		if statusCode >= 400 {
			return
		}

		var resourceID string
		if resourceID = storage["resourceID"]; len(resourceID) == 0 {
			j.out.Errorf("Error setting permissions: %v", ErrResourceIDIsEmpty)
			return
		}

		// Create resource key from resource name and resource id
		resourceKey := NewResourceKey(resourceName, resourceID)

		var err error
		if err = j.SetPermission(resourceKey, userID, actions, adminActions); err != nil {
			j.out.Errorf("Error setting permissons for %s / %s: %v", userID, resourceName, err)
		}

		return
	}
}
