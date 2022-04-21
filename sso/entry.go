package sso

import (
	"strconv"
	"time"

	"github.com/gdbu/uuid"
	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
)

const (
	// ErrEmptyUserID is returned when the user ID for an Entry is empty
	ErrEmptyUserID = errors.Error("invalid user ID, cannot be empty")
	// ErrEmptyLoginCode is returned when the login code is unset
	ErrEmptyLoginCode = errors.Error("invalid login code, cannot be empty")
	// ErrEmptyExpiresAt is returned when the expires at is unset
	ErrEmptyExpiresAt = errors.Error("invalid expires at, cannot be empty")
)

func makeEntry(userID string) (e Entry) {
	e.UserID = userID
	e.LoginCode = uuidgen.Make()
	e.ExpiresAt = time.Now().Add(entryTTL)
	return
}

// Entry represents a stored entry within the Controller
type Entry struct {
	// Include mojura.Entry to auto-populate fields/methods needed to match the
	mojura.Entry

	// UserID is the user which the entry is related to
	UserID string `json:"userID"`
	// LoginCode is the code to be used to log in
	LoginCode uuid.UUID `json:"loginCode"`
	// ExpiresAt will mark when the entry expires
	ExpiresAt time.Time `json:"expiresAt"`
}

// GetRelationships will return the relationship IDs associated with the Entry
func (e *Entry) GetRelationships() (r mojura.Relationships) {
	r.Append(e.UserID)
	r.Append(e.LoginCode.String())
	// Format as datestamp
	r.Append(e.ExpiresAt.UTC().Format("2006-01-02"))
	// Format as hour
	r.Append(e.ExpiresAt.UTC().Format("15"))
	// Format as UNIX timestamp
	r.Append(strconv.FormatInt(e.ExpiresAt.Unix(), 10))
	return
}

// Validate will ensure an Entry is valid
func (e *Entry) Validate() (err error) {
	// An error list allows us to collect all the errors and return them as a group
	var errs errors.ErrorList
	if len(e.UserID) == 0 {
		errs.Push(ErrEmptyUserID)
	}

	if e.LoginCode.IsZero() {
		errs.Push(ErrEmptyLoginCode)
	}

	if e.ExpiresAt.IsZero() {
		errs.Push(ErrEmptyExpiresAt)
	}

	// Note: If error list is empty, a nil value is returned
	return errs.Err()
}
