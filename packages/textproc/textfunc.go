// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package textproc

import (
	"fmt"
)

// Link returns HREF tag where pars[0] - url, pars[1] - text, pars[2] - title
func Link(vars *map[string]string, pars ...string) string {
	var (
		title, name string
	)
	if len(pars) < 1 {
		return ``
	}
	if len(pars) > 1 {
		name = pars[1]
	}
	if len(pars) > 2 {
		title = pars[2]
	}
	return fmt.Sprintf(`<a href="%s" title="%s">%s</a>`, pars[0], title, name)
}

// Tag returns HTML tag where pars[0] - tag's name, pars[1] - body of the tag
func Tag(vars *map[string]string, pars ...string) string {
	if len(pars) != 2 || pars[0] == `script` {
		return ``
	}
	return fmt.Sprintf(`<%s>%s</%[1]s>`, pars[0], pars[1])
}

// Break returns BR tag
func Break(vars *map[string]string, pars ...string) string {
	return `<br>`
}
