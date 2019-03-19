package sessions

import (
	"time"
)

func newSession(userID string) (s session) {
	s.UserID = userID
	s.setAction()
	return
}

type session struct {
	// UserID of the user who owns this session
	UserID string `json:"userID"`
	// Last action taken for this session
	LastAction int64 `json:"lastAction"`
}

func (s *session) setAction() {
	s.LastAction = time.Now().Unix()
}
