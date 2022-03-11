package sessions

import (
	"context"
	"sort"
	"time"

	"github.com/gdbu/scribe"
	"github.com/gdbu/uuid"
	"github.com/hatchify/errors"
	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
)

const (
	// ErrSessionDoesNotExist is returned when an invalid token/key pair is presented
	ErrSessionDoesNotExist = errors.Error("session with that token/key pair does not exist")
)

const (
	// SessionTimeout (in seconds) is the ttl for sessions, an action will refresh the duration
	SessionTimeout = 60 * 60 * 24 * 7 // 7 days
	RefreshPeriod  = 60 * 60 * 24     // 1 day
)

const (
	relationshipKeys  = "keys"
	relationshipUsers = "users"
)

var (
	relationships = []string{relationshipKeys, relationshipUsers}
)

// New will return a new instance of sessions
func New(opts mojura.Opts) (sp *Sessions, err error) {
	opts.Name = "sessions"

	var s Sessions
	s.out = scribe.New("Sessions")
	if s.c, err = mojura.New(opts, &Session{}, relationships...); err != nil {
		return
	}

	s.g = uuid.NewGenerator()

	if !opts.IsMirror {
		// Start purge loop
		go s.loop()
	}

	sp = &s
	return
}

// Sessions manages sessions
type Sessions struct {
	out *scribe.Scribe
	c   *mojura.Mojura
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

func (s *Sessions) getByKey(txn *mojura.Transaction, key string) (sp *Session, err error) {
	var entry Session
	filter := filters.Match(relationshipKeys, key)
	opts := mojura.NewIteratingOpts(filter)
	if err = txn.GetFirst(&entry, opts); err != nil {
		return
	}

	sp = &entry
	return
}

func (s *Sessions) getByUserID(txn *mojura.Transaction, userID string) (ss []*Session, err error) {
	filter := filters.Match(relationshipUsers, userID)
	opts := mojura.NewFilteringOpts(filter)
	if _, err = txn.GetFiltered(&ss, opts); err != nil {
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
			s.out.Errorf("error purging: %v", err)
		}

		time.Sleep(time.Minute)
	}
}

// purge will purge all entries older than the oldest value
func (s *Sessions) purge(txn *mojura.Transaction, oldest int64) (err error) {
	err = txn.ForEach(func(sessionID string, val mojura.Value) (err error) {
		session := val.(*Session)
		if session.LastUsedAt >= oldest {
			return
		}

		return txn.Remove(sessionID)
	}, nil)

	return
}

// Remove will invalidate a provided key/token pair session
func (s *Sessions) invalidateUser(txn *mojura.Transaction, userID string) (err error) {
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
	err = s.c.Transaction(context.Background(), func(txn *mojura.Transaction) (err error) {
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

	if err = s.c.Batch(context.Background(), func(txn *mojura.Transaction) (err error) {
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
func (s *Sessions) Get(key, token string) (sp *Session, err error) {
	// Create session key from the key/token pair
	sessionKey := newSessionKey(key, token)
	err = s.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		if sp, err = s.getByKey(txn, sessionKey); err != nil {
			return
		}

		return
	})

	now := time.Now().Unix()
	delta := now - sp.LastUsedAt
	if delta > RefreshPeriod {
		return
	}

	if err = s.Refesh(key, token); err != nil {
		sp = nil
		return
	}

	return
}

// Refesh will refresh a session
func (s *Sessions) Refesh(key, token string) (err error) {
	// Create session key from the key/token pair
	sessionKey := newSessionKey(key, token)
	err = s.c.Batch(context.Background(), func(txn *mojura.Transaction) (err error) {
		var sp *Session
		if sp, err = s.getByKey(txn, sessionKey); err != nil {
			return
		}

		// Set last action for session
		sp.setAction()
		return txn.Edit(sp.ID, sp)
	})

	return
}

// GetByUserID will retrieve all the sessions for a given user ID
func (s *Sessions) GetByUserID(userID string) (ss []*Session, err error) {
	err = s.c.ReadTransaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		ss, err = s.getByUserID(txn, userID)
		return
	})

	return
}

// Remove will invalidate a provided key/token pair session
func (s *Sessions) Remove(key, token string) (err error) {
	// Create session key from the key/token pair
	sessionKey := newSessionKey(key, token)
	err = s.c.Transaction(context.Background(), func(txn *mojura.Transaction) (err error) {
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
	err = s.c.Transaction(context.Background(), func(txn *mojura.Transaction) (err error) {
		return s.invalidateUser(txn, userID)
	})

	return
}

// Close will close an instance of Sessions
func (s *Sessions) Close() (err error) {
	return s.c.Close()
}
