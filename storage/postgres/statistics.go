package postgres

import (
	"database/sql"
	"sort"
	"time"

	"bw-erp/helper"
	"bw-erp/models"
	"bw-erp/storage/repo"

	"github.com/jmoiron/sqlx"
)

type statisticsRepo struct {
	db *sqlx.DB
}

func NewStatisticsRepo(db *sqlx.DB) repo.StatisticsI {
	return &statisticsRepo{db: db}
}

func (stg *statisticsRepo) GetWorkVolume(companyID string) ([]models.WorkVolume, error) {
	var arr []interface{}
	var workVolumes []models.WorkVolume
	params := make(map[string]interface{})
	query := `SELECT
	round(sum((width::numeric * height::numeric)), 2) as meter_square,
		washed_at::date,
		type,
		is_countable
		FROM order_items oi inner join orders o on oi.order_id = o.id WHERE washed_at::date >= (NOW()::date - INTERVAL '15 days')`

	filter := " "
	order := " ORDER BY washed_at"
	arrangement := " DESC"
	group := " group by washed_at::date, type, is_countable"

	params["company_id"] = companyID
	filter += " AND (o.company_id = :company_id)"

	q := query + filter + group + order + arrangement

	q, arr = helper.ReplaceQueryParams(q, params)
	rows, err := stg.db.Query(q, arr...)
	if err != nil {
		return workVolumes, err
	}
	defer rows.Close()

	for rows.Next() {
		var workVolume models.WorkVolume
		err = rows.Scan(
			&workVolume.MeterSquare,
			&workVolume.WashedAt,
			&workVolume.Type,
			&workVolume.IsCountable)
		if err != nil {
			return nil, err
		}
		workVolumes = append(workVolumes, workVolume)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workVolumes, nil
}

// GetStatusMetrics: har kunda shu holatga o'tgan buyurtmalar soni va buyumlari,
// kun -> xodim -> tur kesimida. Bir buyurtma bir kunda bir marta (birinchi yozuv).
func (stg *statisticsRepo) GetStatusMetrics(companyID string, status int, days int) ([]models.StatusMetricsDay, error) {
	params := map[string]interface{}{
		"company_id": companyID,
		"status":     status,
		"days":       days,
	}
	query := `WITH moved AS (
  SELECT DISTINCT ON (h.historyable_id, h.created_at::date)
         h.historyable_id AS order_id, h.created_at::date AS day, h.user_id
  FROM status_change_histories h
  JOIN orders o ON o.id = h.historyable_id
  WHERE h.historyable_type = 'orders' AND h.status = :status
    AND o.company_id = :company_id
    AND h.created_at::date >= current_date - (:days::int - 1)
  ORDER BY h.historyable_id, h.created_at::date, h.created_at
)
SELECT m.day, m.user_id, COALESCE(u.fullname, ''), m.order_id, COALESCE(oi.type, ''), COALESCE(oi.unit, ''),
       COALESCE(SUM(oi.width * oi.height), 0) AS qty
FROM moved m
LEFT JOIN users u ON u.id = m.user_id
LEFT JOIN order_items oi ON oi.order_id = m.order_id
GROUP BY m.day, m.user_id, u.fullname, m.order_id, oi.type, oi.unit
ORDER BY m.day DESC`

	q, arr := helper.ReplaceQueryParams(query, params)
	rows, err := stg.db.Query(q, arr...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type userAgg struct {
		fullname string
		orders   map[int]bool
		items    map[string]*models.StatusMetricsItem
	}
	type dayAgg struct {
		orders map[int]bool
		items  map[string]*models.StatusMetricsItem
		users  map[int64]*userAgg
	}
	dayOrder := []string{}
	daysAgg := map[string]*dayAgg{}

	for rows.Next() {
		var day time.Time
		var userID sql.NullInt64
		var fullname string
		var orderID int
		var typ, unit string
		var qty float64
		if err := rows.Scan(&day, &userID, &fullname, &orderID, &typ, &unit, &qty); err != nil {
			return nil, err
		}
		dayKey := day.Format("2006-01-02")
		d, ok := daysAgg[dayKey]
		if !ok {
			d = &dayAgg{orders: map[int]bool{}, items: map[string]*models.StatusMetricsItem{}, users: map[int64]*userAgg{}}
			daysAgg[dayKey] = d
			dayOrder = append(dayOrder, dayKey)
		}
		d.orders[orderID] = true
		u, ok := d.users[userID.Int64]
		if !ok {
			u = &userAgg{fullname: fullname, orders: map[int]bool{}, items: map[string]*models.StatusMetricsItem{}}
			d.users[userID.Int64] = u
		}
		u.orders[orderID] = true
		if typ != "" {
			if unit == "" {
				unit = "area"
			}
			key := typ + "|" + unit
			if it, ok := d.items[key]; ok {
				it.Qty += qty
			} else {
				d.items[key] = &models.StatusMetricsItem{Type: typ, Unit: unit, Qty: qty}
			}
			if it, ok := u.items[key]; ok {
				it.Qty += qty
			} else {
				u.items[key] = &models.StatusMetricsItem{Type: typ, Unit: unit, Qty: qty}
			}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	sortItems := func(m map[string]*models.StatusMetricsItem) []models.StatusMetricsItem {
		out := make([]models.StatusMetricsItem, 0, len(m))
		for _, it := range m {
			out = append(out, *it)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Qty > out[j].Qty })
		return out
	}

	result := make([]models.StatusMetricsDay, 0, len(dayOrder))
	for _, dayKey := range dayOrder {
		d := daysAgg[dayKey]
		users := make([]models.StatusMetricsUser, 0, len(d.users))
		for id, u := range d.users {
			users = append(users, models.StatusMetricsUser{
				UserID:   id,
				Fullname: u.fullname,
				Count:    len(u.orders),
				Items:    sortItems(u.items),
			})
		}
		sort.Slice(users, func(i, j int) bool {
			if users[i].Count != users[j].Count {
				return users[i].Count > users[j].Count
			}
			return users[i].Fullname < users[j].Fullname
		})
		result = append(result, models.StatusMetricsDay{
			Date:  dayKey,
			Count: len(d.orders),
			Items: sortItems(d.items),
			Users: users,
		})
	}
	return result, nil
}

func (stg *statisticsRepo) GetServicePaymentStatistics(entity models.GetServicePaymentStatisticsRequest) ([]models.ServicePaymentStatistics, error) {
	var arr []interface{}
	var servicePaymentStatistics []models.ServicePaymentStatistics
	params := make(map[string]interface{})

	query := `select u.id, u.fullname, sum(t.amount) as amount from transactions t
				inner join users u on t.receiver_id = u.id::text and t.payer_type = 'orders'`

	params["company_id"] = entity.CompanyID

	filter := " where (t.company_id = :company_id) and t.deleted_at is null"

	if entity.DateFrom != "" {
		params["date_from"] = entity.DateFrom
		filter += " and (t.created_at::date >= :date_from::date) "
	}

	if entity.DateTo != "" {
		params["date_to"] = entity.DateTo
		filter += " and (t.created_at::date <= :date_to::date) "
	}

	if entity.DateFrom == "" && entity.DateTo == "" {
		filter += " and (t.created_at::date = now()::date) "
	}

	group := " group by u.id, u.fullname"

	q := query + filter + group

	q, arr = helper.ReplaceQueryParams(q, params)
	rows, err := stg.db.Query(q, arr...)

	if err != nil {
		return servicePaymentStatistics, err
	}
	defer rows.Close()

	for rows.Next() {
		var servicePaymentStatistic models.ServicePaymentStatistics
		err = rows.Scan(
			&servicePaymentStatistic.UserID,
			&servicePaymentStatistic.Fullname,
			&servicePaymentStatistic.Amount)
		if err != nil {
			return nil, err
		}
		servicePaymentStatistics = append(servicePaymentStatistics, servicePaymentStatistic)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return servicePaymentStatistics, nil
}
