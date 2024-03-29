package sso

import (
	"context"
	"fmt"
	"time"

	"github.com/gdbu/scribe"
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
	RelationshipUsers               = "users"
	RelationshipLoginCodes          = "loginCodes"
	RelationshipExpiresAtDates      = "expiresAtDates"
	RelationshipExpiresAtHours      = "expiresAtHours"
	RelationshipExpiresAtTimestamps = "expiresAtTimestamps"
)

// relationships is a collection of all the supported relationship keys
var relationships = []string{
	RelationshipUsers,
	RelationshipLoginCodes,
	RelationshipExpiresAtDates,
	RelationshipExpiresAtHours,
	RelationshipExpiresAtTimestamps,
}

const entryTTL = time.Hour

var uuidgen = uuid.NewGenerator()

var (
	sortByExpiresAt = filters.Comparison(RelationshipExpiresAtTimestamps, yesFilter)
)

// New will return a new instance of the Controller
func New(opts mojura.Opts) (cc *Controller, err error) {
	opts.Name = "sso"

	var c Controller
	if c.m, err = mojura.New[*Entry](opts, relationships...); err != nil {
		return
	}

	c.out = scribe.New("SSO")
	c.updateCh = make(chan struct{}, 1)
	c.ctx, c.cancel = context.WithCancel(context.Background())
	go c.expirationScan()
	// Assign pointer reference to our controller
	cc = &c
	return
}

// Controller represents a management layer to facilitate the retrieval and modification of Entries
type Controller struct {
	out *scribe.Scribe

	// Core will manage the data layer and will utilize the underlying back-end
	m *mojura.Mojura[*Entry]

	updateCh chan struct{}

	ctx    context.Context
	cancel func()
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
	if err = c.m.Transaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		// Insert entry into DB
		created, err = c.new(txn, &e)
		return
	}); err != nil {
		return
	}

	notify(c.updateCh)
	return
}

// Get will retrieve an Entry which has the same ID as the provided entryID
func (c *Controller) Get(entryID string) (entry *Entry, err error) {
	return c.m.Get(entryID)
}

// GetByUser will return an entry for a given user (if it exists)
func (c *Controller) GetByUser(ctx context.Context, userID string) (entry *Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		entry, err = c.getByUser(txn, userID)
		return
	})

	return
}

// GetByCode will return an entry for a given login code (if it exists)
func (c *Controller) GetByCode(ctx context.Context, loginCode string) (entry *Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		entry, err = c.getByCode(txn, loginCode)
		return
	})

	return
}

// GetExpiredWithinPreviousHour will return a list of entries which expired in the previous hour
func (c *Controller) GetExpiredWithinPreviousHour(ctx context.Context) (expired []*Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		expired, err = c.getExpiredWithinPreviousHour(txn)
		return
	})

	return
}

// GetExpiredWithinPreviousDay will return a list of entries which expired in the previous day
func (c *Controller) GetExpiredWithinPreviousDay(ctx context.Context) (expired []*Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		expired, err = c.getExpiredWithinPreviousDay(txn)
		return
	})

	return
}

// GetExpiredWithinPreviousDay will return a list of entries which expired in the previous day
func (c *Controller) GetNextToExpire(ctx context.Context) (next *Entry, err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		next, err = c.getNextToExpire(txn)
		return
	})

	return
}

// Login will find a matching entry and return the user ID
func (c *Controller) Login(ctx context.Context, loginCode string) (userID string, err error) {
	if err = c.m.Batch(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		userID, err = c.login(txn, loginCode)
		return
	}); err != nil {
		return
	}

	notify(c.updateCh)
	return
}

// MultiLogin will find a matching entry and return the user ID
// Note: This allows for multiple logins for a single code within a 30 second window
func (c *Controller) MultiLogin(ctx context.Context, loginCode string, ttl time.Duration) (userID string, err error) {
	if err = c.m.Batch(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		userID, err = c.multiLogin(txn, loginCode, ttl)
		return
	}); err != nil {
		return
	}

	notify(c.updateCh)
	return
}

// ForEach will iterate through all Entries
// Note: The error constant mojura.Break can returned by the iterating func to end the iteration early
func (c *Controller) ForEach(fn func(*Entry) error, opts *mojura.FilteringOpts) (err error) {
	// Iterate through all entries
	err = c.m.ForEach(func(_ string, e *Entry) (err error) {
		// Pass iterating Entry to iterating function
		return fn(e)
	}, opts)

	return
}

// Delete will remove an Entry for by entry ID
func (c *Controller) Delete(ctx context.Context, entryID string) (removed *Entry, err error) {
	err = c.m.Transaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		removed, err = c.delete(txn, entryID)
		return
	})

	return
}

// DeleteByUser will remove an Entry for by user ID
func (c *Controller) DeleteByUser(ctx context.Context, userID string) (removed *Entry, err error) {
	err = c.m.Transaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
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
	err = c.m.Transaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		t := newTransaction(txn, c)
		defer t.cleanup()
		return fn(t)
	})

	return
}

// ReadTransaction will initialize a new R-only transaction
func (c *Controller) ReadTransaction(ctx context.Context, fn func(txn *Transaction) (err error)) (err error) {
	err = c.m.ReadTransaction(ctx, func(txn *mojura.Transaction[*Entry]) (err error) {
		t := newTransaction(txn, c)
		defer t.cleanup()
		return fn(t)
	})

	return
}

