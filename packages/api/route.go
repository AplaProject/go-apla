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

package api

import (
	"strings"

	hr "github.com/julienschmidt/httprouter"
)

func Route(route *hr.Router) {
	anyMethod := func(method, pattern, pars string, handler apiHandle) {
		params := make(map[string]int)
		for _, par := range strings.Split(pars, `,`) {
			var vtype int
			types := strings.Split(par, `:`)
			if len(types) != 2 {
				continue
			}
			switch types[1] {
			case `hex`:
				vtype = pHex
			case `int64`:
				vtype = pInt64
			default:
				continue
			}
			vars := strings.Split(types[0], ` `)
			for _, v := range vars {
				if len(v) > 0 {
					if v[0] == '?' {
						if len(v) > 1 {
							params[v[1:]] = vtype | pOptional
						}
					} else {
						params[v] = vtype
					}
				}
			}
		}
		route.Handle(method, `/api/v1/`+pattern, DefaultHandler(params, handler))
	}
	get := func(pattern, params string, handler apiHandle) {
		anyMethod(`GET`, pattern, params, handler)
	}
	post := func(pattern, params string, handler apiHandle) {
		anyMethod(`POST`, pattern, params, handler)
	}
	get(`getuid`, ``, getUID)
	post(`auth`, `pubkey signature:hex,?state:int64`, auth)
}
