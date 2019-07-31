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

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscape(t *testing.T) {
	var cases = []struct {
		Source   string
		Expected template.HTML
	}{
		{`'test'`, `''test''`},
		{"`test`", "` + \"`\" + `test` + \"`\" + `"},
		{`100%`, `100%%`},
	}

	for _, v := range cases {
		assert.Equal(t, v.Expected, escape(v.Source))
	}
}

func tempContract(appID int, conditions, value string) (string, error) {
	file, err := ioutil.TempFile("", "contract")
	if err != nil {
		return "", err
	}
	defer file.Close()

	file.Write([]byte(fmt.Sprintf(`// +prop AppID = %d
// +prop Conditions = '%s'
%s`, appID, conditions, value)))

	return file.Name(), nil
}

func TestLoadSource(t *testing.T) {
	value := "contract Test {}"

	path, err := tempContract(5, "true", value)
	assert.NoError(t, err)

	source, err := loadSource(path)
	assert.NoError(t, err)

	assert.Equal(t, &contract{
		Name:       filepath.Base(path),
		Source:     template.HTML(value + "\n"),
		Conditions: template.HTML("true"),
		AppID:      5,
	}, source)
}
