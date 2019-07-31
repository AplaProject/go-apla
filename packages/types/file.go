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

package types

//type File *Map

func NewFile() *Map {
	return LoadMap(map[string]interface{}{
		"Name":     "",
		"MimeType": "",
		"Body":     []byte{},
	})
}

func NewFileFromMap(m map[interface{}]interface{}) (f *Map, ok bool) {
	var v interface{}
	f = NewFile()

	if v, ok = m["Name"].(string); !ok {
		return
	}
	f.Set("Name", v)
	if v, ok = m["MimeType"].(string); !ok {
		return
	}
	f.Set("MimeType", v)
	if v, ok = m["Body"].([]byte); !ok {
		return
	}
	f.Set("Body", v)

	return
}
