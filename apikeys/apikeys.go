package apikeys

import (
	"context"

	"github.com/gdbu/uuid"
	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
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
	relationshipKeys  = "keys"
	relationshipUsers = "users"
)

var (
	relationships = []string{relationshipKeys, relationshipUsers}
)

// New will return a new instance of APIKeys
func New(opts mojura.Opts) (ap *APIKeys, err error) {
	opts.Name = "apikeys"

	var a APIKeys
	if a.m, err = mojura.New(opts, &APIKey{}, relationships...); err != nil {
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
	m *mojura.Mojura

	gen *uuid.Generator
}

// New will create a new apiKey and return the associated ID
func (a *APIKeys) New(userID, name string) (key string, err error) {
	uuid := a.gen.New()
	apiKey := newAPIKey(userID, name, uuid.String())
	if err = apiKey.Validate(); err != nil {
		return
	}

	if _, err = a.m.New(&apiKey); err != nil {
		return
	}

	key = apiKey.Key
	return
}

// Get will return the APIKey entry associated with the provided api key value
func (a *APIKeys) Get(key string) (apiKey *APIKey, err error) {
	err = a.m.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		apiKey, err = a.get(txn, key)
		return
	})

	return
}

// GetByUser will return the APIKeys associated with the provided user id
func (a *APIKeys) GetByUser(userID string) (apiKeys []*APIKey, err error) {
	filter := filters.Match(relationshipUsers, userID)
	opts := mojura.NewFilteringOpts(filter)
	_, err = a.m.GetFiltered(&apiKeys, opts)
	return
}

// UpdateName will edit an APIKey's name
func (a *APIKeys) UpdateName(apiKey, name string) (err error) {
	err = a.m.Transaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		return a.updateName(txn, apiKey, name)
	})

	return
}

// Remove will delete an apiKey
func (a *APIKeys) Remove(apiKey string) (removed *APIKey, err error) {
	err = a.m.Transaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		removed, err = a.remove(txn, apiKey)
		return
	})

	return
}

// Close will close the apiKeys service
func (a *APIKeys) Close() (err error) {
	return a.m.Close()
}

func (a *APIKeys) get(txn *mojura.Transaction, key string) (apiKey *APIKey, err error) {
	var entry APIKey
	filter := filters.Match(relationshipKeys, key)
	opts := mojura.NewIteratingOpts(filter)
	if err = txn.GetFirst(&entry, opts); err != nil {
		return
	}

	apiKey = &entry
	return
}

func (a *APIKeys) updateName(txn *mojura.Transaction, apiKey, name string) (err error) {
	var match *APIKey
	if match, err = a.get(txn, apiKey); err != nil {
		return
	}

	match.Name = name
	return txn.Edit(match.ID, match)
}

func (a *APIKeys) remove(txn *mojura.Transaction, apiKey string) (removed *APIKey, err error) {
	var match *APIKey
	if match, err = a.get(txn, apiKey); err != nil {
		return
	}

	if err = txn.Remove(match.ID); err != nil {
		return
	}

	removed = match
	return
}
