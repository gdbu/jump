package sessions

import (
	"sort"
	"time"

	"github.com/Hatch1fy/errors"
	core "github.com/Hatch1fy/service-core"
	"github.com/Hatch1fy/uuid"
	"github.com/boltdb/bolt"
	"github.com/hatchify/scribe"
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
	s.out = scribe.New("Sessions")
	if s.c, err = core.New("sessions", dir, &Session{}, relationships...); err != nil {
		return
	}

	s.g = uuid.NewGenerator()

	// Start purge loop
	go s.loop()
	sp = &s
	return
}

// Sessions manages sessions
type Sessions struct {
	out *scribe.Scribe
	c   *core.Core
	g   *uuid.Generator
}

func (s *Sessions) newKeyToken() (key, token string) {
	// Set key
	var id = s.g.New()
	key = id.String()

	// Set token
	id = s.g.New()
	token = id.String()
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

func (s *Sessions) getByUserID(txn *core.Transaction, userID string) (ss []*Session, err error) {
	if err = txn.GetByRelationship(relationshipUsers, userID, &ss); err != nil {
		return
	}

	sort.Slice(ss, func(i, j int) (less bool) {
		// We are sorting descending, so we inverse the lookup
		return ss[i].LastUsedAt > ss[j].LastUsedAt
	})

	return
}

func (s *Sessions) loop() {
	for {
		oldest := time.Now().Add(time.Second * -SessionTimeout).Unix()
		if err := s.Purge(oldest); err != nil {
			if err == bolt.ErrDatabaseNotOpen {
				return
			}

			s.out.Errorf("error purging: %v", err)
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

// Remove will invalidate a provided key/token pair session
func (s *Sessions) invalidateUser(txn *core.Transaction, userID string) (err error) {
	var ss []*Session
	if ss, err = s.getByUserID(txn, userID); err != nil {
		return
	}

	for _, sess := range ss {
		if err = txn.Remove(sess.ID); err != nil {
			return
		}
	}

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

	if err = s.c.Batch(func(txn *core.Transaction) (err error) {
		_, err = txn.New(&session)
		return
	}); err != nil {
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
	err = s.c.Batch(func(txn *core.Transaction) (err error) {
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

// GetByUserID will retrieve all the sessions for a given user ID
func (s *Sessions) GetByUserID(userID string) (ss []*Session, err error) {
	err = s.c.ReadTransaction(func(txn *core.Transaction) (err error) {
		ss, err = s.getByUserID(txn, userID)
		return
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

// InvalidateUser will invalidate all sessions associated with a user
func (s *Sessions) InvalidateUser(userID string) (err error) {
	err = s.c.Transaction(func(txn *core.Transaction) (err error) {
		return s.invalidateUser(txn, userID)
	})

	return
}

// Close will close an instance of Sessions
func (s *Sessions) Close() (err error) {
	return s.c.Close()
}
