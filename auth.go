package jump

// Login will attempt to login with a provided email and password combo
// If successful, a key/token pair will be returned to represent the session pair
func (j *Jump) Login(email, password string) (userID, key, token string, err error) {
	if userID, err = j.usrs.MatchEmail(email, password); err != nil {
		return
	}

	key, token, err = j.NewSession(userID)
	return
}

// Logout will invalidate the session of a given key/token pair
func (j *Jump) Logout(key, token string) (err error) {
	return j.sess.Remove(key, token)
}
