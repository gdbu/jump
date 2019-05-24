package permissions

import core "github.com/Hatch1fy/service-core"

func newTransaction(txn *core.Transaction, p *Permissions) (t Transaction) {
	t.txn = txn
	t.p = p
	return

}

// Transaction is the reminders manager
type Transaction struct {
	txn *core.Transaction
	p   *Permissions
}

// Get will get the resource entry for a given resource ID
func (t *Transaction) Get(resourceID string) (rp *Resource, err error) {
	var r Resource
	if err = t.txn.Get(resourceID, &r); err != nil {
		return
	}

	rp = &r
	return
}

// GetByKey will get the resource entry for a given resource key
func (t *Transaction) GetByKey(resourceKey string) (ep *Resource, err error) {
	return getByKey(t.txn, resourceKey)
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

// AddGroup will add a group to a userID
func (t *Transaction) AddGroup(userID string, groups ...string) (err error) {
	return t.p.addGroup(t.txn, userID, groups)
}

// RemoveGroup will remove a group to a userID
func (t *Transaction) RemoveGroup(userID string, groups ...string) (err error) {
	return t.p.removeGroup(t.txn, userID, groups)
}

// Can will return if a user (userID) can perform a given action on a provided resource id
// Note: This isn't done as a transaction because it's two GET requests which don't need to block
func (t *Transaction) Can(userID, resourceKey string, action Action) (can bool) {
	var (
		e      *Resource
		groups []string
		err    error
	)

	if e, err = getByKey(t.txn, resourceKey); err != nil {
		return
	}

	if groups, err = t.txn.GetLookup(lookupGroups, userID); err != nil {
		return
	}

	for _, group := range groups {
		if can = e.Can(group, action); can {
			return
		}
	}

	return
}

// Has will return whether or not an ID has a particular group associated with it
func (t *Transaction) Has(resourceID, group string) (ok bool) {
	var e Resource
	if err := t.txn.Get(resourceID, &e); err != nil {
		return
	}

	return e.Has(group)
}

// Groups will return a slice of the groups a user belongs to
func (t *Transaction) Groups(userID string) (groups []string, err error) {
	return t.txn.GetLookup(lookupGroups, userID)
}
