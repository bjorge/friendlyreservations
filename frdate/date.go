/*
Package frdate has functions to manage reservation dates and record timestamps.

In general the fr code should not use time package but only these functions.

Also this package should not accept or return "time" objects.

Date or DateTime objects are used for calendar dates and timestamps respectively.

Date or DateTime objects are built with a DataBuilder which requires the location.

Date objects are persisted with strings in the iso8601format layout (see below).

DateTime objects are persisted with strings in the time.RFC3339 layout (see time package).

*/
package frdate

import (
	"errors"
	"fmt"
	"time"
)

// DateBuilder is the base struct for creating Date or DateTime objects
type DateBuilder struct {
	loc time.Location
}

// Date represents a calendar date
type Date struct {
	// t is the the date at time 00:00:00
	t time.Time
}

// DateTime represents a date plus time, ex. for use as a timestamp
type DateTime struct {
	t time.Time
}

// iso8601format is the format convention for storing a Date
const iso8601format = "2006-01-02"

// TestTimeOffsetDays is the time offset used during testing
var TestTimeOffsetDays *int

// NewDateBuilder creates a new DateBuilder given a location or returns an error
func NewDateBuilder(location string) (*DateBuilder, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		return nil, err
	}
	return &DateBuilder{*loc}, nil
}

// MustNewDateBuilder creates a new DateBuilder given a location or panics on error
func MustNewDateBuilder(location string) *DateBuilder {
	loc, err := time.LoadLocation(location)
	if err != nil {
		panic(err)
	}
	return &DateBuilder{*loc}
}

// NewDate produces a Date given the date string, or returns an error
func (r *DateBuilder) NewDate(iso8601short string) (*Date, error) {
	t, err := time.ParseInLocation(iso8601format, iso8601short, &r.loc)
	if err != nil {
		return nil, err
	}
	return &Date{t}, nil
}

// MustNewDate produces a Date given the date string, or panics
func (r *DateBuilder) MustNewDate(iso8601short string) *Date {
	t, err := time.ParseInLocation(iso8601format, iso8601short, &r.loc)
	if err != nil {
		panic(err)
	}
	return &Date{t}
}

// NewDateTime produces a DateTime given the date-time string, or returns an error
func (r *DateBuilder) NewDateTime(rfc3339datetime string) (*DateTime, error) {
	t, err := time.ParseInLocation(time.RFC3339, rfc3339datetime, &r.loc)
	if err != nil {
		return nil, err
	}
	return &DateTime{t}, nil
}

// MustNewDateTime produces a DateTime given the date-time string, or panics
func (r *DateBuilder) MustNewDateTime(rfc3339datetime string) *DateTime {
	t, err := time.ParseInLocation(time.RFC3339, rfc3339datetime, &r.loc)
	if err != nil {
		panic(err)
	}
	return &DateTime{t}
}

// MustNewDate produces a Date given a location and date string or panics on error
func MustNewDate(location string, iso8601short string) *Date {
	builder, err := NewDateBuilder(location)
	if err != nil {
		panic(err)
	}
	date, err := builder.NewDate(iso8601short)
	if err != nil {
		panic(err)
	}
	return date
}

// Year returns the year portion of a Date
func (r *Date) Year() int {
	return r.t.Year()
}

// ToString returns the iso string representation of a Date
func (r *Date) ToString() string {
	return r.t.Format(iso8601format)
}

// ToStringPtr returns a pointer to an iso string representation of a Date
func (r *Date) ToStringPtr() *string {
	isoString := r.t.Format(iso8601format)
	return &isoString
}

// ToString returns the rfc string representation of a DateTime
func (r *DateTime) ToString() string {
	return r.t.Format(time.RFC3339)
}

// ToStringPtr returns a pointer to an rfc string representation of a DateTime
func (r *DateTime) ToStringPtr() *string {
	rfcString := r.t.Format(time.RFC3339)
	return &rfcString
}

// Today produces today's Date
func (r *DateBuilder) Today() *Date {
	now := time.Now().UTC()
	if TestTimeOffsetDays != nil {
		text := fmt.Sprintf("%dh", *TestTimeOffsetDays*24)
		duration, _ := time.ParseDuration(text)
		now = now.Add(duration)
	}
	isoString := now.In(&r.loc).Format(iso8601format)
	date, _ := r.NewDate(isoString)
	return date
}

// CreateDateTimeUTC is used for all timestamps in the project
func CreateDateTimeUTC() string {
	now := time.Now().UTC()
	if TestTimeOffsetDays != nil {
		text := fmt.Sprintf("%dh", *TestTimeOffsetDays*24)
		duration, _ := time.ParseDuration(text)
		now = now.Add(duration)
	}
	return now.Format(time.RFC3339)
}

