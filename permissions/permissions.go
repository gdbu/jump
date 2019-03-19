package permissions

import (
	"github.com/Hatch1fy/errors"
	core "github.com/Hatch1fy/service-core"
)

const (
	// ErrInvalidActions is returned when an invalid permissions value is attempted to be set
	ErrInvalidActions = errors.Error("invalid permissions, please see constant block for reference")
	// ErrPermissionsUnchanged is returned when matching permissions are set for a resource
	ErrPermissionsUnchanged = errors.Error("permissions match, unchanged")
	// ErrResourceNotFound is returned when a requested resource cannot be found
	ErrResourceNotFound = errors.Error("resource not found")
	// ErrGroupNotFound is returned when a requested group cannot be found
	ErrGroupNotFound = errors.Error("group not found")
)

const (
	relationshipResourceKeys = "resourceKeys"
)

const (
	lookupGroups = "groups"
)

// New will return a new instance of Permissions
func New(dir string) (pp *Permissions, err error) {
	var p Permissions
	if p.c, err = core.New("permissions", dir, &Entry{}, relationshipResourceKeys); err != nil {
		return
	}

	pp = &p
	return
}

// Permissions manages permissions
type Permissions struct {
	c *core.Core
}

func (p *Permissions) setPermissions(txn *core.Transaction, resourceKey, group string, actions Action) (err error) {
	var e *Entry
	if e, err = getOrCreateByKey(txn, resourceKey); err != nil {
		return
	}

	if !e.Set(group, actions) {
		return ErrPermissionsUnchanged
	}

	return txn.Edit(e.ID, e)
}

func (p *Permissions) unsetPermissions(txn *core.Transaction, resourceKey, group string) (err error) {
	var e *Entry
	if e, err = getByKey(txn, resourceKey); err != nil {
		return
	}

	if !e.Remove(group) {
		return ErrPermissionsUnchanged
	}

	return txn.Edit(e.ID, e)
}

func (p *Permissions) addGroup(txn *core.Transaction, userID string, groups []string) (err error) {
	updated := false
	for _, group := range groups {
		if err = txn.SetLookup(lookupGroups, userID, group); err != nil {
			return
		}

		updated = true
	}

	if !updated {
		return ErrPermissionsUnchanged
	}

	return
}

func (p *Permissions) removeGroup(txn *core.Transaction, userID string, groups []string) (err error) {
	updated := false
	for _, group := range groups {
		if err = txn.RemoveLookup(lookupGroups, userID, group); err != nil {
			return
		}

		updated = true
	}

	if !updated {
		return ErrPermissionsUnchanged
	}

	return
}

// Get will get the entry for a given resource ID
func (p *Permissions) Get(resourceID string) (ep *Entry, err error) {

	var e Entry
	if err = p.c.Get(resourceID, &e); err != nil {
		return
	}

	ep = &e
	return
}

// GetByKey will get the permissions for a given group for a resource key
func (p *Permissions) GetByKey(resourceKey string) (ep *Entry, err error) {
	return getByKey(p.c, resourceKey)
}

// SetPermissions will set the permissions for a resource key being accessed by given group
func (p *Permissions) SetPermissions(resourceKey, group string, actions Action) (err error) {
	err = p.c.Transaction(func(txn *core.Transaction) (err error) {
		return p.setPermissions(txn, resourceKey, group, actions)
	})

	return
}

// UnsetPermissions will remove the permissions for a resource key being accessed by given group
func (p *Permissions) UnsetPermissions(resourceKey, group string) (err error) {
	err = p.c.Transaction(func(txn *core.Transaction) (err error) {
		return p.unsetPermissions(txn, resourceKey, group)
	})

	return
}

// AddGroup will add a group to a userID
func (p *Permissions) AddGroup(userID string, groups ...string) (err error) {
	err = p.c.Transaction(func(txn *core.Transaction) (err error) {
		return p.addGroup(txn, userID, groups)
	})

	return
}

// RemoveGroup will remove a group to a userID
func (p *Permissions) RemoveGroup(userID string, groups ...string) (err error) {
	err = p.c.Transaction(func(txn *core.Transaction) (err error) {
		return p.removeGroup(txn, userID, groups)
	})

	return
}

// Can will return if a user (userID) can perform a given action on a provided resource id
// Note: This isn't done as a transaction because it's two GET requests which don't need to block
func (p *Permissions) Can(userID, resourceKey string, action Action) (can bool) {
	var (
		e      *Entry
		groups []string
		err    error
	)

	if e, err = getByKey(p.c, resourceKey); err != nil {
		return
	}

	if groups, err = p.c.GetLookup(lookupGroups, userID); err != nil {
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
func (p *Permissions) Has(resourceID, group string) (ok bool) {
	var e Entry
	if err := p.c.Get(resourceID, &e); err != nil {
		return
	}

	return e.Has(group)
}

// Groups will return a slice of the groups a user belongs to
func (p *Permissions) Groups(userID string) (groups []string, err error) {
	return p.c.GetLookup(lookupGroups, userID)
}

// Transaction will initialize a transaction for all methods to be executed under
func (p *Permissions) Transaction(fn func(*Transaction) error) (err error) {
	err = p.c.Transaction(func(txn *core.Transaction) (err error) {
		t := newTransaction(txn, p)
		err = fn(&t)
		t.txn = nil
		return
	})

	return
}

// Close will close permissions
func (p *Permissions) Close() (err error) {
	return p.c.Close()
}
