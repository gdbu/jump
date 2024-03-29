package sso

import (
	"time"

	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
)

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

func yesFilter(_ string) (ok bool, err error) {
	return true, nil
}

func wait(waitUntil time.Time, ch chan struct{}) (cancelled bool) {
	now := time.Now()
	duration := waitUntil.Sub(now)
	if duration <= 0 {
		return false
	}

	timer := time.NewTimer(duration)
	select {
	case <-timer.C:
		return false
	case <-ch:
		return true

	}
}

func notify(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}
