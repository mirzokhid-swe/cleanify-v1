package postgres

import (
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
