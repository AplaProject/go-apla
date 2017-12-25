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

// VERSION is current version
const VERSION = "0.1.6b9"

// BLOCK_VERSION is block version
const BLOCK_VERSION = 1

// DEFAULT_TCP_PORT used when port number missed in host addr
const DEFAULT_TCP_PORT = 7078

// FIRST_QDLT is default amount
const FIRST_QDLT = 1e+26

// EGS_DIGIT money_digit for EGS 1000000000000000000
const EGS_DIGIT = 18

// WAIT_CONFIRMED_NODES is used in confirmations
const WAIT_CONFIRMED_NODES = 10

// MIN_CONFIRMED_NODES The number of nodes which should have the same block as we have for regarding this block belongs to the major part of DC-net. For get_confirmed_block_id()
const MIN_CONFIRMED_NODES = 0

// DOWNLOAD_CHAIN_TRY_COUNT is number of attempt
const DOWNLOAD_CHAIN_TRY_COUNT = 10

// MAX_TX_FORW How fast could the time of transaction pass
const MAX_TX_FORW = 0

// MAX_TX_BACK transaction may wander in the net for a day and then get into a block
const MAX_TX_BACK = 86400

// ERROR_TIME is error time
const ERROR_TIME = 1

// ROUND_FIX is rounding constant
const ROUND_FIX = 0.00000000001

// READ_TIMEOUT is timeout for TCP
const READ_TIMEOUT = 20

// WRITE_TIMEOUT is timeout for TCP
const WRITE_TIMEOUT = 20

// DATA_TYPE_MAX_BLOCK_ID is block id max datatype
const DATA_TYPE_MAX_BLOCK_ID = 10

// DATA_TYPE_BLOCK_BODY is body block datatype
const DATA_TYPE_BLOCK_BODY = 7

// UPD_AND_VER_URL is root url
const UPD_AND_VER_URL = "http://apla.io"

// AddressLength is length of address
const AddressLength = 20

// PubkeySizeLength is pubkey length
const PubkeySizeLength = 64

// PrivkeyLength is privkey length
const PrivkeyLength = 32

// BlockSize is size of block
const BlockSize = 16

// TxTypes is the list of the embedded transactions
var TxTypes = map[int]string{
	1: "FirstBlock",
}

// DefaultConfigFile name of config file (toml format)
const DefaultConfigFile = "config.toml"

// PidFilename name of pid file
const PidFilename = "apla.pid"

// FirstBlockFilename name of first block binary file
const FirstBlockFilename = "1block"

// PrivateKeyFilename name of wallet private key file
const PrivateKeyFilename = "PrivateKey"

// PublicKeyFilename name of wallet public key file
const PublicKeyFilename = "PublicKey"

// NodePrivateKeyFilename name of node private key file
const NodePrivateKeyFilename = "NodePrivateKey"

// NodePublicKeyFilename name of node public key file
const NodePublicKeyFilename = "NodePublicKey"

// KeyIDFilename generated KeyID
const KeyIDFilename = "KeyID"

// RollbackResultFilename rollback result file
const RollbackResultFilename = "rollback_result"

// WellKnownRoute TLS route
const WellKnownRoute = "/.well-known/*filepath"

// TLSFullchainPem fullchain pem file
const TLSFullchainPem = "/fullchain.pem"

// TLSPrivkeyPem privkey pem file
const TLSPrivkeyPem = "/privkey.pem"
