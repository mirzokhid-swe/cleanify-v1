package postgres

import (
	"bw-erp/models"
	"bw-erp/storage/repo"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type attendanceRepo struct {
	db *sqlx.DB
}

func NewAttendanceRepo(db *sqlx.DB) repo.AttendanceI {
	return &attendanceRepo{db: db}
}

func (stg attendanceRepo) GetEmployees(companyID string, employeeID int) ([]models.AttendanceEmployee, error) {
	list := []models.AttendanceEmployee{}
	err := stg.db.Select(&list, `SELECT id, fullname
		FROM employees
		WHERE company_id = $1::uuid
		  AND ($2::int = 0 OR id = $2::int)
		ORDER BY fullname, id`, companyID, employeeID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (stg attendanceRepo) GetMarks(companyID string, employeeID int, from string, to string) ([]models.AttendanceMark, error) {
	list := []models.AttendanceMark{}
	err := stg.db.Select(&list, `SELECT employee_id, to_char(date, 'YYYY-MM-DD') AS date, status
		FROM employee_attendance
		WHERE company_id = $1::uuid
		  AND ($2::int = 0 OR employee_id = $2::int)
		  AND date BETWEEN $3::date AND $4::date
		ORDER BY date, employee_id`, companyID, employeeID, from, to)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (stg attendanceRepo) LastMarkedDate(companyID string, employeeID int, today string) (*time.Time, error) {
	var s sql.NullString
	err := stg.db.Get(&s, `SELECT to_char(max(date), 'YYYY-MM-DD')
		FROM employee_attendance
		WHERE company_id = $1::uuid
		  AND ($2::int = 0 OR employee_id = $2::int)
		  AND date <= $3::date`, companyID, employeeID, today)
	if err != nil {
		return nil, err
	}
	if !s.Valid {
		return nil, nil
	}
	t, err := time.Parse(models.AttendanceDateLayout, s.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (stg attendanceRepo) SetMark(companyID string, employeeID int, date string, status int) (int64, error) {
	res, err := stg.db.Exec(`INSERT INTO employee_attendance (company_id, employee_id, date, status)
		SELECT $1::uuid, e.id, $3::date, $4::smallint
		FROM employees e
		WHERE e.id = $2::int AND e.company_id = $1::uuid
		ON CONFLICT (employee_id, date) DO UPDATE
		SET status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP`, companyID, employeeID, date, status)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (stg attendanceRepo) ClearMark(companyID string, employeeID int, date string) error {
	_, err := stg.db.Exec(`DELETE FROM employee_attendance
		WHERE company_id = $1::uuid AND employee_id = $2::int AND date = $3::date`, companyID, employeeID, date)
	return err
}

func (stg attendanceRepo) MarkDay(companyID string, date string) (int64, error) {
	res, err := stg.db.Exec(`INSERT INTO employee_attendance (company_id, employee_id, date, status)
		SELECT company_id, id, $2::date, 2
		FROM employees
		WHERE company_id = $1::uuid
		ON CONFLICT (employee_id, date) DO NOTHING`, companyID, date)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (stg attendanceRepo) Fill(companyID string, employeeID int, from string, to string) (int64, error) {
	res, err := stg.db.Exec(`INSERT INTO employee_attendance (company_id, employee_id, date, status)
		SELECT e.company_id, e.id, g::date, 2
		FROM employees e, generate_series($3::date, $4::date, interval '1 day') g
		WHERE e.id = $2::int AND e.company_id = $1::uuid
		ON CONFLICT (employee_id, date) DO NOTHING`, companyID, employeeID, from, to)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
