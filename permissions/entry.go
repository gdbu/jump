package permissions

func newEntry(key string) (e Entry) {
	e.Key = key
	e.Groups = make(Groups)
	return
}

// Entry represents a permisssions entry
type Entry struct {
	ID string `json:"id"`

	Key    string `json:"key"`
	Groups `json:"groups"`

	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
}

// core.Value interface methods below

// GetID will get the message ID
func (e *Entry) GetID() (id string) { return e.ID }

// GetCreatedAt will get the created at timestamp
func (e *Entry) GetCreatedAt() (createdAt int64) { return e.CreatedAt }

// GetUpdatedAt will get the updated at timestamp
func (e *Entry) GetUpdatedAt() (updatedAt int64) { return e.UpdatedAt }

// GetRelationshipIDs will get the associated relationship IDs
func (e *Entry) GetRelationshipIDs() (ids []string) {
	ids = append(ids, e.Key)
	return
}

// SetID will get the message ID
func (e *Entry) SetID(id string) { e.ID = id }

// SetCreatedAt will get the created at timestamp
func (e *Entry) SetCreatedAt(createdAt int64) { e.CreatedAt = createdAt }

// SetUpdatedAt will get the updated at timestamp
func (e *Entry) SetUpdatedAt(updatedAt int64) { e.UpdatedAt = updatedAt }
