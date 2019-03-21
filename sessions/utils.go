package sessions

func newSessionKey(key, token string) (mapkey string) {
	return key + "::" + token
}
