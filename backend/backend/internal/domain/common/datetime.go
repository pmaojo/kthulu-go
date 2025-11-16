package common

import "time"

// DateTimeFormat is the standard datetime format used throughout the application
const DateTimeFormat = "2006-01-02 15:04:05"

// Now returns the current time formatted as a string
func Now() string {
	return time.Now().Format(DateTimeFormat)
}

// ParseDateTime parses a datetime string into time.Time
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(DateTimeFormat, s)
}

// FormatDateTime formats a time.Time into our standard string format
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}
