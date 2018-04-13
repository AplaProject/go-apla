package model

const tableNameMetrics = "1_metrics"

type Metric struct {
	ID     int64  `gorm:"primary_key;not null"`
	Time   int64  `gorm:"not null"`
	Metric string `gorm:"not null"`
	Key    string `gorm:"not null"`
	Value  int64  `gorm:"not null"`
}

func (Metric) TableName() string {
	return tableNameMetrics
}

type EcosystemTx struct {
	UnixTime  int64
	Ecosystem string
	Count     int64
}

// GetEcosystemTxPerDay returns the count of transactions per day for ecosystems,
// processes data for two days
func GetEcosystemTxPerDay() ([]*EcosystemTx, error) {
	sql := `SELECT
		EXTRACT(EPOCH FROM to_timestamp(bc.time)::date)::int "unix_time",
		SUBSTRING(rtx.table_name FROM '^\d+') "ecosystem",
		COUNT(*)
	FROM rollback_tx rtx
		INNER JOIN block_chain bc ON bc.id = rtx.block_id
	WHERE to_timestamp(bc.time)::date >= current_date-1
	GROUP BY unix_time, ecosystem`

	var ecosystemTx []*EcosystemTx
	err := DBConn.Raw(sql).Scan(&ecosystemTx).Error
	if err != nil {
		return nil, err
	}

	return ecosystemTx, err
}

// GetMetricValues returns aggregated metric values in the time interval
func GetMetricValues(metric, timeInterval, aggregateFunc string) ([]interface{}, error) {
	rows, err := DBConn.Table(tableNameMetrics).Select("key,"+aggregateFunc+"(value)").
		Where("metric = ? AND time >= EXTRACT(EPOCH FROM NOW() - CAST(? AS INTERVAL))", metric, timeInterval).
		Group("key").Rows()
	if err != nil {
		return nil, err
	}

	var (
		result = []interface{}{}
		key    string
		value  string
	)
	for rows.Next() {
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}

		result = append(result, map[string]string{
			"key":   key,
			"value": value,
		})
	}

	return result, nil
}
