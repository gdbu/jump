package apikeys

import (
	"github.com/Hatch1fy/errors"
	"github.com/Hatch1fy/uuid"
	"github.com/boltdb/bolt"

	core "github.com/Hatch1fy/service-core"
)

const (
	// ErrNotInitialized is returned when actions are performed on a non-initialized instance of APIKeys
	ErrNotInitialized = errors.Error("apiKeys library has not been properly initialized")
	// ErrAPIKeyNotFound is returned when a api key is not found
	ErrAPIKeyNotFound = errors.Error("api key not found")
	// ErrInvalidUserID is returned when a user id is empty
	ErrInvalidUserID = errors.Error("invalid user id, cannot be empty")
	// ErrInvalidName is returned when a APIKey's name is empty
	ErrInvalidName = errors.Error("invalid name, cannot be empty")
)

const (
	relationshipUsers = "users"
	relationshipKeys  = "keys"
)

// New will return a new instance of APIKeys
func New(dir string) (ap *APIKeys, err error) {
	var a APIKeys
	if a.c, err = core.New("apikeys", dir, &APIKey{}, relationshipUsers); err != nil {
		return
	}

	// Create UUID generator
	a.gen = uuid.NewGenerator()

	// Assign pointer to created instance of APIKeys
	ap = &a
	return
}

// APIKeys manages the apiKeys service
type APIKeys struct {
	c *core.Core

	db  *bolt.DB
	gen *uuid.Generator
}

func (a *APIKeys) updateName(txn *core.Transaction, apiKey, name string) (err error) {
	var key APIKey
	if err = txn.Get(apiKey, &key); err != nil {
		return
	}

	key.Name = name
	return txn.Edit(apiKey, &key)
}

// New will create a new apiKey and return the associated ID
func (a *APIKeys) New(userID, name string) (key string, err error) {
	uuid := a.gen.New()
	apiKey := newAPIKey(userID, name, uuid.String())
	if err = apiKey.Validate(); err != nil {
		return
	}

	if _, err = a.c.New(&apiKey); err != nil {
		return
	}

	key = apiKey.Key
	return
}

// Get will return the APIKey associated with the provided id
func (a *APIKeys) Get(id string) (apiKey *APIKey, err error) {
	var key APIKey
	if err = a.c.Get(id, &key); err != nil {
		return
	}

	apiKey = &key
	return
}

// GetByKey will return the APIKey associated with the provided apiKey
func (a *APIKeys) GetByKey(key string) (apiKey *APIKey, err error) {
	var as []*APIKey
	if err = a.c.GetByRelationship(relationshipKeys, key, &as); err != nil {
		return
	}

	if len(as) == 0 {
		err = ErrAPIKeyNotFound
		return
	}

	apiKey = as[0]
	return
}

// GetByUser will return the APIKeys associated with the provided user id
func (a *APIKeys) GetByUser(userID string) (apiKeys []*APIKey, err error) {
	err = a.c.GetByRelationship(relationshipUsers, userID, &apiKeys)
	return
}

// UpdateName will edit an APIKey's name
func (a *APIKeys) UpdateName(apiKey, name string) (err error) {
	err = a.c.Transaction(func(txn *core.Transaction) (err error) {
		return a.updateName(txn, apiKey, name)
	})

	return
}

// Remove will delete an apiKey
func (a *APIKeys) Remove(apiKey string) (err error) {
	return a.c.Remove(apiKey)
}

// Close will close the apiKeys service
func (a *APIKeys) Close() (err error) {
	return a.c.Close()
}
