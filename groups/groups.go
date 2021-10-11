package groups

import (
	"context"

	"github.com/gdbu/stringset"
	"github.com/mojura/kiroku"
	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
)

const (
	relationshipUsers  = "users"
	relationshipGroups = "groups"
)

var relationships = []string{
	relationshipUsers,
	relationshipGroups,
}

// New will return a new instance of users
func New(dir string, exporter kiroku.Exporter) (gp *Groups, err error) {
	var opts mojura.Opts
	opts.Name = "usergroups"
	opts.Dir = dir
	opts.Exporter = exporter

	var g Groups
	if g.c, err = mojura.New(opts, &Entry{}, relationships...); err != nil {
		return
	}

	gp = &g
	return
}

// Groups manages the users
type Groups struct {
	c *mojura.Mojura
}

// Get will get an Entry by user ID
func (g *Groups) Get(userID string) (groups []string, err error) {
	err = g.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		var e *Entry
		e, err = g.get(txn, userID)
		switch err {
		case nil:
			groups = e.Groups.Slice()
		case mojura.ErrEntryNotFound:
			err = nil

		default:
			return
		}

		return
	})

	return
}

// AddGroups will add the provdied groups to a user
func (g *Groups) AddGroups(userID string, groups ...string) (updated *Entry, err error) {
	// Set update func
	updateFn := func(e *Entry) (err error) {
		// Set the provided list of groups
		e.Groups.SetMany(groups)
		return
	}

	err = g.c.Transaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		// Attempt to update the Entry for the given user ID
		if updated, err = g.update(txn, userID, updateFn); err != mojura.ErrEntryNotFound {
			// Error is either nil or an unexpected error. Either way, we want to return
			return
		}

		// Entry was not found for the given user ID, create a new entry and return
		updated, err = g.new(txn, userID, groups)
		return
	})

	return
}

// RemoveGroups will remove the provdied groups from a user
func (g *Groups) RemoveGroups(userID string, groups ...string) (updated *Entry, err error) {
	// Set update func
	updateFn := func(e *Entry) (err error) {
		// Unset the provided list of groups
		e.Groups.UnsetMany(groups)
		return
	}

	err = g.c.Transaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		// Attempt to update the Entry for the given user ID
		if updated, err = g.update(txn, userID, updateFn); err != mojura.ErrEntryNotFound {
			// Error is either nil or an unexpected error. Either way, we want to return
			return
		}

		// Entry was not found for the given user ID, create a new entry and return
		updated, err = g.new(txn, userID, groups)
		return
	})

	return
}

// HasGroup will determine if a user ID has a given group
func (g *Groups) HasGroup(userID string, group string) (hasGroup bool, err error) {
	err = g.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		var e *Entry
		e, err = g.get(txn, userID)
		switch err {
		case nil:
			hasGroup = e.Groups.Has(group)
			return
		case mojura.ErrEntryNotFound:
			return nil

		default:
			return
		}
	})

	return
}

// ForEach will iterate through all users in the database
func (g *Groups) ForEach(seekTo string, fn func(*Entry) error, filters ...mojura.Filter) (err error) {
	opts := mojura.NewIteratingOpts(filters...)
	err = g.c.ForEach(func(userID string, val mojura.Value) (err error) {
		entry := val.(*Entry)
		return fn(entry)
	}, opts)
	return
}

// Close will close the selected instance of users
func (g *Groups) Close() (err error) {
	return g.c.Close()
}

// new will create an Entry for a given user ID
func (g *Groups) new(txn *mojura.Transaction, userID string, groups []string) (created *Entry, err error) {
	var e Entry
	e.UserID = userID
	e.Groups = stringset.MakeMap(groups...)

	if _, err = txn.New(&e); err != nil {
		return
	}

	created = &e
	return
}

// update will edit an Entry
func (g *Groups) update(txn *mojura.Transaction, userID string, fn func(*Entry) error) (updated *Entry, err error) {
	var e *Entry
	if e, err = g.get(txn, userID); err != nil {
		return
	}

	if err = fn(e); err != nil {
		return
	}

	if err = txn.Edit(e.ID, e); err != nil {
		return
	}

	updated = e
	return
}

// get will get an Entry by user ID
func (g *Groups) get(txn *mojura.Transaction, userID string) (entry *Entry, err error) {
	var e Entry
	filter := filters.Match(relationshipUsers, userID)
	opts := mojura.NewIteratingOpts(filter)
	if err = txn.GetFirst(&e, opts); err != nil {
		return
	}

	entry = &e
	return
}
