package sso

import (
	"context"
	"fmt"
	"time"

	"github.com/gdbu/uuid"
	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
)

const (
	// ErrNoCodeMatchFound is returned when a login code match cannot be found
	ErrNoCodeMatchFound = errors.Error("no login code match was found")
)

// Relationship key const block
const (
	RelationshipUsers          = "users"
	RelationshipLoginCodes     = "loginCodes"
	RelationshipExpiresAtDates = "expiresAtDates"
	RelationshipExpiresAtHours = "expiresAtHours"
)

// relationships is a collection of all the supported relationship keys
var relationships = []string{
	RelationshipUsers,
	RelationshipLoginCodes,
	RelationshipExpiresAtDates,
	RelationshipExpiresAtHours,
}

const entryTTL = time.Hour

var uuidgen = uuid.NewGenerator()

// New will return a new instance of the Controller
func New(dir string) (cc *Controller, err error) {
	var c Controller
	if c.m, err = mojura.New("sso", dir, &Entry{}, relationships...); err != nil {
		return
	}

	// Assign pointer reference to our controller
	cc = &c
	return
}

// Controller represents a management layer to facilitate the retrieval and modification of Entries
type Controller struct {
	// Core will manage the data layer and will utilize the underlying back-end
	m *mojura.Mojura
}

// New will insert a new Entry to the back-end
func (c *Controller) New(ctx context.Context, userID string) (created *Entry, err error) {
	// Create new entry
	e := makeEntry(userID)

	// Validate entry
	if err = e.Validate(); err != nil {
		return
	}

	// Initialize new R/W transaction
	err = c.m.Transaction(ctx, func(txn *mojura.Transaction) (err error) {
		// Insert entry into DB
		created, err = c.new(txn, &e)
		return
	})

	return
}

// Get will retrieve an Entry which has the same ID as the provided entryID
func (c *Controller) Get(entryID string) (entry *Entry, err error) {
	var e Entry
	// Attempt to get Entry with the provided ID, pass reference to entry for which values to be applied
	if err = c.m.Get(entryID, &e); err != nil {
		return
	}

	// Assign reference to retrieved Entry
	entry = &e
	return
}

// GetByUser will return an entry for a given user (if it exists)
func (c *Controller) GetByUser(ctx context.Context, userID string) (entry *Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction) (err error) {
		entry, err = c.getByUser(txn, userID)
		return
	})

	return
}

// GetByCode will return an entry for a given login code (if it exists)
func (c *Controller) GetByCode(ctx context.Context, loginCode string) (entry *Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction) (err error) {
		entry, err = c.getByCode(txn, loginCode)
		return
	})

	return
}

// GetExpiredWithinPreviousHour will return a list of entries which expired in the previous hour
func (c *Controller) GetExpiredWithinPreviousHour(ctx context.Context) (expired []*Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction) (err error) {
		expired, err = c.getExpiredWithinPreviousHour(txn)
		return
	})

	return
}

// GetExpiredWithinPreviousDay will return a list of entries which expired in the previous day
func (c *Controller) GetExpiredWithinPreviousDay(ctx context.Context) (expired []*Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction) (err error) {
		expired, err = c.getExpiredWithinPreviousDay(txn)
		return
	})

	return
}

// Login will find a matching entry and return the user ID
func (c *Controller) Login(ctx context.Context, loginCode string) (userID string, err error) {
	err = c.m.Batch(ctx, func(txn *mojura.Transaction) (err error) {
		userID, err = c.login(txn, loginCode)
		return
	})

	return
}

// ForEach will iterate through all Entries
// Note: The error constant mojura.Break can returned by the iterating func to end the iteration early
func (c *Controller) ForEach(fn func(*Entry) error, opts *mojura.IteratingOpts) (err error) {
	// Iterate through all entries
	err = c.m.ForEach(func(key string, val mojura.Value) (err error) {
		var e *Entry
		if e, err = asEntry(val); err != nil {
			return
		}

		// Pass iterating Entry to iterating function
		return fn(e)
	}, opts)

	return
}

// Delete will remove an Entry for by entry ID
func (c *Controller) Delete(ctx context.Context, entryID string) (removed *Entry, err error) {
	err = c.m.Transaction(ctx, func(txn *mojura.Transaction) (err error) {
		removed, err = c.delete(txn, entryID)
		return
	})

	return
}

// DeleteByUser will remove an Entry for by user ID
func (c *Controller) DeleteByUser(ctx context.Context, userID string) (removed *Entry, err error) {
	err = c.m.Transaction(ctx, func(txn *mojura.Transaction) (err error) {
		removed, err = c.deleteByUser(txn, userID)
		return
	})

	return
}

// DeleteExpiredInPastHour will remove all entries in the past hour
func (c *Controller) DeleteExpiredInPastHour(ctx context.Context) (err error) {
	err = c.m.Transaction(ctx, c.deleteExpiredInPastHour)
	return
}

// DeleteExpiredInPastDay will remove all entries in the past day
func (c *Controller) DeleteExpiredInPastDay(ctx context.Context) (err error) {
	err = c.m.Transaction(ctx, c.deleteExpiredInPastDay)
	return
}

