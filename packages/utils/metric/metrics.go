// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package metric

import (
	"strconv"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

const (
	metricEcosystemPages   = "ecosystem_pages"
	metricEcosystemMembers = "ecosystem_members"
	metricEcosystemTx      = "ecosystem_tx"
)

// CollectMetricDataForEcosystemTables returns metrics for some tables of ecosystems
func CollectMetricDataForEcosystemTables(timeBlock int64) (metricValues []*Value, err error) {
	stateIDs, _, err := model.GetAllSystemStatesIDs()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get all system states ids")
		return nil, err
	}

	now := time.Unix(timeBlock, 0)
	unixDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Unix()

	for _, stateID := range stateIDs {
		var pagesCount, membersCount int64

		tablePrefix := strconv.FormatInt(stateID, 10)

		p := &model.Page{}
		p.SetTablePrefix(tablePrefix)
		if pagesCount, err = p.Count(); err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get count of pages")
			return nil, err
		}
		metricValues = append(metricValues, &Value{
			Time:   unixDate,
			Metric: metricEcosystemPages,
			Key:    tablePrefix,
			Value:  pagesCount,
		})

		m := &model.Member{}
		m.SetTablePrefix(tablePrefix)
		if membersCount, err = m.Count(); err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get count of members")
			return nil, err
		}
		metricValues = append(metricValues, &Value{
			Time:   unixDate,
			Metric: metricEcosystemMembers,
			Key:    tablePrefix,
			Value:  membersCount,
		})
	}

	return metricValues, nil
}

// CollectMetricDataForEcosystemTx returns metrics for transactions of ecosystems
func CollectMetricDataForEcosystemTx(timeBlock int64) (metricValues []*Value, err error) {
	ecosystemTx, err := model.GetEcosystemTxPerDay(timeBlock)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get ecosystem transactions by period")
		return nil, err
	}
	for _, item := range ecosystemTx {
		if len(item.Ecosystem) == 0 {
			continue
		}

		metricValues = append(metricValues, &Value{
			Time:   item.UnixTime,
			Metric: metricEcosystemTx,
			Key:    item.Ecosystem,
			Value:  item.Count,
		})
	}

	return metricValues, nil
}
