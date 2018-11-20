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

package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// This is the modified version of https://github.com/jabong/florest-core/
// https://github.com/jabong/florest-core/blob/master/src/common/collections/maps/linkedhashmap/linkedhashmap.go

// Link represents a node of doubly linked list
type Link struct {
	key   string
	value interface{}
	next  *Link
	prev  *Link
}

// Map holds the elements in go's native map, also maintains the head and tail link
// to keep the elements in insertion order
type Map struct {
	m    map[string]*Link
	head *Link
	tail *Link
}

func newLink(key string, value interface{}) *Link {
	return &Link{key: key, value: value, next: nil, prev: nil}
}

// NewMap instantiates a linked hash map.
func NewMap() *Map {
	return &Map{m: make(map[string]*Link), head: nil, tail: nil}
}

func ConvertMap(in interface{}) interface{} {
	switch v := in.(type) {
	case map[string]interface{}:
		out := NewMap()
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			out.Set(key, ConvertMap(v[key]))
		}
		return out
	case []interface{}:
		for i, item := range v {
			v[i] = ConvertMap(item)
		}
	}
	return in
}

// LoadMap instantiates a linked hash map and initializing it from map[string]interface{}.
func LoadMap(init map[string]interface{}) (ret *Map) {
	ret = NewMap()
	for key, val := range init {
		ret.Set(key, ConvertMap(val))
	}
	return
}

// Put inserts an element into the map.
func (m *Map) Set(key string, value interface{}) {
	link, found := m.m[key]
	if !found {
		link = newLink(key, value)
		if m.tail == nil {
			m.head = link
			m.tail = link
		} else {
			m.tail.next = link
			link.prev = m.tail
			m.tail = link
		}
		m.m[key] = link
	} else {
		link.value = value
	}
}

// Get searches the element in the map by key and returns its value or nil if key doesn't exists.
// Second return parameter is true if key was found, otherwise false.
func (m *Map) Get(key string) (value interface{}, found bool) {
	var link *Link
	link, found = m.m[key]
	if found {
		value = link.value
	} else {
		value = nil
	}
	return
}

// Remove removes the element from the map by key.
func (m *Map) Remove(key string) {
	link, found := m.m[key]
	if found {
		delete(m.m, key)
		if m.head == link && m.tail == link {
			m.head = nil
			m.tail = nil
		} else if m.tail == link {
			m.tail = link.prev
			link.prev.next = nil
		} else if m.head == link {
			m.head = link.next
			link.next.prev = nil
		} else {
			link.prev.next = link.next
			link.next.prev = link.prev
		}
	}
}

// IsEmpty returns true if map does not contain any elements
func (m *Map) IsEmpty() bool {
	return m == nil || m.Size() == 0
}

// Size returns number of elements in the map.
func (m *Map) Size() int {
	return len(m.m)
}

// Keys returns all keys of the map (insertion order).
func (m *Map) Keys() []string {
	keys := make([]string, m.Size())
	count := 0
	for current := m.head; current != nil; current = current.next {
		keys[count] = current.key
		count++
	}
	return keys
}

// Values returns all values of the map (insertion order).
func (m *Map) Values() []interface{} {
	values := make([]interface{}, m.Size())
	count := 0
	for current := m.head; current != nil; current = current.next {
		values[count] = current.value
		count++
	}
	return values
}

// Clear removes all elements from the map.
func (m *Map) Clear() {
	m.m = make(map[string]*Link)
	m.head = nil
	m.tail = nil
}

// String returns a string representation of container
func (m *Map) String() string {
	str := "map["
	for current := m.head; current != nil; current = current.next {
		str += fmt.Sprintf("%v:%v ", current.key, current.value)
	}
	return strings.TrimRight(str, " ") + "]"
}

func (m *Map) MarshalJSON() ([]byte, error) {
	s := "{"
	for current := m.head; current != nil; current = current.next {
		k := current.key
		escaped := strings.Replace(k, `"`, `\"`, -1)
		s = s + `"` + escaped + `":`
		v := current.value
		vBytes, err := json.Marshal(v)
		if err != nil {
			return []byte{}, err
		}
		s = s + string(vBytes) + ","
	}
	if len(s) > 1 {
		s = s[0 : len(s)-1]
	}
	s = s + "}"
	return []byte(s), nil
}
