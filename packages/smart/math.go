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
	"math"
	"strconv"
)

func parseFloat(x interface{}) (float64, error) {
	var (
		fx  float64
		err error
	)
	switch v := x.(type) {
	case float64:
		fx = v
	case int64:
		fx = float64(v)
	case string:
		if fx, err = strconv.ParseFloat(v, 64); err != nil {
			return 0, errFloat
		}
	default:
		return 0, errFloat
	}
	return fx, nil
}

func isValidFloat(x float64) bool {
	return !(math.IsNaN(x) || math.IsInf(x, 1) || math.IsInf(x, -1))
}

// Floor returns the greatest integer value less than or equal to x
func Floor(x interface{}) (int64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Floor(fx); isValidFloat(fx) {
		return int64(fx), nil
	}
	return 0, errFloatResult
}

// Log returns the natural logarithm of x
func Log(x interface{}) (float64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Log(fx); isValidFloat(fx) {
		return fx, nil
	}
	return 0, errFloatResult
}

// Log10 returns the decimal logarithm of x
func Log10(x interface{}) (float64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Log10(fx); isValidFloat(fx) {
		return fx, nil
	}
	return 0, errFloatResult
}

// Pow returns x**y, the base-x exponential of y
func Pow(x, y interface{}) (float64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	fy, err := parseFloat(y)
	if err != nil {
		return 0, err
	}
	if fx = math.Pow(fx, fy); isValidFloat(fx) {
		return fx, nil
	}
	return 0, errFloatResult
}

// Round returns the nearest integer, rounding half away from zero
func Round(x interface{}) (int64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Round(fx); isValidFloat(fx) {
		return int64(fx), nil
	}
	return 0, errFloatResult
}

// Sqrt returns the square root of x
func Sqrt(x interface{}) (float64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Sqrt(fx); isValidFloat(fx) {
		return fx, nil
	}
	return 0, errFloatResult
}
