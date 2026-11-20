package handlers

import (
	"bw-erp/api/http"
	"bw-erp/api/middleware"
	"bw-erp/models"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAttendance(c *gin.Context) {
	var req models.AttendanceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if !middleware.UserCan(c, req.CompanyID, "attendance_employee") && !middleware.UserCan(c, req.CompanyID, "show_employee") {
		h.handleResponse(c, http.PermissionDenied, nil)
		return
	}
	if !models.AttendanceRangeValid(req.From, req.To) {
		h.handleResponse(c, http.InvalidArgument, "invalid date range")
		return
	}

	employees, err := h.Stg.Attendance().GetEmployees(req.CompanyID, req.EmployeeID)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	if req.EmployeeID != 0 && len(employees) == 0 {
		h.handleResponse(c, http.NOT_FOUND, "employee not found")
		return
	}
	marks, err := h.Stg.Attendance().GetMarks(req.CompanyID, req.EmployeeID, req.From, req.To)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	if employees == nil {
		employees = []models.AttendanceEmployee{}
	}
	if marks == nil {
		marks = []models.AttendanceMark{}
	}

	today := models.AttendanceToday(time.Now())
	resp := models.AttendanceListResponse{
		Today:     today.Format(models.AttendanceDateLayout),
		Employees: employees,
		Marks:     marks,
	}
	if req.EmployeeID != 0 {
		last, err := h.Stg.Attendance().LastMarkedDate(req.CompanyID, req.EmployeeID, resp.Today)
		if err != nil {
			h.handleResponse(c, http.InternalServerError, err.Error())
			return
		}
		from, to, days := models.AttendanceFillRange(today, last)
		if days > 0 {
			resp.Fill = &models.AttendanceFill{
				From: from.Format(models.AttendanceDateLayout),
				To:   to.Format(models.AttendanceDateLayout),
				Days: days,
			}
		}
	}
	h.handleResponse(c, http.OK, resp)
}

func (h *Handler) MarkAttendance(c *gin.Context) {
	var req models.AttendanceMarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if _, err := time.Parse(models.AttendanceDateLayout, req.Date); err != nil {
		h.handleResponse(c, http.InvalidArgument, "invalid date")
		return
	}

	if req.Status == nil {
		employees, err := h.Stg.Attendance().GetEmployees(req.CompanyID, req.EmployeeID)
		if err != nil {
			h.handleResponse(c, http.InternalServerError, err.Error())
			return
		}
		if len(employees) == 0 {
			h.handleResponse(c, http.NOT_FOUND, "employee not found")
			return
		}
		if err := h.Stg.Attendance().ClearMark(req.CompanyID, req.EmployeeID, req.Date); err != nil {
			h.handleResponse(c, http.InternalServerError, err.Error())
			return
		}
		h.handleResponse(c, http.OK, "OK!")
		return
	}

	if !models.AttendanceStatusValid(*req.Status) {
		h.handleResponse(c, http.InvalidArgument, "invalid status")
		return
	}
	rows, err := h.Stg.Attendance().SetMark(req.CompanyID, req.EmployeeID, req.Date, *req.Status)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	if rows == 0 {
		h.handleResponse(c, http.NOT_FOUND, "employee not found")
		return
	}
	h.handleResponse(c, http.OK, "OK!")
}

func (h *Handler) MarkAttendanceDay(c *gin.Context) {
	var req models.AttendanceMarkDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if _, err := time.Parse(models.AttendanceDateLayout, req.Date); err != nil {
		h.handleResponse(c, http.InvalidArgument, "invalid date")
		return
	}

	marked, err := h.Stg.Attendance().MarkDay(req.CompanyID, req.Date)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	h.handleResponse(c, http.OK, gin.H{"marked": marked})
}

func (h *Handler) FillAttendance(c *gin.Context) {
	var req models.AttendanceFillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	employees, err := h.Stg.Attendance().GetEmployees(req.CompanyID, req.EmployeeID)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	if len(employees) == 0 {
		h.handleResponse(c, http.NOT_FOUND, "employee not found")
		return
	}

	today := models.AttendanceToday(time.Now())
	last, err := h.Stg.Attendance().LastMarkedDate(req.CompanyID, req.EmployeeID, today.Format(models.AttendanceDateLayout))
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	from, to, days := models.AttendanceFillRange(today, last)
	if days == 0 {
		h.handleResponse(c, http.OK, gin.H{"marked": 0})
		return
	}

	marked, err := h.Stg.Attendance().Fill(req.CompanyID, req.EmployeeID, from.Format(models.AttendanceDateLayout), to.Format(models.AttendanceDateLayout))
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	h.handleResponse(c, http.OK, gin.H{"marked": marked})
}
