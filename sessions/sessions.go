package sessions

import (
	"time"

	"github.com/Hatch1fy/errors"
	core "github.com/Hatch1fy/service-core"
	"github.com/boltdb/bolt"
	"github.com/missionMeteora/journaler"
	"github.com/missionMeteora/uuid"
)

const (
	// ErrSessionDoesNotExist is returned when an invalid token/key pair is presented
	ErrSessionDoesNotExist = errors.Error("session with that token/key pair does not exist")
)

const (
	// SessionTimeout (in seconds) is the ttl for sessions, an action will refresh the duration
	SessionTimeout = 60 * 60 * 12 // 12 hours
)

var (
	sessionsBktKey = []byte("sessions")
)

const (
	relationshipKeys  = "keys"
	relationshipUsers = "users"
)

var (
	relationships = []string{relationshipKeys, relationshipUsers}
)

// New will return a new instance of sessions
func New(dir string) (sp *Sessions, err error) {
	var s Sessions
	if s.c, err = core.New("sessions", dir, &Session{}, relationships...); err != nil {
		return
	}

	s.g = uuid.NewGen()
	s.out = journaler.New("Sessions")

	// Start purge loop
	go s.loop()
	sp = &s
	return
}

// Sessions manages sessions
type Sessions struct {
	out *journaler.Journaler
	c   *core.Core
	g   *uuid.Gen
}

func (s *Sessions) newKeyToken() (key, token string) {
	// Set key
	key = s.g.New().String()
	// Set token
	token = s.g.New().String()
	return
}

func (s *Sessions) newSession(key, token, userID string) Session {
	// Set session key
	sessionKey := newSessionKey(key, token)
	// Create new session
	return newSession(sessionKey, userID)
}

func (s *Sessions) getByKey(txn *core.Transaction, key string) (sp *Session, err error) {
	var ss []*Session
	if err = txn.GetByRelationship(relationshipKeys, key, &ss); err != nil {
		return
	}

	if len(ss) == 0 {
		err = core.ErrEntryNotFound
		return
	}

	sp = ss[0]
	return
}

func (s *Sessions) loop() {
	for {
		oldest := time.Now().Add(time.Second * -SessionTimeout).Unix()
		if err := s.Purge(oldest); err != nil {
			if err == bolt.ErrDatabaseNotOpen {
				return
			}

			s.out.Error("error purging: %v", err)
		}

		time.Sleep(time.Minute)
	}
}

// purge will purge all entries oldest than the oldest value
func (s *Sessions) purge(txn *core.Transaction, oldest int64) (err error) {
	err = txn.ForEach(func(sessionID string, val core.Value) (err error) {
		session := val.(*Session)
		if session.LastUsedAt >= oldest {
			return
		}

		return txn.Remove(sessionID)
	})

	return
}

// Purge will purge all entries oldest than the oldest value
func (s *Sessions) Purge(oldest int64) (err error) {
	err = s.c.Transaction(func(txn *core.Transaction) (err error) {
		return s.purge(txn, oldest)
	})

	return
}

// New will create a new token/key pair
func (s *Sessions) New(userID string) (key, token string, err error) {
	// Set key/token
	key, token = s.newKeyToken()
	// Create new session
	session := s.newSession(key, token, userID)

	if _, err = s.c.New(&session); err != nil {
		key = ""
		token = ""
		return
	}

	return
}

// Get will retrieve the user id associated with a provided key/token pair
func (s *Sessions) Get(key, token string) (userID string, err error) {
	// Create session key from the key/token pair
	sessionKey := newSessionKey(key, token)
	err = s.c.Transaction(func(txn *core.Transaction) (err error) {
		var sp *Session
		if sp, err = s.getByKey(txn, sessionKey); err != nil {
			return
		}

		// Set uuid as session uuid
		userID = sp.UserID
		// Set last action for session
		sp.setAction()
		return txn.Edit(sp.ID, sp)
	})

	return
}

// Remove will invalidate a provided key/token pair session
func (s *Sessions) Remove(key, token string) (err error) {
	// Create session key from the key/token pair
	sessionKey := newSessionKey(key, token)
	err = s.c.Transaction(func(txn *core.Transaction) (err error) {
		var sp *Session
		if sp, err = s.getByKey(txn, sessionKey); err != nil {
			return
		}

		return txn.Remove(sp.ID)
	})

	return
}

// Close will close an instance of Sessions
func (s *Sessions) Close() (err error) {
	return s.c.Close()
}
