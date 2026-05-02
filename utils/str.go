package utils

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
)

// FormatPrettyExpiry returns a string indicating when the given time expires, and if it is already expired.
func FormatPrettyExpiry(t time.Time) string {
	msg := "expires"
	expired := IsExpiredNow(t)

	// If the time is already expired, change the message and mark it as expired.
	if expired {
		msg = "EXPIRED"
	}

	s := fmt.Sprintf("%s on %s", msg, t.Format("2006-01-02"))

	// If the time is already expired, color the string red and make it bold.
	if expired {
		s = text.Colors{
			text.FgRed,
			text.Bold,
		}.Sprint(s)
	}
	return s
}
