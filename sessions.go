package jump

// NewSession will generate a new session for a given user ID
func (j *Jump) NewSession(userID string) (key, token string, err error) {
	return j.sess.New(userID)
}

// TODO: Clear session
