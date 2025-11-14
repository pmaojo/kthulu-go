// @kthulu:module:calendar
package domain

import (
	"fmt"
	"time"
)

// WorkingHours defines the working hours for appointment scheduling
type WorkingHours struct {
	Monday    DaySchedule `json:"monday"`
	Tuesday   DaySchedule `json:"tuesday"`
	Wednesday DaySchedule `json:"wednesday"`
	Thursday  DaySchedule `json:"thursday"`
	Friday    DaySchedule `json:"friday"`
	Saturday  DaySchedule `json:"saturday"`
	Sunday    DaySchedule `json:"sunday"`
}

// DaySchedule defines the schedule for a specific day
type DaySchedule struct {
	IsWorkingDay bool          `json:"isWorkingDay"`
	StartTime    time.Duration `json:"startTime"`  // Duration from midnight
	EndTime      time.Duration `json:"endTime"`    // Duration from midnight
	BreakStart   time.Duration `json:"breakStart"` // Optional break time
	BreakEnd     time.Duration `json:"breakEnd"`   // Optional break time
}

// DefaultWorkingHours returns standard Monday-Friday 9AM-5PM schedule
func DefaultWorkingHours() WorkingHours {
	return WorkingHours{
		Monday:    DaySchedule{IsWorkingDay: true, StartTime: 9 * time.Hour, EndTime: 17 * time.Hour},
		Tuesday:   DaySchedule{IsWorkingDay: true, StartTime: 9 * time.Hour, EndTime: 17 * time.Hour},
		Wednesday: DaySchedule{IsWorkingDay: true, StartTime: 9 * time.Hour, EndTime: 17 * time.Hour},
		Thursday:  DaySchedule{IsWorkingDay: true, StartTime: 9 * time.Hour, EndTime: 17 * time.Hour},
		Friday:    DaySchedule{IsWorkingDay: true, StartTime: 9 * time.Hour, EndTime: 17 * time.Hour},
		Saturday:  DaySchedule{IsWorkingDay: false},
		Sunday:    DaySchedule{IsWorkingDay: false},
	}
}

// GetDaySchedule returns the schedule for a specific weekday
func (wh *WorkingHours) GetDaySchedule(weekday time.Weekday) DaySchedule {
	switch weekday {
	case time.Monday:
		return wh.Monday
	case time.Tuesday:
		return wh.Tuesday
	case time.Wednesday:
		return wh.Wednesday
	case time.Thursday:
		return wh.Thursday
	case time.Friday:
		return wh.Friday
	case time.Saturday:
		return wh.Saturday
	case time.Sunday:
		return wh.Sunday
	default:
		return DaySchedule{IsWorkingDay: false}
	}
}

// IsWorkingDay checks if a given date is a working day
func (wh *WorkingHours) IsWorkingDay(date time.Time) bool {
	schedule := wh.GetDaySchedule(date.Weekday())
	return schedule.IsWorkingDay
}

// GetWorkingHoursForDate returns the working hours for a specific date
func (wh *WorkingHours) GetWorkingHoursForDate(date time.Time, timezone *time.Location) (start, end time.Time, hasBreak bool, breakStart, breakEnd time.Time) {
	schedule := wh.GetDaySchedule(date.Weekday())
	if !schedule.IsWorkingDay {
		return time.Time{}, time.Time{}, false, time.Time{}, time.Time{}
	}

	if timezone == nil {
		timezone = time.UTC
	}

	year, month, day := date.Date()

	start = time.Date(year, month, day, 0, 0, 0, 0, timezone).Add(schedule.StartTime)
	end = time.Date(year, month, day, 0, 0, 0, 0, timezone).Add(schedule.EndTime)

	if schedule.BreakStart > 0 && schedule.BreakEnd > 0 {
		breakStart = time.Date(year, month, day, 0, 0, 0, 0, timezone).Add(schedule.BreakStart)
		breakEnd = time.Date(year, month, day, 0, 0, 0, 0, timezone).Add(schedule.BreakEnd)
		hasBreak = true
	}

	return start, end, hasBreak, breakStart, breakEnd
}

// ValidateAppointmentTime checks if an appointment time is valid
func (wh *WorkingHours) ValidateAppointmentTime(startTime, endTime time.Time, timezone *time.Location) error {
	// Check if it's a working day
	if !wh.IsWorkingDay(startTime) {
		return fmt.Errorf("appointment cannot be scheduled on non-working day")
	}

	// Check if the appointment is within working hours
	workStart, workEnd, hasBreak, breakStart, breakEnd := wh.GetWorkingHoursForDate(startTime, timezone)

	if startTime.Before(workStart) || endTime.After(workEnd) {
		return fmt.Errorf("appointment is outside working hours")
	}

	// Check if the appointment conflicts with break time
	if hasBreak {
		if startTime.Before(breakEnd) && endTime.After(breakStart) {
			return fmt.Errorf("appointment conflicts with break time")
		}
	}

	// Check if start time is before end time
	if !startTime.Before(endTime) {
		return fmt.Errorf("start time must be before end time")
	}

	// Check if appointment is in the future
	if startTime.Before(time.Now()) {
		return fmt.Errorf("appointment cannot be scheduled in the past")
	}

	return nil
}
