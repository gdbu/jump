package permissions

import (
	"context"

	core "github.com/gdbu/dbl"
	"github.com/hatchify/errors"
)

const (
	// ErrInvalidActions is returned when an invalid permissions value is attempted to be set
	ErrInvalidActions = errors.Error("invalid permissions, please see constant block for reference")
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
	if p.c, err = core.New("permissions", dir, &Resource{}, relationshipResourceKeys); err != nil {
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
	var r *Resource
	if r, err = getOrCreateByKey(txn, resourceKey); err != nil {
		return
	}

	if !r.Set(group, actions) {
		return
	}

	return txn.Edit(r.ID, r)
}

func (p *Permissions) unsetPermissions(txn *core.Transaction, resourceKey, group string) (err error) {
	var r *Resource
	if r, err = getByKey(txn, resourceKey); err != nil {
		return
	}

	if !r.Remove(group) {
		return
	}

	return txn.Edit(r.ID, r)
}

func (p *Permissions) removeResource(txn *core.Transaction, resourceKey string) (err error) {
	var r *Resource
	if r, err = getByKey(txn, resourceKey); err != nil {
		return
	}

	return txn.Remove(r.ID)
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
		return
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
		return
	}

	return
}

// Get will get the resource entry for a given resource ID
func (p *Permissions) Get(resourceID string) (ep *Resource, err error) {

	var e Resource
	if err = p.c.Get(resourceID, &e); err != nil {
		return
	}

	ep = &e
	return
}

// GetByKey will get the resource entry for a given resource key
func (p *Permissions) GetByKey(resourceKey string) (r *Resource, err error) {
	return getByKey(p.c, resourceKey)
}

// SetPermissions will set the permissions for a resource key being accessed by given group
func (p *Permissions) SetPermissions(resourceKey, group string, actions Action) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
		return p.setPermissions(txn, resourceKey, group, actions)
	})

	return
}

// SetMultiPermissions will set the permissions for a resource key being accessed by given group
func (p *Permissions) SetMultiPermissions(resourceKey string, pairs ...Pair) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
		for _, pair := range pairs {
			if err = p.setPermissions(txn, resourceKey, pair.Group, pair.Actions); err != nil {
				return
			}
		}

		return
	})

	return
}

// UnsetPermissions will remove the permissions for a resource key being accessed by given group
func (p *Permissions) UnsetPermissions(resourceKey, group string) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
		return p.unsetPermissions(txn, resourceKey, group)
	})

	return
}

// UnsetMultiPermissions will remove the permissions for a resource key being accessed set of groups
func (p *Permissions) UnsetMultiPermissions(resourceKey string, groups ...string) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
		for _, group := range groups {
			if err = p.unsetPermissions(txn, resourceKey, group); err != nil {
				return
			}
		}

		return
	})

	return
}

// AddGroup will add a group to a userID
func (p *Permissions) AddGroup(userID string, groups ...string) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
		return p.addGroup(txn, userID, groups)
	})

	return
}

// RemoveGroup will remove a group from a userID
func (p *Permissions) RemoveGroup(userID string, groups ...string) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
		return p.removeGroup(txn, userID, groups)
	})

	return
}

// RemoveAllGroups will remove all groups from a userID
func (p *Permissions) RemoveAllGroups(userID string, groups ...string) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
		var groups []string
		if groups, err = p.c.GetLookup(lookupGroups, userID); err != nil {
			return
		}

		return p.removeGroup(txn, userID, groups)
	})

	return
}

// Can will return if a user (userID) can perform a given action on a provided resource id
// Note: This isn't done as a transaction because it's two GET requests which don't need to block
func (p *Permissions) Can(userID, resourceKey string, action Action) (can bool) {
	var (
		e      *Resource
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
	var e Resource
	if err := p.c.Get(resourceID, &e); err != nil {
		return
	}

	return e.Has(group)
}

// Groups will return a slice of the groups a user belongs to
func (p *Permissions) Groups(userID string) (groups []string, err error) {
	return p.c.GetLookup(lookupGroups, userID)
}

// HasGroup will return if a user belongs to a given group
func (p *Permissions) HasGroup(userID, group string) (has bool) {
	var (
		groups []string
		err    error
	)

	if groups, err = p.c.GetLookup(lookupGroups, userID); err != nil {
		return
	}

	for _, g := range groups {
		if g == group {
			return true
		}
	}

	return false
}

// RemoveResource will remove a resource by key
func (p *Permissions) RemoveResource(resourceKey string) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
		return p.removeResource(txn, resourceKey)
	})

	return
}

// Transaction will initialize a transaction for all methods to be executed under
func (p *Permissions) Transaction(fn func(*Transaction) error) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *core.Transaction) (err error) {
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

// NewPair will return a new permissions pair
func NewPair(group string, actions Action) (p Pair) {
	p.Group = group
	p.Actions = actions
	return
}

// Pair represents a permissions pair
type Pair struct {
	Group   string
	Actions Action
}
