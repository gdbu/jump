package permissions

import "github.com/mojura/mojura"

func newTransaction(txn *mojura.Transaction[Resource, *Resource], p *Permissions) (t Transaction) {
	t.txn = txn
	t.p = p
	return

}

// Transaction is the reminders manager
type Transaction struct {
	txn *mojura.Transaction[Resource, *Resource]
	p   *Permissions
}

// Get will get the resource entry for a given resource ID
func (t *Transaction) Get(resourceID string) (rp *Resource, err error) {
	return t.txn.Get(resourceID)
}

// GetByKey will get the resource entry for a given resource key
func (t *Transaction) GetByKey(resourceKey string) (ep *Resource, err error) {
	return t.p.getByKey(t.txn, resourceKey)
}

// SetPermissions will set the permissions for a resource key being accessed by given group
func (t *Transaction) SetPermissions(resourceKey, group string, actions Action) (err error) {
	return t.p.setPermissions(t.txn, resourceKey, group, actions)
}

// SetMultiPermissions will set the permissions for a resource key being accessed by given group
func (t *Transaction) SetMultiPermissions(resourceKey string, pairs ...Pair) (err error) {
	for _, pair := range pairs {
		if err = t.p.setPermissions(t.txn, resourceKey, pair.Group, pair.Actions); err != nil {
			return
		}
	}

	return
}

// UnsetPermissions will remove the permissions for a resource key being accessed by given group
func (t *Transaction) UnsetPermissions(resourceKey, group string) (err error) {
	return t.p.unsetPermissions(t.txn, resourceKey, group)
}

// Can will return if a user (userID) can perform a given action on a provided resource id
// Note: This isn't done as a transaction because it's two GET requests which don't need to block
func (t *Transaction) Can(userID, resourceKey string, action Action) (can bool) {
	return t.p.can(t.txn, userID, resourceKey, action)
}

// Has will return whether or not an ID has a particular group associated with it
func (t *Transaction) Has(resourceID, group string) (ok bool) {
	return t.p.has(t.txn, resourceID, group)
}

// Groups will return a slice of the groups a user belongs to
func (t *Transaction) Groups(userID string) (groups []string, err error) {
	return t.p.g.Get(userID)
}

// RemoveResource will remove a resource by key
func (t *Transaction) RemoveResource(resourceKey string) (err error) {
	return t.p.removeResource(t.txn, resourceKey)
}
