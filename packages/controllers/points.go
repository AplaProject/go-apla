package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type PointsPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	UserId       int64
	Lang         map[string]string
	CountSignArr []int
	PointsStatus []map[string]string
	VotesOk      string
	MyPoints     int64
	Mean         float64
}

func (c *Controller) Points() (string, error) {

	// список отравленных нами запросов
	pointsStatus, err := c.GetAll("SELECT * FROM "+c.MyPrefix+"points_status WHERE user_id = ? ORDER BY time_start DESC", -1, c.SessUserId)

	myPoints, err := c.Single("SELECT points FROM points WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// среднее значение
	mean, err := c.Single("SELECT sum(points)/count(points) FROM points WHERE points > 0").Float64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	mean = utils.Round(mean*c.Variables.Float64["points_factor"], 0)

	// есть ли тр-ия с голосованием votes_complex за послдение 4 недели
	count, err := c.Single("SELECT count(user_id) FROM votes_miner_pct WHERE user_id  =  ? AND time > ?", c.SessUserId, utils.Time()-c.Variables.Int64["limit_votes_complex_period"]*2).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	votesOk := "NO"
	if count > 0 {
		votesOk = "YES"
	}

	TemplateStr, err := makeTemplate("points", "points", &PointsPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		SignData:     "",
		VotesOk:      votesOk,
		MyPoints:     myPoints,
		PointsStatus: pointsStatus,
		Mean:         mean})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
