package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/storage"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

type fakeAttendanceRepo struct {
	repo.AttendanceI
	lastMarked   *time.Time
	markDayCount int64
	employees    []models.AttendanceEmployee
	marks        []models.AttendanceMark

	getMarksCalled bool
	setCalls       []string
	setStatus      int
	clearCalled    bool
	markDayDate    string
	fillCalled     bool
	fillFrom       string
	fillTo         string
}

func (f *fakeAttendanceRepo) GetEmployees(companyID string, employeeID int) ([]models.AttendanceEmployee, error) {
	if employeeID != 0 && (companyID != companyA || employeeID != 7) {
		return nil, nil
	}
	return f.employees, nil
}

func (f *fakeAttendanceRepo) GetMarks(companyID string, employeeID int, from string, to string) ([]models.AttendanceMark, error) {
	f.getMarksCalled = true
	return f.marks, nil
}

func (f *fakeAttendanceRepo) LastMarkedDate(companyID string, employeeID int, today string) (*time.Time, error) {
	return f.lastMarked, nil
}

func (f *fakeAttendanceRepo) SetMark(companyID string, employeeID int, date string, status int) (int64, error) {
	if companyID != companyA || employeeID != 7 {
		return 0, nil
	}
	f.setCalls = append(f.setCalls, companyID+"|"+date)
	f.setStatus = status
	return 1, nil
}

func (f *fakeAttendanceRepo) ClearMark(companyID string, employeeID int, date string) error {
	f.clearCalled = true
	return nil
}

func (f *fakeAttendanceRepo) MarkDay(companyID string, date string) (int64, error) {
	f.markDayDate = date
	return f.markDayCount, nil
}

func (f *fakeAttendanceRepo) Fill(companyID string, employeeID int, from string, to string) (int64, error) {
	f.fillCalled = true
	f.fillFrom = from
	f.fillTo = to
	return 5, nil
}

type attendanceStorage struct {
	storage.StorageI
	att *fakeAttendanceRepo
}

func (s attendanceStorage) Attendance() repo.AttendanceI { return s.att }

func newAttendanceRouter(user gin.HandlerFunc, att *fakeAttendanceRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handlers.Handler{Stg: attendanceStorage{att: att}}
	r := gin.New()
	usersRoutes(r.Group("/users"), &h, user)
	return r
}

const attBase = "/users/employees/attendance"

func attDate(d time.Time) string { return d.Format(models.AttendanceDateLayout) }

func TestAttendanceWritesRequirePermission(t *testing.T) {
	att := &fakeAttendanceRepo{}
	r := newAttendanceRouter(userWith("show_employee"), att)
	bodies := map[string]string{
		attBase + "/mark":     `{"company_id":"` + companyA + `","employee_id":7,"date":"2026-11-14","status":2}`,
		attBase + "/mark-day": `{"company_id":"` + companyA + `","date":"2026-11-20"}`,
		attBase + "/fill":     `{"company_id":"` + companyA + `","employee_id":7}`,
	}
	for path, body := range bodies {
		w := do(r, "POST", path, body)
		if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" {
			t.Fatalf("%s: code=%d body=%s", path, w.Code, w.Body.String())
		}
	}
	if len(att.setCalls) != 0 || att.clearCalled || att.markDayDate != "" || att.fillCalled {
		t.Fatalf("repo written: %+v", att)
	}
}