// Before is true if the passed Date is before the receiver Date
func (r *Date) Before(date *Date) bool {
	return r.t.Before(date.t)
}

// After is true if the passed Date is after the receiver Date
func (r *Date) After(date *Date) bool {
	return r.t.After(date.t)
}

// Equal is true if the passed Date is equal to the receiver Date
func (r *Date) Equal(date *Date) bool {
	return r.t.Equal(date.t)
}

// Before is true if the passed DateTime is before the receiver DateTime
func (r *DateTime) Before(date *DateTime) bool {
	return r.t.Before(date.t)
}

// After is true if the passed DateTime is after the receiver DateTime
func (r *DateTime) After(date *DateTime) bool {
	return r.t.After(date.t)
}

// Equal is true if the passed DateTime is equal to the receiver DateTime
func (r *DateTime) Equal(date *DateTime) bool {
	return r.t.Equal(date.t)
}

// Copy returns a copy of the receiver Date
func (r *Date) Copy() *Date {
	newTime := r.t
	return &Date{newTime}
}

// Format produces a string representation of a Date based on the layout argument
func (r *Date) Format(layout string) string {
	return r.t.Format(layout)
}

// Format produces a string representation of a DateTime based on the layout argument
func (r *DateTime) Format(layout string) string {
	return r.t.Format(layout)
}

// ToDate returns the Date of a DateTime
func (r *DateTime) ToDate() *Date {
	year, month, day := r.t.Date()
	// round to exactly midnight
	newTime := time.Date(year, month, day, 0, 0, 0, 0, r.t.Location())
	return &Date{newTime}
}

// AddDays returns the Date after num days
func (r *Date) AddDays(num int) *Date {
	newTime := r.t.AddDate(0, 0, num)
	year, month, day := newTime.Date()
	// round to exactly midnight
	newTime = time.Date(year, month, day, 0, 0, 0, 0, newTime.Location())
	return &Date{newTime}
}

// AddDays returns the DateTime after num days
func (r *DateTime) AddDays(num int) *DateTime {
	newTime := r.t.AddDate(0, 0, num)
	year, month, day := newTime.Date()
	// round to exactly midnight
	newTime = time.Date(year, month, day, 0, 0, 0, 0, newTime.Location())
	return &DateTime{newTime}
}

// DaysList returns a list of Date's inclusive or exclusive of the last Date argument
func DaysList(first *Date, last *Date, includeLast bool) ([]*Date, error) {

	if last.Before(first) {
		return nil, errors.New("last " + last.ToString() + " before first " + first.ToString())
	}

	days := []*Date{}

	iterator := first.Copy()

	for iterator.Before(last) {
		days = append(days, iterator.Copy())
		iterator = iterator.AddDays(1)
	}

	if includeLast {
		days = append(days, iterator.Copy())
	}

	return days, nil
}

// MonthInOut produces the first Date of the current month, and the first Date of the following month
func (r *Date) MonthInOut() (*Date, *Date) {
	year, month, _ := r.t.Date()
	first := time.Date(year, month, 1, 0, 0, 0, 0, r.t.Location())
	last := first.AddDate(0, 1, 0)
	return &Date{first}, &Date{last}
}

// Sub returns the number of days between two dates
func (r *Date) Sub(start *Date) int {
	duration := r.t.Sub(start.t)
	return int(duration.Hours() / 24.0)
}

// YearInOut produces the first Date of the current year and the first Date of the following year
func (r *Date) YearInOut() (*Date, *Date) {
	year, _, _ := r.t.Date()
	first := time.Date(year, time.January, 1, 0, 0, 0, 0, r.t.Location())
	last := first.AddDate(1, 0, 0)
	return &Date{first}, &Date{last}
}

// DateOverlap returns true if there exists an overlap between the first and second date ranges
func DateOverlap(first1 *Date, last1 *Date, first2 *Date, last2 *Date) bool {
	// ----F----L------ (1) or (2)
	// ----------F--L-- (2) or (1)

	// ----F----L------ (1) or (2)
	// -F-L------------ (2) or (1)

	return !last1.Before(first2) && !last2.Before(first1)
}

// DateOverlapInOut returns true if there exists an overlap between the first and last in-out date ranges
func DateOverlapInOut(inDate1 *Date, outDate1 *Date, inDate2 *Date, outDate2 *Date) bool {
	lastInDate1 := outDate1.AddDays(-1)
	firstOutDate1 := inDate1.AddDays(1)

	lastInDate2 := outDate2.AddDays(-1)
	firstOutDate2 := inDate2.AddDays(1)

	overlap1 := DateOverlap(inDate1, lastInDate1, inDate2, lastInDate2)
	overlap2 := DateOverlap(firstOutDate1, outDate1, firstOutDate2, outDate2)

	return overlap1 || overlap2
}
