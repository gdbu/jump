package sso

import (
	"context"

	"github.com/mojura/mojura"
)

func newTransaction(txn *mojura.Transaction, c *Controller) *Transaction {
	var t Transaction
	t.txn = txn
	t.c = c
	return &t
}

// Transaction represents a controller transaction
type Transaction struct {
	txn *mojura.Transaction
	c   *Controller
}

// New will insert a new Entry to the back-end
func (t *Transaction) New(ctx context.Context, userID string) (created *Entry, err error) {
	// Create new entry
	e := makeEntry(userID)

	return t.c.new(t.txn, &e)
}

// Get will retrieve an Entry which has the same ID as the provided entryID
func (t *Transaction) Get(entryID string) (entry *Entry, err error) {
	var e Entry
	// Attempt to get Entry with the provided ID, pass reference to entry for which values to be applied
	if err = t.txn.Get(entryID, &e); err != nil {
		return
	}

	// Assign reference to retrieved Entry
	entry = &e
	return
}

// GetByUser will return an entry for a given user (if it exists)
func (t *Transaction) GetByUser(userID string) (entry *Entry, err error) {
	return t.c.getByUser(t.txn, userID)
}

// GetByCode will return an entry for a given login code (if it exists)
func (t *Transaction) GetByCode(loginCode string) (entry *Entry, err error) {
	return t.c.getByCode(t.txn, loginCode)
}

// GetExpiredWithinPreviousHour will return a list of entries which expired in the previous hour
func (t *Transaction) GetExpiredWithinPreviousHour() (expired []*Entry, err error) {
	return t.c.getExpiredWithinPreviousHour(t.txn)
}

// GetExpiredWithinPreviousDay will return a list of entries which expired in the previous day
func (t *Transaction) GetExpiredWithinPreviousDay() (expired []*Entry, err error) {
	return t.c.getExpiredWithinPreviousDay(t.txn)
}

// Login will find a matching entry and return the user ID
func (t *Transaction) Login(loginCode string) (userID string, err error) {
	return t.c.login(t.txn, loginCode)
}

// ForEach will iterate through all Entries
// Note: The error constant mojura.Break can returned by the iterating func to end the iteration early
func (t *Transaction) ForEach(fn func(*Entry) error, opts *mojura.IteratingOpts) (err error) {
	// Iterate through all entries
	err = t.txn.ForEach(func(key string, val mojura.Value) (err error) {
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
func (t *Transaction) Delete(ctx context.Context, entryID string) (removed *Entry, err error) {
	return t.c.delete(t.txn, entryID)
}

// DeleteByUser will remove an Entry for by user ID
func (t *Transaction) DeleteByUser(ctx context.Context, userID string) (removed *Entry, err error) {
	return t.c.deleteByUser(t.txn, userID)
}

// DeleteExpiredInPastHour will remove all entries in the past hour
func (t *Transaction) DeleteExpiredInPastHour(ctx context.Context) (err error) {
	return t.c.deleteExpiredInPastHour(t.txn)
}

// DeleteExpiredInPastDay will remove all entries in the past day
func (t *Transaction) DeleteExpiredInPastDay(ctx context.Context) (err error) {
	return t.c.deleteExpiredInPastDay(t.txn)
}
func (t *Transaction) cleanup() {
	t.txn = nil
	t.c = nil
}
