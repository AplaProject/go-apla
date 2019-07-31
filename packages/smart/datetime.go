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

package smart

import (
	"time"

	"github.com/pkg/errors"
)

const (
	dateTimeFormat = "2006-01-02 15:04:05"
)

// Date formats timestamp to specified date format
func Date(timeFormat string, timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format(timeFormat)
}

func BlockTime(sc *SmartContract) string {
	var blockTime int64
	if sc.BlockData != nil {
		blockTime = sc.BlockData.Time
	}
	if sc.OBS {
		blockTime = time.Now().Unix()
	}
	return Date(dateTimeFormat, blockTime)
}

func DateTime(unix int64) string {
	return Date(dateTimeFormat, unix)
}

func DateTimeLocation(unix int64, locationName string) (string, error) {
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		return "", errors.Wrap(err, "Load location")
	}

	return time.Unix(unix, 0).In(loc).Format(dateTimeFormat), nil
}

func UnixDateTime(value string) int64 {
	t, err := time.Parse(dateTimeFormat, value)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func UnixDateTimeLocation(value, locationName string) (int64, error) {
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		return 0, errors.Wrap(err, "Load location")
	}

	t, err := time.ParseInLocation(dateTimeFormat, value, loc)
	if err != nil {
		return 0, errors.Wrap(err, "Parse time")
	}

	return t.Unix(), nil
}
