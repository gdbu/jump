package jump

// Login will attempt to login with a provided email and password combo
// If successful, a key/token pair will be returned to represent the session pair
func (j *Jump) Login(email, password string) (key, token string, err error) {
	var userID string
	if userID, err = j.usrs.MatchEmail(email, password); err != nil {
		return
	}

	return j.NewSession(userID)
}
