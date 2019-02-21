package jump

import "gitlab.com/itsMontoya/permissions"

// SetPermission will give permissions to a provided group for a resourceName:resourceID
func (j *Jump) SetPermission(resourceName, resourceID, group string, actions, adminActions permissions.Action) (err error) {
	resourceKey := NewResourceKey(resourceName, resourceID)
	return j.setPermission(resourceKey, group, actions, adminActions)
}
