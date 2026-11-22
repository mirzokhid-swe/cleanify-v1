package models

import (
	"testing"
	"time"
)

func d(s string) time.Time {
	t, err := time.Parse(AttendanceDateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func p(s string) *time.Time {
	t := d(s)
	return &t
}

func TestAttendanceToday(t *testing.T) {
	// 2026-11-19 20:30 UTC = 2026-11-20 01:30 Toshkent
	now := time.Date(2026, 11, 19, 20, 30, 0, 0, time.UTC)
	if got := AttendanceToday(now).Format(AttendanceDateLayout); got != "2026-11-20" {
		t.Fatal(got)
	}
	now = time.Date(2026, 11, 20, 18, 59, 0, 0, time.UTC) // 23:59 Toshkent
	if got := AttendanceToday(now).Format(AttendanceDateLayout); got != "2026-11-20" {
		t.Fatal(got)
	}
	got := AttendanceToday(now)
	if got.Location() != time.UTC || got.Hour() != 0 {
		t.Fatal(got)
	}
}

func TestAttendanceFillRange(t *testing.T) {
	today := d("2026-11-20")
	cases := []struct {
		name     string
		last     *time.Time
		from, to string
		days     int
	}{
		{"hech narsa belgilanmagan", nil, "2026-10-22", "2026-11-20", 30},
		{"14-noyabrgacha belgilangan", p("2026-11-14"), "2026-11-15", "2026-11-20", 6},
		{"kecha belgilangan", p("2026-11-19"), "2026-11-20", "2026-11-20", 1},
		{"bugun belgilangan", p("2026-11-20"), "", "", 0},
		{"40 kun oldin belgilangan", p("2026-10-11"), "2026-10-22", "2026-11-20", 30},
		{"roppa-rosa 30 kun oldin", p("2026-10-21"), "2026-10-22", "2026-11-20", 30},
	}
	for _, c := range cases {
		from, to, days := AttendanceFillRange(today, c.last)
		if days != c.days {
			t.Errorf("%s: days = %d, want %d", c.name, days, c.days)
		}
		if c.days > 0 {
			if got := from.Format(AttendanceDateLayout); got != c.from {
				t.Errorf("%s: from = %s, want %s", c.name, got, c.from)
			}
			if got := to.Format(AttendanceDateLayout); got != c.to {
				t.Errorf("%s: to = %s, want %s", c.name, got, c.to)
			}
		}
	}
}

func TestAttendanceRangeValid(t *testing.T) {
	// 62 kunlik: 2026-10-01 .. 2026-12-01; 63 kunlik: 2026-10-01 .. 2026-12-02
	if !AttendanceRangeValid("2026-10-22", "2026-11-20") {
		t.Error("30 kunlik oraliq true bo'lishi kerak")
	}
	if !AttendanceRangeValid("2026-11-20", "2026-11-20") {
		t.Error("bir kunlik oraliq true bo'lishi kerak")
	}
	if !AttendanceRangeValid("2026-10-01", "2026-12-01") {
		t.Error("62 kunlik oraliq true bo'lishi kerak")
	}
	if AttendanceRangeValid("2026-11-20", "2026-11-19") {
		t.Error("from > to false bo'lishi kerak")
	}
	if AttendanceRangeValid("20.11.2026", "2026-11-20") {
		t.Error("noto'g'ri format false bo'lishi kerak")
	}
	if AttendanceRangeValid("", "2026-11-20") {
		t.Error("bo'sh from false bo'lishi kerak")
	}
	if AttendanceRangeValid("2026-10-01", "2026-12-02") {
		t.Error("63 kunlik oraliq false bo'lishi kerak")
	}
}

func TestAttendanceStatusValid(t *testing.T) {
	for _, s := range []int{0, 1, 2} {
		if !AttendanceStatusValid(s) {
			t.Errorf("%d true bo'lishi kerak", s)
		}
	}
	for _, s := range []int{-1, 3} {
		if AttendanceStatusValid(s) {
			t.Errorf("%d false bo'lishi kerak", s)
		}
	}
}