// Transaction will initialize a new R/W transaction
func (c *Controller) Transaction(ctx context.Context, fn func(txn *Transaction) (err error)) (err error) {
	err = c.m.Transaction(ctx, func(txn *mojura.Transaction) (err error) {
		t := newTransaction(txn, c)
		defer t.cleanup()
		return fn(t)
	})

	return
}

// ReadTransaction will initialize a new R-only transaction
func (c *Controller) ReadTransaction(ctx context.Context, fn func(txn *Transaction) (err error)) (err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction) (err error) {
		t := newTransaction(txn, c)
		defer t.cleanup()
		return fn(t)
	})

	return
}

// Close will close the controller and it's underlying dependencies
func (c *Controller) Close() (err error) {
	// Since we only have one dependency, we can just call this func directly
	return c.m.Close()
}

func (c *Controller) new(txn *mojura.Transaction, e *Entry) (created *Entry, err error) {
	// Attempt to validate Entry
	if err = e.Validate(); err != nil {
		// Entry is not valid, return validation error
		return
	}

	// Delete any entries associated with a user
	if _, err = c.deleteByUser(txn, e.UserID); err != nil {
		return
	}

	// Insert Entry into mojura.Core and return the results
	if _, err = txn.New(e); err != nil {
		return
	}

	created = e
	return
}

func (c *Controller) getByUser(txn *mojura.Transaction, userID string) (entry *Entry, err error) {
	var e Entry
	userFilter := filters.Match(RelationshipUsers, userID)
	opts := mojura.NewIteratingOpts(userFilter)
	// Get list of entries which expired during the last hour
	if err = txn.GetFirst(&e, opts); err != nil {
		return
	}

	entry = &e
	return
}

func (c *Controller) getByCode(txn *mojura.Transaction, loginCode string) (entry *Entry, err error) {
	var e Entry
	codeFilter := filters.Match(RelationshipLoginCodes, loginCode)
	opts := mojura.NewIteratingOpts(codeFilter)
	// Get list of entries which expired during the last hour
	if err = txn.GetFirst(&e, opts); err != nil {
		return
	}

	entry = &e
	return
}

func (c *Controller) getExpiredWithinPreviousHour(txn *mojura.Transaction) (expired []*Entry, err error) {
	filter := newExpiredWithinPreviousHourFilter()
	opts := mojura.NewFilteringOpts(filter)
	// Get list of entries which expired during the last hour
	_, err = txn.GetFiltered(&expired, opts)
	return
}

func (c *Controller) getExpiredWithinPreviousDay(txn *mojura.Transaction) (expired []*Entry, err error) {
	filter := newExpiredWithinPreviousDayFilter()
	opts := mojura.NewFilteringOpts(filter)

	// Get list of entries which expired during the last day
	_, err = txn.GetFiltered(&expired, opts)
	return
}

func (c *Controller) delete(txn *mojura.Transaction, userID string) (removed *Entry, err error) {
	var e Entry
	if err = txn.Get(userID, &e); err != nil {
		return
	}

	if err = txn.Remove(e.ID); err != nil {
		return
	}

	removed = &e
	return
}

func (c *Controller) deleteByUser(txn *mojura.Transaction, userID string) (removed *Entry, err error) {
	var e *Entry
	e, err = c.getByUser(txn, userID)
	switch err {
	case nil:
	case mojura.ErrEntryNotFound:
		err = nil
		return
	default:
		return
	}

	if err = txn.Remove(e.ID); err != nil {
		return
	}

	removed = e
	return
}

func (c *Controller) deleteByCode(txn *mojura.Transaction, loginCode string) (removed *Entry, err error) {
	var e *Entry
	e, err = c.getByCode(txn, loginCode)
	switch err {
	case nil:
	case mojura.ErrEntryNotFound:
		err = nil
		return
	default:
		return
	}

	if err = txn.Remove(e.ID); err != nil {
		return
	}

	removed = e
	return
}

func (c *Controller) deleteExpiredInPastHour(txn *mojura.Transaction) (err error) {
	filter := newExpiredWithinPreviousHourFilter()
	opts := mojura.NewIteratingOpts(filter)
	err = txn.ForEachID(func(entryID string) (err error) {
		_, err = c.delete(txn, entryID)
		return
	}, opts)
	return
}

func (c *Controller) deleteExpiredInPastDay(txn *mojura.Transaction) (err error) {
	filter := newExpiredWithinPreviousDayFilter()
	opts := mojura.NewIteratingOpts(filter)
	err = txn.ForEachID(func(entryID string) (err error) {
		_, err = c.delete(txn, entryID)
		return
	}, opts)
	return
}

// Login will find a matching entry and return the user ID
func (c *Controller) login(txn *mojura.Transaction, loginCode string) (userID string, err error) {
	var removed *Entry
	// Remove entry which matches the login code
	if removed, err = c.deleteByCode(txn, loginCode); err != nil {
		// No entry was found, return
		return
	} else if removed == nil {
		err = ErrNoCodeMatchFound
		return
	}

	// Get current timestamp
	now := time.Now()

	// Check to see if entry has expired
	if now.After(removed.ExpiresAt) {
		err = fmt.Errorf("cannot login, entry expired at: %v", removed.ExpiresAt)
		return
	}

	// Set the return user ID value as the user ID of the deleted entry
	userID = removed.UserID
	return
}
