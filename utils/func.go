package utils

import "time"

type ShouldReportExpiry = func(time.Time) bool

// WithExpiringWithin returns a function that checks if a given expiry time is within the configured expiry window
func WithExpiringWithin(days int, includeExpired bool) ShouldReportExpiry {
	now := time.Now().UTC()

	// If days is 0, include all non-expired items, and optionally include expired items
	if days <= 0 {
		return func(expiry time.Time) bool {
			return includeExpired || !IsExpired(expiry, now)
		}
	}

	// Otherwise, include items that are expiring within the specified number of days, and optionally include expired items
	return func(expiry time.Time) bool {
		if includeExpired && IsExpired(expiry, now) {
			return true
		}
		return expiry.Before(now.Add(time.Duration(days)*24*time.Hour)) && !expiry.Before(now)
	}
}

// IsExpiredNow checks if the given expiry time is before the current time (i.e., if it has already expired)
func IsExpiredNow(expiry time.Time) bool {
	return IsExpired(expiry, time.Now().UTC())
}

// IsExpired checks if the given expiry time is before the specified time
func IsExpired(expiry time.Time, now time.Time) bool {
	return expiry.Before(now)
}
