package groups

import (
	"context"

	"github.com/gdbu/stringset"
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
func New(opts mojura.Opts) (gp *Groups, err error) {
	opts.Name = "usergroups"

	var g Groups
	if g.c, err = mojura.New[*Entry](opts, relationships...); err != nil {
		return
	}

	gp = &g
	return
}

// Groups manages the users
type Groups struct {
	c *mojura.Mojura[*Entry]
}

// Get will get an Entry by user ID
func (g *Groups) Get(userID string) (groups []string, err error) {
	err = g.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction[*Entry]) (err error) {
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

// GetByGroup will get user IDs associated with a given group
func (g *Groups) GetByGroup(group string) (userIDs []string, err error) {
	var es []*Entry
	err = g.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction[*Entry]) (err error) {
		es, err = g.getByGroup(txn, group)
		return
	})

	userIDs = make([]string, 0, len(es))
	for _, e := range es {
		userIDs = append(userIDs, e.UserID)
	}

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

	err = g.c.Transaction(context.Background(), func(txn *mojura.Transaction[*Entry]) (err error) {
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

	err = g.c.Transaction(context.Background(), func(txn *mojura.Transaction[*Entry]) (err error) {
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
	err = g.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction[*Entry]) (err error) {
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
	opts := mojura.NewFilteringOpts(filters...)
	err = g.c.ForEach(func(_ string, entry *Entry) (err error) {
		return fn(entry)
	}, opts)
	return
}

// Close will close the selected instance of users
func (g *Groups) Close() (err error) {
	return g.c.Close()
}

// new will create an Entry for a given user ID
func (g *Groups) new(txn *mojura.Transaction[*Entry], userID string, groups []string) (created *Entry, err error) {
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
func (g *Groups) update(txn *mojura.Transaction[*Entry], userID string, fn func(*Entry) error) (updated *Entry, err error) {
	var e *Entry
	if e, err = g.get(txn, userID); err != nil {
		return
	}

	return txn.Update(e.ID, fn)
}

// get will get an Entry by user ID
func (g *Groups) get(txn *mojura.Transaction[*Entry], userID string) (entry *Entry, err error) {
	filter := filters.Match(relationshipUsers, userID)
	opts := mojura.NewFilteringOpts(filter)
	return txn.GetFirst(opts)
}

// getByGroup will get entries by group
func (g *Groups) getByGroup(txn *mojura.Transaction[*Entry], group string) (es []*Entry, err error) {
	filter := filters.Match(relationshipGroups, group)
	opts := mojura.NewFilteringOpts(filter)
	es, _, err = txn.GetFiltered(opts)
	return
}
