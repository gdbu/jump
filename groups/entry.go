package groups

import (
	"github.com/gdbu/stringset"
	"github.com/mojura/mojura"
)

// Entry represents a user
type Entry struct {
	mojura.Entry

	UserID string        `json:"userID"`
	Groups stringset.Map `json:"groups"`
}

// Validate will validate a user
func (e *Entry) Validate() (err error) {
	return
}

// GetRelationships will get the associated relationship IDs
func (e *Entry) GetRelationships() (r mojura.Relationships) {
	r.Append(e.UserID)
	r.Append(e.Groups.Slice()...)
	return
}
