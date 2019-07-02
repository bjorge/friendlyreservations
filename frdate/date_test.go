package frdate

import (
	"testing"
)

func TestDate(t *testing.T) {

	_, err := NewDateBuilder("America/Los_Angeles1")
	if err == nil {
		t.Fatal("expected an error with bad time zone")
	}

	b, err := NewDateBuilder("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	_, err = b.NewDate("2018-11-1")
	if err == nil {
		t.Fatalf("expected an error with bad iso date input")
	}

	testDate1 := "2018-11-01"
	time1, err := b.NewDate(testDate1)
	if err != nil {
		t.Fatal(err)
	}

	if testDate1 != time1.ToString() {
		t.Fatalf("date ToString() failed")
	}

	time2, _ := b.NewDate("2018-11-02")

	if time1.After(time2) {
		t.Fatalf("After logic does not work")
	}

	if time2.Before(time1) {
		t.Fatalf("Before logic does not work")
	}

	days, err := DaysList(time1, time2, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 2 {
		t.Fatalf("expected 2 days")
	}

	days, _ = DaysList(time1, time2, false)
	if len(days) != 1 {
		t.Fatalf("expected 1 day")
	}

	_, err = DaysList(time2, time1, false)
	if err == nil {
		t.Fatalf("expected an error")
	}

}

func TestDateMonthInOutNormal(t *testing.T) {
	b, err := NewDateBuilder("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	month, _ := b.NewDate("2001-02-03")

	in, out := month.MonthInOut()

	t.Logf("in %+v out %+v", in.ToString(), out.ToString())

	if in.ToString() != "2001-02-01" {
		t.Fatalf("wrong in date of month, got: %+v", in.ToString())
	}

	if out.ToString() != "2001-03-01" {
		t.Fatalf("wrong out date of month")
	}
}

func TestDateMonthInOutSpecial(t *testing.T) {
	// from time.Date spec:
	// AddDate normalizes its result in the same way that Date does,
	// so, for example, adding one month to October 31 yields
	// December 1, the normalized form for November 31.

	b, err := NewDateBuilder("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	month, _ := b.NewDate("2001-10-31")

	in, out := month.MonthInOut()

	t.Logf("in %+v out %+v", in.ToString(), out.ToString())

	if in.ToString() != "2001-10-01" {
		t.Fatalf("wrong in date of month, got: %+v", in.ToString())
	}

	if out.ToString() != "2001-11-01" {
		t.Fatalf("wrong out date of month")
	}

}

func dateOverLapResult(t *testing.T, inDate1 string, outDate1 string, inDate2 string, outDate2 string, expected bool, name string) {
	b, _ := NewDateBuilder("America/Los_Angeles")

	inDateA, _ := b.NewDate(inDate1)
	outDateA, _ := b.NewDate(outDate1)
	inDateB, _ := b.NewDate(inDate2)
	outDateB, _ := b.NewDate(outDate2)

	if overlap := DateOverlapInOut(inDateA, outDateA, inDateB, outDateB); overlap != expected {
		t.Fatalf("mismatch overlap expected (%+v) got (%+v) test %+v", expected, overlap, name)
	}
}

func TestDateOverLapInOut(t *testing.T) {
	dateOverLapResult(t, "2018-01-01", "2019-01-01", "2019-01-01", "2020-01-01", false, "a")
	dateOverLapResult(t, "2019-01-01", "2020-01-01", "2018-01-01", "2019-01-01", false, "b")
	dateOverLapResult(t, "2018-01-01", "2019-01-02", "2019-01-01", "2020-01-01", true, "c")
	dateOverLapResult(t, "2019-01-01", "2020-01-01", "2018-01-01", "2019-01-02", true, "d")
}

func TestDateTime(t *testing.T) {
	b, _ := NewDateBuilder("America/Los_Angeles")

	_, err := b.NewDateTime("2018-11-1")
	if err == nil {
		t.Fatalf("expected an error with bad rfc date input")
	}

	// 2018-10-22T17:53:16Z
	testDateTime := "2018-10-22T17:53:16Z"
	dateTime, err := b.NewDateTime(testDateTime)
	if err != nil {
		t.Fatal(err)
	}

	formattedDateTime := dateTime.ToString()

	if formattedDateTime != testDateTime {
		t.Fatalf("format error, got %+v and expected %+v", formattedDateTime, testDateTime)
	}

}
