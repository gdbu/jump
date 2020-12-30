package apikeys

import (
	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
)

func newAPIKey(userID, name, key string) (a APIKey) {
	a.UserID = userID
	a.Name = name
	a.Key = key
	return
}

// APIKey represents an api key reference
type APIKey struct {
	mojura.Entry

	UserID string `json:"userID"`

	Name string `json:"name"`
	Key  string `json:"key"`

	UpdatedAt int64 `json:"updatedAt"`
}

// Validate will validate an API key
func (a *APIKey) Validate() (err error) {
	var errs errors.ErrorList
	if len(a.UserID) == 0 {
		errs.Push(ErrInvalidUserID)
	}

	if len(a.Name) == 0 {
		errs.Push(ErrInvalidName)
	}

	return errs.Err()
}

// GetRelationships will get the associated relationship IDs
func (a *APIKey) GetRelationships() (r mojura.Relationships) {
	r.Append(a.Key)
	r.Append(a.UserID)
	return
}
