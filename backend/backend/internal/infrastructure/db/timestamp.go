package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Timestamp handles automatic conversion between SQLite strings and Go time.Time
// Similar to Rails ActiveRecord timestamps
type Timestamp struct {
	time.Time
}

// NewTimestamp creates a new Timestamp with current time
func NewTimestamp() Timestamp {
	return Timestamp{Time: time.Now()}
}

// Scan implements the Scanner interface for reading from database
func (t *Timestamp) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		if v == "" {
			t.Time = time.Time{}
			return nil
		}
		// Try multiple formats that SQLite might use
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			time.RFC3339,
		}

		for _, format := range formats {
			if parsed, err := time.Parse(format, v); err == nil {
				t.Time = parsed
				return nil
			}
		}
		return fmt.Errorf("cannot parse time string: %s", v)
	case []byte:
		return t.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan %T into Timestamp", value)
	}
}

// Value implements the driver Valuer interface for writing to database
func (t Timestamp) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	// Always store in SQLite-friendly format
	return t.Time.Format("2006-01-02 15:04:05"), nil
}

// NullableTimestamp handles nullable timestamps
type NullableTimestamp struct {
	Time  time.Time
	Valid bool
}

// NewNullableTimestamp creates a new NullableTimestamp
func NewNullableTimestamp(t time.Time) NullableTimestamp {
	return NullableTimestamp{Time: t, Valid: !t.IsZero()}
}

// Scan implements the Scanner interface
func (nt *NullableTimestamp) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}

	var t Timestamp
	if err := t.Scan(value); err != nil {
		return err
	}

	nt.Time = t.Time
	nt.Valid = !t.Time.IsZero()
	return nil
}

// Value implements the driver Valuer interface
func (nt NullableTimestamp) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return Timestamp{Time: nt.Time}.Value()
}
