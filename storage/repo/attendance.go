package repo

import (
	"bw-erp/models"
	"time"
)

type AttendanceI interface {
	// employeeID = 0: korxonaning hamma xodimi, fullname bo'yicha. Aks holda bitta (yoki bo'sh ro'yxat).
	GetEmployees(companyID string, employeeID int) ([]models.AttendanceEmployee, error)
	// from, to — "YYYY-MM-DD", ikkala chet kiradi. employeeID = 0: hamma xodim. Date "YYYY-MM-DD" bo'lib qaytadi.
	GetMarks(companyID string, employeeID int, from string, to string) ([]models.AttendanceMark, error)
	// today yoki undan oldingi eng oxirgi belgilangan sana; yo'q bo'lsa nil.
	LastMarkedDate(companyID string, employeeID int, today string) (*time.Time, error)
	// Upsert. Xodim shu korxonada bo'lmasa 0 qaytadi va hech narsa yozilmaydi.
	SetMark(companyID string, employeeID int, date string, status int) (int64, error)
	// Qatorni o'chiradi (yo'q bo'lsa ham xato emas).
	ClearMark(companyID string, employeeID int, date string) error
	// Shu kunda qatori yo'q hamma xodimga status 2. Qo'shilgan qatorlar soni.
	MarkDay(companyID string, date string) (int64, error)
	// from..to oralig'idagi qatori yo'q kunlarga status 2. Xodim shu korxonada bo'lmasa 0.
	Fill(companyID string, employeeID int, from string, to string) (int64, error)
}
