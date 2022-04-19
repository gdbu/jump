package sessions

func makeSessionKey(key, token string) (mapkey string) {
	return key + "::" + token
}
