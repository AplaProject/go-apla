package model

import "time"

const tableNameMetrics = "1_metrics"

// Metric represents record of system_metrics table
type Metric struct {
	ID     int64  `gorm:"primary_key;not null"`
	Time   int64  `gorm:"not null"`
	Metric string `gorm:"not null"`
	Key    string `gorm:"not null"`
	Value  int64  `gorm:"not null"`
}

// TableName returns name of table
func (Metric) TableName() string {
	return tableNameMetrics
}

// EcosystemTx represents value of metric
type EcosystemTx struct {
	UnixTime  int64
	Ecosystem string
	Count     int64
}

// GetEcosystemTxPerDay returns the count of transactions per day for ecosystems,
// processes data for two days
func GetEcosystemTxPerDay(timeBlock int64) ([]*EcosystemTx, error) {
	curDate := time.Unix(timeBlock, 0).Format(`2006-01-02`)
	sql := `SELECT
		EXTRACT(EPOCH FROM to_timestamp(bc.time)::date)::int "unix_time",
		SUBSTRING(rtx.table_name FROM '^\d+') "ecosystem",
		COUNT(*)
	FROM rollback_tx rtx
		INNER JOIN block_chain bc ON bc.id = rtx.block_id
	WHERE to_timestamp(bc.time)::date >= (DATE('` + curDate + `') - interval '1' day)::date
	GROUP BY unix_time, ecosystem`

	var ecosystemTx []*EcosystemTx
	err := DBConn.Raw(sql).Scan(&ecosystemTx).Error
	if err != nil {
		return nil, err
	}

	return ecosystemTx, err
}

// GetMetricValues returns aggregated metric values in the time interval
func GetMetricValues(metric, timeInterval, aggregateFunc, timeBlock string) ([]interface{}, error) {
	rows, err := DBConn.Table(tableNameMetrics).Select("key,"+aggregateFunc+"(value)").
		Where("metric = ? AND time >= EXTRACT(EPOCH FROM TIMESTAMP ? - CAST(? AS INTERVAL))",
			metric, timeBlock, timeInterval).
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
