package models

import "time"

const (
	AttendanceAbsent = 0 // kelmadi
	AttendanceHalf   = 1 // yarim kun
	AttendanceFull   = 2 // keldi

	AttendanceFillMaxDays  = 30
	AttendanceMaxRangeDays = 62
	AttendanceDateLayout   = "2006-01-02"
)

// AttendanceToday: now ni Toshkent vaqtiga o'tkazadi va shu sananing UTC 00:00 ini qaytaradi.
func AttendanceToday(now time.Time) time.Time {
	local := now.In(time.FixedZone("Asia/Tashkent", 5*3600))
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

// AttendanceFillRange: lastMarked — bugun yoki undan oldingi oxirgi belgilangan sana (nil = yo'q).
// days = 0 bo'lsa to'ldiradigan kun yo'q.
func AttendanceFillRange(today time.Time, lastMarked *time.Time) (from time.Time, to time.Time, days int) {
	to = today
	earliest := today.AddDate(0, 0, -(AttendanceFillMaxDays - 1))
	from = earliest
	if lastMarked != nil {
		from = lastMarked.AddDate(0, 0, 1)
		if from.Before(earliest) {
			from = earliest
		}
	}
	if from.After(to) {
		return from, to, 0
	}
	days = int(to.Sub(from).Hours()/24) + 1
	return from, to, days
}

// AttendanceRangeValid: ikkala sana AttendanceDateLayout da, from <= to, oraliq AttendanceMaxRangeDays dan oshmaydi.
func AttendanceRangeValid(from string, to string) bool {
	f, err := time.Parse(AttendanceDateLayout, from)
	if err != nil {
		return false
	}
	t, err := time.Parse(AttendanceDateLayout, to)
	if err != nil {
		return false
	}
	if f.After(t) {
		return false
	}
	days := int(t.Sub(f).Hours()/24) + 1
	return days <= AttendanceMaxRangeDays
}

// AttendanceStatusValid: 0, 1 yoki 2.
func AttendanceStatusValid(status int) bool {
	return status == AttendanceAbsent || status == AttendanceHalf || status == AttendanceFull
}

type AttendanceListRequest struct {
	CompanyID  string `form:"company_id" binding:"required"`
	From       string `form:"from" binding:"required"`
	To         string `form:"to" binding:"required"`
	EmployeeID int    `form:"employee_id"`
}

type AttendanceEmployee struct {
	ID       int    `json:"id" db:"id"`
	Fullname string `json:"fullname" db:"fullname"`
}

type AttendanceMark struct {
	EmployeeID int    `json:"employee_id" db:"employee_id"`
	Date       string `json:"date" db:"date"`
	Status     int    `json:"status" db:"status"`
}

type AttendanceFill struct {
	From string `json:"from"`
	To   string `json:"to"`
	Days int    `json:"days"`
}

// Employees va Marks bo'sh bo'lsa JSON da [] bo'lishi uchun nil bermang.
type AttendanceListResponse struct {
	Today     string               `json:"today"`
	Employees []AttendanceEmployee `json:"employees"`
	Marks     []AttendanceMark     `json:"marks"`
	Fill      *AttendanceFill      `json:"fill"`
}

type AttendanceMarkRequest struct {
	CompanyID  string `json:"company_id" binding:"required"`
	EmployeeID int    `json:"employee_id" binding:"required"`
	Date       string `json:"date" binding:"required"`
	Status     *int   `json:"status"` // nil = bo'sh (qator o'chiriladi)
}

type AttendanceMarkDayRequest struct {
	CompanyID string `json:"company_id" binding:"required"`
	Date      string `json:"date" binding:"required"`
}

type AttendanceFillRequest struct {
	CompanyID  string `json:"company_id" binding:"required"`
	EmployeeID int    `json:"employee_id" binding:"required"`
}
