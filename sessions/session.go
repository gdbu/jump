package sessions

import (
	"time"
)

func newSession(key, userID string) (s Session) {
	s.Key = key
	s.UserID = userID
	s.setAction()
	return
}

// Session represents a user session
type Session struct {
	// ID of the Session
	ID string `json:"id"`

	// Session key
	Key string `json:"key"`
	// UserID of the user who owns this Session
	UserID string `json:"userID"`

	LastUsedAt int64 `json:"lastUsedAt"`
	CreatedAt  int64 `json:"createdAt"`
	UpdatedAt  int64 `json:"updatedAt"`
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

// GetRelationshipIDs will get the associated relationship IDs
func (s *Session) GetRelationshipIDs() (ids []string) {
	ids = append(ids, s.Key)
	ids = append(ids, s.UserID)
	return
}

// SetID will get the message ID
func (s *Session) SetID(id string) { s.ID = id }

// SetCreatedAt will get the created at timestamp
func (s *Session) SetCreatedAt(createdAt int64) { s.CreatedAt = createdAt }

// SetUpdatedAt will get the updated at timestamp
func (s *Session) SetUpdatedAt(updatedAt int64) { s.UpdatedAt = updatedAt }
