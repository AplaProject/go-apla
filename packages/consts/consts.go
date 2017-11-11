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

package consts

// Current version
const VERSION = "0.1.6b9"
const BLOCK_VERSION = 1

const FIRST_QDLT = 1e+26
const EGS_DIGIT = 18 //money_digit for EGS 1000000000000000000

// is used in confirmations
const WAIT_CONFIRMED_NODES = 10

// The number of nodes which should have the same block as we have for regarding this block belongs to the major part of DC-net. For get_confirmed_block_id()
const MIN_CONFIRMED_NODES = 0

const DOWNLOAD_CHAIN_TRY_COUNT = 10

// How fast could the time of transaction pass
const MAX_TX_FORW = 0

// transaction may wander in the net for a day and then get into a block
const MAX_TX_BACK = 86400

const ERROR_TIME = 1

const ROUND_FIX = 0.00000000001

// timeouts for TCP
const READ_TIMEOUT = 20
const WRITE_TIMEOUT = 20

const TCP_PORT = "7078"

const DATA_TYPE_MAX_BLOCK_ID = 10
const DATA_TYPE_BLOCK_BODY = 7

const UPD_AND_VER_URL = "http://apla.io"

var AddressLength = 20
var PubkeySizeLength = 64
var PrivkeyLength = 32
var BlockSize = 16

// TxTypes is the list of the embedded transactions
var TxTypes = map[int]string{
	1:  "FirstBlock",
}

func init() {
}
