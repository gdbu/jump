package sso

import (
	"fmt"
	"time"

	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
)

func asEntry(val mojura.Value) (e *Entry, err error) {
	var ok bool
	// Attempt to assert the value as an *Entry
	if e, ok = val.(*Entry); !ok {
		// Invalid type provided, return error
		err = fmt.Errorf("invalid entry type, expected %T and received %T", e, val)
		return
	}

	return
}

func getTS() time.Time {
	return time.Now().UTC()
}

func getTSString(delta time.Duration, layout string) string {
	return getTS().Add(delta).Format(layout)
}

// TODO: Utilize comparison filter for this
func newExpiredWithinPreviousHourFilter() mojura.Filter {
	previousHour := getTSString(time.Hour*-1, "15")
	return filters.Match(RelationshipExpiresAtHours, previousHour)
}

// TODO: Utilize comparison filter for this
func newExpiredWithinPreviousDayFilter() mojura.Filter {
	previousHour := getTSString(time.Hour*-24, "2006-01-02")
	return filters.Match(RelationshipExpiresAtHours, previousHour)
}
