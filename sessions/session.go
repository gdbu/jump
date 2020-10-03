package sessions

import (
	"time"

	"github.com/gdbu/dbl"
)

func newSession(key, userID string) (s Session) {
	s.Key = key
	s.UserID = userID
	s.setAction()
	return
}

// Session represents a user session
type Session struct {
	dbl.Entry

	// Session key
	Key string `json:"key"`
	// UserID of the user who owns this Session
	UserID string `json:"userID"`

	LastUsedAt int64 `json:"lastUsedAt"`
}

func (s *Session) setAction() {
	s.LastUsedAt = time.Now().Unix()
}

// core.Value interface methods below

// GetID will get the message ID
func (s *Session) GetID() (id string) { return s.ID }

// GetCreatedAt will get the created at timestamp
func (s *Session) GetCreatedAt() (createdAt int64) { return s.CreatedAt }

// GetUpdatedAt will get the updated at timestamp
func (s *Session) GetUpdatedAt() (updatedAt int64) { return s.UpdatedAt }

// GetRelationships will get the associated relationship IDs
func (s *Session) GetRelationships() (r dbl.Relationships) {
	r.Append(s.Key)
	r.Append(s.UserID)
	return
}

// SetID will get the message ID
func (s *Session) SetID(id string) { s.ID = id }

// SetCreatedAt will get the created at timestamp
func (s *Session) SetCreatedAt(createdAt int64) { s.CreatedAt = createdAt }

// SetUpdatedAt will get the updated at timestamp
func (s *Session) SetUpdatedAt(updatedAt int64) { s.UpdatedAt = updatedAt }
