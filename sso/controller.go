package sso

import (
	"context"
	"time"

	"github.com/gdbu/uuid"
	"github.com/mojura/mojura"
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

const entryTTL = time.Minute * 15

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
func (c *Controller) GetByUser(userID string) (entry *Entry, err error) {
	err = c.m.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		entry, err = c.getByUser(txn, userID)
		return
	})

	return
}

// GetByCode will return an entry for a given login code (if it exists)
func (c *Controller) GetByCode(loginCode string) (entry *Entry, err error) {
	err = c.m.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		entry, err = c.getByCode(txn, loginCode)
		return
	})

	return
}

// GetExpiredWithinPreviousHour will return a list of entries which expired in the previous hour
func (c *Controller) GetExpiredWithinPreviousHour() (expired []*Entry, err error) {
	err = c.m.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		expired, err = c.getExpiredWithinPreviousHour(txn)
		return
	})

	return
}

// GetExpiredWithinPreviousDay will return a list of entries which expired in the previous day
func (c *Controller) GetExpiredWithinPreviousDay() (expired []*Entry, err error) {
	err = c.m.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		expired, err = c.getExpiredWithinPreviousDay(txn)
		return
	})

	return
}

// Login will find a matching entry and return the user ID
func (c *Controller) Login(loginCode string) (userID string, err error) {
	err = c.m.Batch(context.Background(), func(txn *mojura.Transaction) (err error) {
		userID, err = c.login(txn, loginCode)
		return
	})

	return
}

// ForEach will iterate through all Entries
// Note: The error constant mojura.Break can returned by the iterating func to end the iteration early
func (c *Controller) ForEach(seekTo string, fn func(*Entry) error, filters ...mojura.Filter) (err error) {
	// Iterate through all entries
	err = c.m.ForEach(seekTo, func(key string, val mojura.Value) (err error) {
		var e *Entry
		if e, err = asEntry(val); err != nil {
			return
		}

		// Pass iterating Entry to iterating function
		return fn(e)
	}, filters...)

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
	userFilter := mojura.MakeFilter(RelationshipUsers, userID, false)
	// Get list of entries which expired during the last hour
	if err = txn.GetFirst(&e, userFilter); err != nil {
		return
	}

	return
}

func (c *Controller) getByCode(txn *mojura.Transaction, loginCode string) (entry *Entry, err error) {
	var e Entry
	codeFilter := mojura.MakeFilter(RelationshipLoginCodes, loginCode, false)
	// Get list of entries which expired during the last hour
	if err = txn.GetFirst(&e, codeFilter); err != nil {
		return
	}

	return
}

func (c *Controller) getExpiredWithinPreviousHour(txn *mojura.Transaction) (expired []*Entry, err error) {
	previousHour := time.Now().Add(time.Hour * -1).UTC()
	hourFilter := mojura.MakeFilter(RelationshipExpiresAtHours, previousHour.Format("15"), false)

	// Get list of entries which expired during the last hour
	if err = txn.GetFiltered("", &expired, -1, hourFilter); err != nil {
		return
	}

	return
}

func (c *Controller) getExpiredWithinPreviousDay(txn *mojura.Transaction) (expired []*Entry, err error) {
	previousDay := time.Now().Add(time.Hour * 24 * -1).UTC()
	hourFilter := mojura.MakeFilter(RelationshipExpiresAtDates, previousDay.Format("15"), false)

	// Get list of entries which expired during the last hour
	if err = txn.GetFiltered("", &expired, -1, hourFilter); err != nil {
		return
	}

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

// Login will find a matching entry and return the user ID
func (c *Controller) login(txn *mojura.Transaction, loginCode string) (userID string, err error) {
	var removed *Entry
	// Remove entry which matches the login code
	if removed, err = c.deleteByCode(txn, loginCode); removed == nil {
		// No entry was found, return
		return
	}

	// Set the return user ID value as the user ID of the deleted entry
	userID = removed.UserID
	return
}
