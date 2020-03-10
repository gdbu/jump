package jump

import "github.com/Hatch1fy/apikeys"

// GetAPIKeysByUser will return the api keys for a user
func (j *Jump) GetAPIKeysByUser(userID string) (as []*apikeys.APIKey, err error) {
	return j.api.GetByUser(userID)
}