// Close will close the controller and it's underlying dependencies
func (c *Controller) Close() (err error) {
	c.cancel()
	// Since we only have one dependency, we can just call this func directly
	return c.m.Close()
}

func (c *Controller) new(txn *mojura.Transaction[*Entry], e *Entry) (created *Entry, err error) {
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

func (c *Controller) getByUser(txn *mojura.Transaction[*Entry], userID string) (entry *Entry, err error) {
	userFilter := filters.Match(RelationshipUsers, userID)
	opts := mojura.NewFilteringOpts(userFilter)
	return txn.GetFirst(opts)
}

func (c *Controller) getByCode(txn *mojura.Transaction[*Entry], loginCode string) (entry *Entry, err error) {
	codeFilter := filters.Match(RelationshipLoginCodes, loginCode)
	opts := mojura.NewFilteringOpts(codeFilter)
	return txn.GetFirst(opts)
}

func (c *Controller) getExpiredWithinPreviousHour(txn *mojura.Transaction[*Entry]) (expired []*Entry, err error) {
	filter := newExpiredWithinPreviousHourFilter()
	opts := mojura.NewFilteringOpts(filter)
	expired, _, err = txn.GetFiltered(opts)
	return
}

func (c *Controller) getExpiredWithinPreviousDay(txn *mojura.Transaction[*Entry]) (expired []*Entry, err error) {
	filter := newExpiredWithinPreviousDayFilter()
	opts := mojura.NewFilteringOpts(filter)
	expired, _, err = txn.GetFiltered(opts)
	return
}

func (c *Controller) getNextToExpire(txn *mojura.Transaction[*Entry]) (next *Entry, err error) {
	opts := mojura.NewFilteringOpts(sortByExpiresAt)
	return txn.GetFirst(opts)
}

func (c *Controller) delete(txn *mojura.Transaction[*Entry], userID string) (removed *Entry, err error) {
	var e *Entry
	if e, err = txn.Get(userID); err != nil {
		return
	}

	if _, err = txn.Delete(e.ID); err != nil {
		return
	}

	removed = e
	return
}

func (c *Controller) deleteByUser(txn *mojura.Transaction[*Entry], userID string) (removed *Entry, err error) {
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

	if _, err = txn.Delete(e.ID); err != nil {
		return
	}

	removed = e
	return
}

func (c *Controller) deleteByCode(txn *mojura.Transaction[*Entry], loginCode string) (removed *Entry, err error) {
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

	if _, err = txn.Delete(e.ID); err != nil {
		return
	}

	removed = e
	return
}

func (c *Controller) deleteExpiredInPastHour(txn *mojura.Transaction[*Entry]) (err error) {
	filter := newExpiredWithinPreviousHourFilter()
	opts := mojura.NewFilteringOpts(filter)
	err = txn.ForEachID(func(entryID string) (err error) {
		_, err = c.delete(txn, entryID)
		return
	}, opts)
	return
}

func (c *Controller) deleteExpiredInPastDay(txn *mojura.Transaction[*Entry]) (err error) {
	filter := newExpiredWithinPreviousDayFilter()
	opts := mojura.NewFilteringOpts(filter)
	err = txn.ForEachID(func(entryID string) (err error) {
		_, err = c.delete(txn, entryID)
		return
	}, opts)
	return
}

// Login will find a matching entry and return the user ID
func (c *Controller) login(txn *mojura.Transaction[*Entry], loginCode string) (userID string, err error) {
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

// multiLogin will allow multiple logins through a single code
func (c *Controller) multiLogin(txn *mojura.Transaction[*Entry], loginCode string, ttl time.Duration) (userID string, err error) {
	var entry *Entry
	// Remove entry which matches the login code
	if entry, err = c.getByCode(txn, loginCode); err != nil {
		// No entry was found, return
		return
	} else if entry == nil {
		err = ErrNoCodeMatchFound
		return
	}

	// Get current timestamp
	now := time.Now()

	// Check to see if entry has expired
	if now.After(entry.ExpiresAt) {
		err = fmt.Errorf("cannot login, entry expired at: %v", entry.ExpiresAt)
		return
	}

	newExpiration := now.Add(ttl)
	if entry.ExpiresAt.After(newExpiration) {
		entry.ExpiresAt = newExpiration
		if _, err = txn.Put(entry.ID, entry); err != nil {
			return
		}
	}

	// Set the return user ID value as the user ID of the deleted entry
	userID = entry.UserID
	return
}

func (c *Controller) expirationScan() {
	var (
		next *Entry
		err  error
	)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		next, err = c.GetNextToExpire(c.ctx)
		switch err {
		case nil:
		case mojura.ErrEntryNotFound:
			// Wait for new update to come through update channel
			<-c.updateCh
			continue

		default:
			c.out.Errorf("error getting next to expire: %v", err)
			// Wait for new update to come through update channel
			<-c.updateCh
			continue
		}

		if wait(next.ExpiresAt, c.updateCh) {
			continue
		}

		if _, err = c.Delete(context.Background(), next.ID); err != nil {
			c.out.Errorf("error deleting next to expire: %v", err)
			continue
		}
	}
}
