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

package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WsMsg is a structure for sending information about transactions.
type WsMsg struct {
	Data   []TxInfo `json:"data"`
	Latest int64    `json:"latest"`
}

// WsBlockchain is a handle websocket function.
func WsBlockchain(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		if latest, err := strconv.ParseInt(string(msg), 10, 64); err == nil && latest >= 0 {
			var answer WsMsg
			answer.Data = make([]TxInfo, 0)
			answer.Latest = txTop.ID
			start := txTop
			for start.ID > latest && len(answer.Data) < 20 {
				answer.Data = append(answer.Data, *start)
				start = start.prev
			}
			err = conn.WriteJSON(answer)
			if err != nil {
				break
			}
		} else {
			break
		}
	}
	if err = conn.Close(); err != nil {
		fmt.Println(err)
	}
}
