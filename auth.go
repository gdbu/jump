package jump

import "github.com/Hatch1fy/httpserve"

// Login will attempt to login with a provided email and password combo
// If successful, a key/token pair will be returned to represent the session pair
func (j *Jump) Login(ctx *httpserve.Context, email, password string) (userID string, err error) {
	if userID, err = j.usrs.MatchEmail(email, password); err != nil {
		return
	}

	err = j.NewSession(ctx, userID)
	return
}

// Logout will invalidate the session of a given key/token pair
func (j *Jump) Logout(key, token string) (err error) {
	return j.sess.Remove(key, token)
}
