package jump

import (
	"fmt"

	"github.com/Hatch1fy/errors"
	"github.com/Hatch1fy/httpserve"
)

// NewResourceKey will return a new resource key from a given resource name and resource ID
// Note: Providing an empty resourceID will treat the resource as a grouping resource (No ID association)
func NewResourceKey(resourceName, resourceID string) (resourceKey string) {
	if len(resourceID) == 0 {
		return resourceName
	}

	return fmt.Sprintf("%s::%s", resourceName, resourceID)
}

func getUserID(ctx *httpserve.Context) (userID string, err error) {
	if userID = ctx.Get("userID"); len(userID) == 0 {
		err = errors.Error("cannot assert permissions, userID is empty")
		return
	}

	return
}

func getAPIKey(ctx *httpserve.Context) (apiKey string) {
	q := ctx.Request.URL.Query()

	if apiKey = q.Get("apiKey"); len(apiKey) > 0 {
		return
	}

	var (
		vals []string
		ok   bool
	)

	if vals, ok = ctx.Request.Header["X-Api-Key"]; !ok {
		return
	}

	if len(vals) == 0 {
		return
	}

	apiKey = vals[0]
	return
}
