package postgres

import (
	"bw-erp/models"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

const otherCompany = "11111111-1111-1111-1111-111111111111"

func newAttendanceEmployee(t *testing.T, db *sqlx.DB, name string) int {
	var id int
	err := db.QueryRow(`insert into employees(company_id, fullname, phone) values ($1, $2, '+998000000000') returning id`,
		demo2Company, name).Scan(&id)
	if err != nil {
		t.Fatalf("insert employee: %v", err)
	}
	t.Cleanup(func() { db.Exec(`delete from employees where id = $1`, id) })
	return id
}

func TestAttendanceSetAndClearMark(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAttendanceRepo(db)
	emp := newAttendanceEmployee(t, db, "Davomat test 1")

	n, err := repo.SetMark(demo2Company, emp, "2026-11-14", models.AttendanceFull)
	if err != nil || n != 1 {
		t.Fatalf("set: n=%d err=%v", n, err)
	}
	n, err = repo.SetMark(demo2Company, emp, "2026-11-14", models.AttendanceAbsent)
	if err != nil || n != 1 {
		t.Fatalf("set again: n=%d err=%v", n, err)
	}
	marks, err := repo.GetMarks(demo2Company, emp, "2026-11-14", "2026-11-14")
	if err != nil {
		t.Fatalf("marks: %v", err)
	}
	if len(marks) != 1 || marks[0].Status != 0 || marks[0].Date != "2026-11-14" || marks[0].EmployeeID != emp {
		t.Fatalf("marks: %+v", marks)
	}

	if err := repo.ClearMark(demo2Company, emp, "2026-11-14"); err != nil {
		t.Fatalf("clear: %v", err)
	}
	marks, err = repo.GetMarks(demo2Company, emp, "2026-11-14", "2026-11-14")
	if err != nil || len(marks) != 0 || marks == nil {
		t.Fatalf("after clear: %+v %v", marks, err)
	}
	if err := repo.ClearMark(demo2Company, emp, "2026-11-14"); err != nil {
		t.Fatalf("clear again: %v", err)
	}
}

func TestAttendanceSetMarkOtherCompany(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAttendanceRepo(db)
	emp := newAttendanceEmployee(t, db, "Davomat test 2")

	n, err := repo.SetMark(otherCompany, emp, "2026-11-14", models.AttendanceFull)
	if err != nil || n != 0 {
		t.Fatalf("set: n=%d err=%v", n, err)
	}
	var count int
	db.Get(&count, `select count(*) from employee_attendance where employee_id = $1`, emp)
	if count != 0 {
		t.Fatalf("rows: %d", count)
	}
}

func TestAttendanceLastMarkedDate(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAttendanceRepo(db)
	emp := newAttendanceEmployee(t, db, "Davomat test 3")

	last, err := repo.LastMarkedDate(demo2Company, emp, "2026-11-20")
	if err != nil || last != nil {
		t.Fatalf("empty: %v %v", last, err)
	}
	repo.SetMark(demo2Company, emp, "2026-11-10", models.AttendanceFull)
	repo.SetMark(demo2Company, emp, "2026-11-25", models.AttendanceFull)
	last, err = repo.LastMarkedDate(demo2Company, emp, "2026-11-20")
	if err != nil || last == nil {
		t.Fatalf("last: %v %v", last, err)
	}
	want := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	if !last.Equal(want) {
		t.Fatalf("last = %v, want %v", last, want)
	}
}

func TestAttendanceFill(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAttendanceRepo(db)
	emp := newAttendanceEmployee(t, db, "Davomat test 4")

	repo.SetMark(demo2Company, emp, "2026-11-16", models.AttendanceAbsent)
	n, err := repo.Fill(demo2Company, emp, "2026-11-15", "2026-11-20")
	if err != nil || n != 5 {
		t.Fatalf("fill: n=%d err=%v", n, err)
	}
	marks, _ := repo.GetMarks(demo2Company, emp, "2026-11-15", "2026-11-20")
	if len(marks) != 6 {
		t.Fatalf("marks: %+v", marks)
	}
	for _, m := range marks {
		want := 2
		if m.Date == "2026-11-16" {
			want = 0
		}
		if m.Status != want {
			t.Fatalf("mark %s status %d, want %d", m.Date, m.Status, want)
		}
	}
	n, err = repo.Fill(demo2Company, emp, "2026-11-15", "2026-11-20")
	if err != nil || n != 0 {
		t.Fatalf("fill again: n=%d err=%v", n, err)
	}

	emp2 := newAttendanceEmployee(t, db, "Davomat test 4b")
	n, err = repo.Fill(otherCompany, emp2, "2026-11-15", "2026-11-20")
	if err != nil || n != 0 {
		t.Fatalf("fill other company: n=%d err=%v", n, err)
	}
	var count int
	db.Get(&count, `select count(*) from employee_attendance where employee_id = $1`, emp2)
	if count != 0 {
		t.Fatalf("rows: %d", count)
	}
}

func TestAttendanceMarkDay(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAttendanceRepo(db)
	a := newAttendanceEmployee(t, db, "Davomat test 5a")
	b := newAttendanceEmployee(t, db, "Davomat test 5b")
	t.Cleanup(func() {
		db.Exec(`delete from employee_attendance where company_id = $1 and date = '2026-11-20'`, demo2Company)
	})

	repo.SetMark(demo2Company, a, "2026-11-20", models.AttendanceHalf)
	n, err := repo.MarkDay(demo2Company, "2026-11-20")
	if err != nil || n < 1 {
		t.Fatalf("markday: n=%d err=%v", n, err)
	}
	marks, _ := repo.GetMarks(demo2Company, 0, "2026-11-20", "2026-11-20")
	got := map[int]int{}
	for _, m := range marks {
		got[m.EmployeeID] = m.Status
	}
	if got[a] != 1 || got[b] != 2 {
		t.Fatalf("a=%d b=%d", got[a], got[b])
	}
}

func TestAttendanceGetEmployeesAndMarksRange(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAttendanceRepo(db)
	emp := newAttendanceEmployee(t, db, "Davomat test 6")

	list, err := repo.GetEmployees(demo2Company, emp)
	if err != nil || len(list) != 1 || list[0].ID != emp || list[0].Fullname != "Davomat test 6" {
		t.Fatalf("employees: %+v %v", list, err)
	}
	list, err = repo.GetEmployees(otherCompany, emp)
	if err != nil || len(list) != 0 || list == nil {
		t.Fatalf("other company employees: %+v %v", list, err)
	}

	for _, d := range []string{"2026-11-21", "2026-11-09", "2026-11-20", "2026-11-10"} {
		repo.SetMark(demo2Company, emp, d, models.AttendanceFull)
	}
	marks, err := repo.GetMarks(demo2Company, emp, "2026-11-10", "2026-11-20")
	if err != nil || len(marks) != 2 || marks[0].Date != "2026-11-10" || marks[1].Date != "2026-11-20" {
		t.Fatalf("marks: %+v %v", marks, err)
	}
}