func TestAttendanceGetAllowedWithShowEmployee(t *testing.T) {
	r := newAttendanceRouter(userWith("show_employee"), &fakeAttendanceRepo{employees: []models.AttendanceEmployee{{ID: 7, Fullname: "Ali"}}})
	w := do(r, "GET", attBase+"?company_id="+companyA+"&from=2026-10-22&to=2026-11-20&employee_id=7", "")
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAttendanceGetDeniedWithoutAnySlug(t *testing.T) {
	att := &fakeAttendanceRepo{}
	r := newAttendanceRouter(userWith("get_orders_list"), att)
	w := do(r, "GET", attBase+"?company_id="+companyA+"&from=2026-11-14&to=2026-11-20", "")
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" || att.getMarksCalled {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAttendanceGetBadRange(t *testing.T) {
	r := newAttendanceRouter(userWith("attendance_employee"), &fakeAttendanceRepo{})
	for _, q := range []string{"from=2026-11-20&to=2026-11-19", "from=2026-01-01&to=2026-11-20"} {
		w := do(r, "GET", attBase+"?company_id="+companyA+"&"+q, "")
		if w.Code != http.StatusBadRequest || responseStatus(w.Body.Bytes()) != "INVALID_ARGUMENT" {
			t.Fatalf("%s: code=%d body=%s", q, w.Code, w.Body.String())
		}
	}
}

func TestAttendanceGetOtherCompanyEmployeeIs404(t *testing.T) {
	att := &fakeAttendanceRepo{}
	r := newAttendanceRouter(userWith("attendance_employee"), att)
	w := do(r, "GET", attBase+"?company_id="+companyA+"&from=2026-11-14&to=2026-11-20&employee_id=99", "")
	if w.Code != http.StatusNotFound || responseStatus(w.Body.Bytes()) != "NOT_FOUND" || att.getMarksCalled {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAttendanceGetFill(t *testing.T) {
	today := models.AttendanceToday(time.Now())
	type resp struct {
		Data struct {
			Today     string                 `json:"today"`
			Employees json.RawMessage        `json:"employees"`
			Marks     json.RawMessage        `json:"marks"`
			Fill      *models.AttendanceFill `json:"fill"`
		} `json:"data"`
	}
	url := attBase + "?company_id=" + companyA + "&from=2026-11-14&to=2026-11-20"

	threeDaysAgo := today.AddDate(0, 0, -3)
	att := &fakeAttendanceRepo{lastMarked: &threeDaysAgo, employees: []models.AttendanceEmployee{{ID: 7, Fullname: "Ali"}}}
	r := newAttendanceRouter(userWith("attendance_employee"), att)

	w := do(r, "GET", url+"&employee_id=7", "")
	var got resp
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if w.Code != http.StatusOK || got.Data.Fill == nil || got.Data.Fill.Days != 3 || got.Data.Fill.To != attDate(today) || got.Data.Today != attDate(today) {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}

	att.lastMarked = &today
	w = do(r, "GET", url+"&employee_id=7", "")
	got = resp{}
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if w.Code != http.StatusOK || got.Data.Fill != nil {
		t.Fatalf("today marked: body=%s", w.Body.String())
	}

	att.lastMarked = &threeDaysAgo
	w = do(r, "GET", url, "")
	got = resp{}
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if w.Code != http.StatusOK || got.Data.Fill != nil {
		t.Fatalf("no employee_id: body=%s", w.Body.String())
	}

	empty := newAttendanceRouter(userWith("attendance_employee"), &fakeAttendanceRepo{})
	w = do(empty, "GET", url, "")
	got = resp{}
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if string(got.Data.Employees) != "[]" || string(got.Data.Marks) != "[]" {
		t.Fatalf("empty lists: body=%s", w.Body.String())
	}
}

func TestAttendanceMark(t *testing.T) {
	att := &fakeAttendanceRepo{employees: []models.AttendanceEmployee{{ID: 7, Fullname: "Ali"}}}
	r := newAttendanceRouter(userWith("attendance_employee"), att)
	post := func(body string) int {
		return do(r, "POST", attBase+"/mark", `{"company_id":"`+companyA+`",`+body+`}`).Code
	}

	if code := post(`"employee_id":7,"date":"2026-11-14","status":2`); code != http.StatusOK ||
		len(att.setCalls) != 1 || att.setCalls[0] != companyA+"|2026-11-14" || att.setStatus != 2 {
		t.Fatalf("set: code=%d att=%+v", code, att)
	}

	att.setCalls = nil
	if code := post(`"employee_id":7,"date":"2026-11-14","status":null`); code != http.StatusOK || !att.clearCalled || len(att.setCalls) != 0 {
		t.Fatalf("clear: code=%d att=%+v", code, att)
	}

	att.clearCalled = false
	if code := post(`"employee_id":7,"date":"2026-11-14","status":3`); code != http.StatusBadRequest {
		t.Fatalf("status 3: code=%d", code)
	}
	if code := post(`"employee_id":7,"date":"14.11.2026","status":2`); code != http.StatusBadRequest {
		t.Fatalf("bad date: code=%d", code)
	}

	if code := post(`"employee_id":99,"date":"2026-11-14","status":2`); code != http.StatusNotFound {
		t.Fatalf("foreign set: code=%d", code)
	}
	if code := post(`"employee_id":99,"date":"2026-11-14","status":null`); code != http.StatusNotFound || att.clearCalled {
		t.Fatalf("foreign clear: code=%d clear=%v", code, att.clearCalled)
	}
}

func TestAttendanceMarkDay(t *testing.T) {
	att := &fakeAttendanceRepo{markDayCount: 4}
	r := newAttendanceRouter(userWith("attendance_employee"), att)
	w := do(r, "POST", attBase+"/mark-day", `{"company_id":"`+companyA+`","date":"2026-11-20"}`)
	var got struct {
		Data struct {
			Marked int64 `json:"marked"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if w.Code != http.StatusOK || got.Data.Marked != 4 || att.markDayDate != "2026-11-20" {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	w = do(r, "POST", attBase+"/mark-day", `{"company_id":"`+companyA+`","date":"20.11.2026"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad date: code=%d", w.Code)
	}
}

func TestAttendanceFillUsesServerRange(t *testing.T) {
	today := models.AttendanceToday(time.Now())
	att := &fakeAttendanceRepo{employees: []models.AttendanceEmployee{{ID: 7, Fullname: "Ali"}}}
	r := newAttendanceRouter(userWith("attendance_employee"), att)

	w := do(r, "POST", attBase+"/fill", `{"company_id":"`+companyA+`","employee_id":7,"from":"2020-01-01"}`)
	if w.Code != http.StatusOK || att.fillFrom != attDate(today.AddDate(0, 0, -29)) || att.fillTo != attDate(today) {
		t.Fatalf("code=%d from=%s to=%s", w.Code, att.fillFrom, att.fillTo)
	}

	att.fillCalled = false
	w = do(r, "POST", attBase+"/fill", `{"company_id":"`+companyA+`","employee_id":99}`)
	if w.Code != http.StatusNotFound || att.fillCalled {
		t.Fatalf("foreign: code=%d fill=%v", w.Code, att.fillCalled)
	}

	att.lastMarked = &today
	w = do(r, "POST", attBase+"/fill", `{"company_id":"`+companyA+`","employee_id":7}`)
	if w.Code != http.StatusOK || att.fillCalled {
		t.Fatalf("nothing to fill: code=%d fill=%v", w.Code, att.fillCalled)
	}
}
