package permissions

import (
	"context"
	"log"

	"github.com/gdbu/jump/groups"
	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
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

// New will return a new instance of Permissions
func New(opts mojura.Opts) (pp *Permissions, err error) {
	opts.Name = "permissions"

	var p Permissions
	if p.c, err = mojura.New[Resource](opts, relationshipResourceKeys); err != nil {
		return
	}

	pp = &p
	return
}

// Permissions manages permissions
type Permissions struct {
	c *mojura.Mojura[Resource, *Resource]
	g *groups.Groups
}

// Get will get the resource entry for a given resource ID
func (p *Permissions) Get(resourceID string) (ep *Resource, err error) {
	return p.c.Get(resourceID)
}

// GetByKey will get the resource entry for a given resource key
func (p *Permissions) GetByKey(resourceKey string) (r *Resource, err error) {
	err = p.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
		r, err = p.getByKey(txn, resourceKey)
		return
	})

	return
}

// SetPermissions will set the permissions for a resource key being accessed by given group
func (p *Permissions) SetPermissions(resourceKey, group string, actions Action) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
		return p.setPermissions(txn, resourceKey, group, actions)
	})

	return
}

// SetMultiPermissions will set the permissions for a resource key being accessed by given group
func (p *Permissions) SetMultiPermissions(resourceKey string, pairs ...Pair) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
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
	err = p.c.Transaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
		return p.unsetPermissions(txn, resourceKey, group)
	})

	return
}

// UnsetMultiPermissions will remove the permissions for a resource key being accessed set of groups
func (p *Permissions) UnsetMultiPermissions(resourceKey string, groups ...string) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
		for _, group := range groups {
			if err = p.unsetPermissions(txn, resourceKey, group); err != nil {
				return
			}
		}

		return
	})

	return
}

// Can will return if a user (userID) can perform a given action on a provided resource id
// Note: This isn't done as a transaction because it's two GET requests which don't need to block
func (p *Permissions) Can(userID, resourceKey string, action Action) (can bool) {
	if err := p.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
		can = p.can(txn, userID, resourceKey, action)
		return
	}); err != nil {
		log.Printf("Permissions.Can(): Error checking can state: %v", err)
		return
	}

	return
}

// Has will return whether or not an ID has a particular group associated with it
func (p *Permissions) Has(resourceID, group string) (has bool) {
	if err := p.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
		has = p.has(txn, resourceID, group)
		return
	}); err != nil {
		log.Printf("Permissions.Can(): Error checking can state: %v", err)
		return
	}

	return
}

// RemoveResource will remove a resource by key
func (p *Permissions) RemoveResource(resourceKey string) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
		return p.removeResource(txn, resourceKey)
	})

	return
}

// Transaction will initialize a transaction for all methods to be executed under
func (p *Permissions) Transaction(fn func(*Transaction) error) (err error) {
	err = p.c.Transaction(context.Background(), func(txn *mojura.Transaction[Resource, *Resource]) (err error) {
		t := newTransaction(txn, p)
		err = fn(&t)
		t.txn = nil
		return
	})

	return
}

// SetGroups will set the groups controller
func (p *Permissions) SetGroups(g *groups.Groups) {
	p.g = g
}

// Close will close permissions
func (p *Permissions) Close() (err error) {
	return p.c.Close()
}

func (p *Permissions) setPermissions(txn *mojura.Transaction[Resource, *Resource], resourceKey, group string, actions Action) (err error) {
	var r *Resource
	if r, err = p.getOrCreateByKey(txn, resourceKey); err != nil {
		return
	}

	if !r.Set(group, actions) {
		return
	}

	_, err = txn.Put(r.ID, *r)
	return
}

func (p *Permissions) unsetPermissions(txn *mojura.Transaction[Resource, *Resource], resourceKey, group string) (err error) {
	var r *Resource
	if r, err = p.getByKey(txn, resourceKey); err != nil {
		return
	}

	if !r.Remove(group) {
		return
	}

	_, err = txn.Put(r.ID, *r)
	return
}

func (p *Permissions) removeResource(txn *mojura.Transaction[Resource, *Resource], resourceKey string) (err error) {
	var r *Resource
	if r, err = p.getByKey(txn, resourceKey); err != nil {
		return
	}

	_, err = txn.Delete(r.ID)
	return
}

func (p *Permissions) getByKey(txn *mojura.Transaction[Resource, *Resource], resourceKey string) (r *Resource, err error) {
	var rs []*Resource
	filter := filters.Match(relationshipResourceKeys, resourceKey)
	opts := mojura.NewFilteringOpts(filter)
	if rs, _, err = txn.GetFiltered(opts); err != nil {
		return
	}

	if len(rs) == 0 {
		err = ErrResourceNotFound
		return
	}

	r = rs[0]
	return
}

func (p *Permissions) getOrCreateByKey(txn *mojura.Transaction[Resource, *Resource], resourceKey string) (rp *Resource, err error) {
	if rp, err = p.getByKey(txn, resourceKey); err != ErrResourceNotFound {
		return
	}

	r := makeResource(resourceKey)
	return txn.New(r)
}

// Can will return if a user (userID) can perform a given action on a provided resource id
// Note: This isn't done as a transaction because it's two GET requests which don't need to block
func (p *Permissions) can(txn *mojura.Transaction[Resource, *Resource], userID, resourceKey string, action Action) (can bool) {
	var (
		e      *Resource
		groups []string
		err    error
	)

	if e, err = p.getByKey(txn, resourceKey); err != nil {
		return
	}

	if groups, err = p.g.Get(userID); err != nil {
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
func (p *Permissions) has(txn *mojura.Transaction[Resource, *Resource], resourceID, group string) (ok bool) {
	var (
		e   *Resource
		err error
	)

	if e, err = txn.Get(resourceID); err != nil {
		return
	}

	return e.Has(group)
}
