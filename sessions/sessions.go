package sessions

import (
	"encoding/json"
	"path"
	"time"

	"github.com/Hatch1fy/errors"
	"github.com/boltdb/bolt"
	"github.com/missionMeteora/journaler"
	"github.com/missionMeteora/uuid"
)

const (
	// ErrSessionDoesNotExist is returned when an invalid token/key pair is presented
	ErrSessionDoesNotExist = errors.Error("session with that token/key pair does not exist")
	// ErrNotInitialized is returned when actions are performed on a non-initialized instance of Retargeting
	ErrNotInitialized = errors.Error("sessions library has not been properly initialized")
)

const (
	// SessionTimeout (in seconds) is the ttl for sessions, an action will refresh the duration
	SessionTimeout = 60 * 60 * 12 // 12 hours
)

var (
	sessionsBktKey = []byte("sessions")
)

// New will return a new instance of sessions
func New(dir string) (sp *Sessions, err error) {
	var s Sessions
	if s.db, err = bolt.Open(path.Join(dir, "sessions.bdb"), 0644, nil); err != nil {
		return
	}

	if err = s.init(); err != nil {
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
	db  *bolt.DB
	g   *uuid.Gen
	out *journaler.Journaler
}

func (s *Sessions) init() (err error) {
	err = s.db.Update(func(txn *bolt.Tx) (err error) {
		if _, err = txn.CreateBucketIfNotExists(sessionsBktKey); err != nil {
			return
		}

		return
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

			s.out.Error("error purging: %v", err)
		}

		time.Sleep(time.Minute)
	}
}

func (s *Sessions) get(txn *bolt.Tx, sessionKey []byte) (sp *session, err error) {
	var bkt *bolt.Bucket
	if bkt = txn.Bucket(sessionsBktKey); bkt == nil {
		return nil, ErrNotInitialized
	}

	var bs []byte
	if bs = bkt.Get(sessionKey); len(bs) == 0 {
		err = ErrSessionDoesNotExist
		return
	}

	var session session
	if err = json.Unmarshal(bs, &session); err != nil {
		return
	}

	sp = &session
	return
}

func (s *Sessions) put(txn *bolt.Tx, sessionKey []byte, sp *session) (err error) {
	var bkt *bolt.Bucket
	if bkt = txn.Bucket(sessionsBktKey); bkt == nil {
		return ErrNotInitialized
	}

	var bs []byte
	if bs, err = json.Marshal(sp); err != nil {
		return
	}

	return bkt.Put(sessionKey, bs)
}

func (s *Sessions) delete(txn *bolt.Tx, sessionKey []byte) (err error) {
	var bkt *bolt.Bucket
	if bkt = txn.Bucket(sessionsBktKey); bkt == nil {
		return ErrNotInitialized
	}

	return bkt.Delete(sessionKey)
}

// purge will purge all entries oldest than the oldest value
func (s *Sessions) purge(txn *bolt.Tx, oldest int64) (err error) {
	var bkt *bolt.Bucket
	if bkt = txn.Bucket(sessionsBktKey); bkt == nil {
		return ErrNotInitialized
	}

	return bkt.ForEach(func(key, bs []byte) (err error) {
		var session session
		if err = json.Unmarshal(bs, &session); err != nil {
			return
		}

		if session.LastAction >= oldest {
			return
		}

		return bkt.Delete(key)
	})
}

// Purge will purge all entries oldest than the oldest value
func (s *Sessions) Purge(oldest int64) (err error) {
	err = s.db.Update(func(txn *bolt.Tx) (err error) {
		return s.purge(txn, oldest)
	})

	return
}

// New will create a new token/key pair
func (s *Sessions) New(userID string) (key, token string, err error) {
	var session session
	session.UserID = userID
	session.setAction()

	// Set key
	key = s.g.New().String()
	// Set token
	token = s.g.New().String()
	// Set session key
	sessionKey := newSessionKey(key, token)

	err = s.db.Update(func(txn *bolt.Tx) (err error) {
		return s.put(txn, []byte(sessionKey), &session)
	})

	return
}

// Get will retrieve the user id associated with a provided key/token pair
func (s *Sessions) Get(key, token string) (userID string, err error) {
	// Create session key from the key/token pair
	sessionKey := newSessionKey(key, token)
	// Get byteslice version of sessionKey
	sessionKeyBytes := []byte(sessionKey)
	err = s.db.Update(func(txn *bolt.Tx) (err error) {
		var sp *session
		if sp, err = s.get(txn, sessionKeyBytes); err != nil {
			return
		}

		// Set uuid as session uuid
		userID = sp.UserID
		// Set last action for session
		sp.setAction()
		return s.put(txn, sessionKeyBytes, sp)
	})

	return
}

// Delete will invalidate a provided key/token pair session
func (s *Sessions) Delete(key, token string) (err error) {
	// Create session key from the key/token pair
	sessionKey := newSessionKey(key, token)
	// Get byteslice version of sessionKey
	sessionKeyBytes := []byte(sessionKey)
	err = s.db.Update(func(txn *bolt.Tx) (err error) {
		return s.delete(txn, sessionKeyBytes)
	})

	return
}

// Close will close an instance of Sessions
func (s *Sessions) Close() (err error) {
	return s.db.Close()
}
