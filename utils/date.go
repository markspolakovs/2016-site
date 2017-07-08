package utils

import (
	"fmt"
	"strconv"
	"time"
)

//
// Calculations involving schedule hours and start-of-day.
//

// The hour of the day (local time) at which the scheduled day begins.
const startHour = 6

// StartOfDayOn gets the schedule start-of-day on a given date.
// This is in terms of startHour.
func StartOfDayOn(date time.Time) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, startHour, 0, 0, 0, time.Local)
}

// StartOffsetToHour takes a number of hours since the last day start (0-23) and gives the actual hour.
// It returns an error if the hour is invalid.
func StartOffsetToHour(hour int) (int, error) {
	if 23 < hour || hour < 0 {
		return 0, fmt.Errorf("StartOffsetToHour: hour %d not between 0 and 23", hour)
	}
	return (hour + startHour) % 24, nil
}

// HourToStartOffset takes an hour (0-23) and gives the number of hours elapsed since the last day start.
// It returns an error if the hour is invalid.
func HourToStartOffset(hour int) (int, error) {
	if 23 < hour || hour < 0 {
		return 0, fmt.Errorf("HourToStartOffset: hour %d not between 0 and 23", hour)
	}
	// Adding 24 to ensure we don't go negative.  Negative modulo is scary.
	return ((hour + 24) - startHour) % 24, nil
}

//
// Conversions from ISO year-week to more amenable formats.
//

// ParseIsoWeek parses an ISO weekday from year, week, and weekday strings.
// It performs bounds checking.
func ParseIsoWeek(year, week, weekday string) (int, int, time.Weekday, error) {
	y, err := strconv.Atoi(year)
	if err != nil {
		return 0, 0, 0, err
	}
	if y < 0 {
		return 0, 0, 0, fmt.Errorf("Invalid year: %d", y)
	}

	w, err := strconv.Atoi(week)
	if err != nil {
		return 0, 0, 0, err
	}
	if w < 1 || 53 < w {
		return 0, 0, 0, fmt.Errorf("Invalid week: %d", w)
	}

	// Two-stage conversion: first to int, then to Weekday.
	// Go treats Sunday as day 0: we must correct this grave mistake.
	dI, err := strconv.Atoi(weekday)
	if err != nil {
		return 0, 0, 0, err
	}
	if dI < 1 || 7 < dI {
		return 0, 0, 0, fmt.Errorf("Invalid day: %d", dI)
	}

	var d time.Weekday
	if dI == 7 {
		d = time.Sunday
	} else {
		d = time.Weekday(dI)
	}

	return y, w, d, nil
}

// IsoWeekToDate interprets year, week, and weekday strings as an ISO weekday.
// The time is set to local midnight.
func IsoWeekToDate(year, week int, weekday time.Weekday) (time.Time, error) {
	// This is based on the calculation given at:
	// https://en.wikipedia.org/wiki/ISO_week_date#Calculating_a_date_given_the_year.2C_week_number_and_weekday

	// We need to find the first week in the year.
	// This always contains the 4th of January, so find that, and get
	// ISOWeek on it.
	fj := time.Date(year, time.January, 4, 0, 0, 0, 0, time.Local)

	// Correct Go's stupid Sunday is 0 decision, making the weekdays ISO 8601 compliant
	intWeekday := int(weekday)
	if intWeekday == 0 {
		intWeekday = 7
	}
	fjWeekday := int(fj.Weekday())
	if fjWeekday == 0 {
		fjWeekday = 7
	}

	// Sanity check to make sure time (and our intuition) is still working.
	fjYear, fjWeek := fj.ISOWeek()
	if fjYear != year {
		return time.Time{}, fmt.Errorf("ISO weekday year %d != calendar year %d", fjYear, year)
	}
	if fjWeek != 1 {
		return time.Time{}, fmt.Errorf("ISO weekday week of 4 Jan (%d) not week 1", fjWeek)
	}

	// The ISO 8601 ordinal date, which may belong to the next or previous
	// year.
	ord := (week * 7) + intWeekday - (fjWeekday + 3)

	// The ordinal date is just the number of days since 1 Jan y plus one,
	// so calculate the year from that.
	oj := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	return oj.AddDate(0, 0, ord-1), nil
}

// MostRecentMonday returns the most recent Monday before d.
func MostRecentMonday(d time.Time) time.Time {
	/* The weekday is the number of days since the most recent Sunday, so
	   shifting it by 1 modulo 7 gives us the correct result for Monday. */
	dmon := int(d.Weekday()) - 1
	if dmon < 0 {
		// Correct for Sunday
		dmon = 6
	}

	return d.AddDate(0, 0, -dmon)
}

// FormatWeekRelative pretty-prints the name of a week starting on start.
// start must be a Monday.
func FormatWeekRelative(start time.Time) string {
	/* If we're on the same week, or the week either end of current, we can (and
	   should) use short, human-friendly week names. */

	// To work out which week we're in, get the boundaries of last, this, and next week.
	tm := MostRecentMonday(time.Now())
	lm := tm.AddDate(0, 0, -7)
	nm := tm.AddDate(0, 0, 7)
	fm := tm.AddDate(0, 0, 14)

	switch {
	case start.Before(lm):
		break
	case start.Before(tm):
		return "last week"
	case start.Before(nm):
		return "this week"
	case start.Before(fm):
		return "next week"
	default:
		break
	}

	// If we got here, we can't give a fancy name to this week.
	sun := start.AddDate(0, 0, 6)
	return start.Format("02 Jan 2006") + " to " + sun.Format("02 Jan 2006")
}
