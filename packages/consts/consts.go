// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package consts

// VERSION is current version
const VERSION = "0.1.6b13"

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

const HashSize = 32

// TxTypes is the list of the embedded transactions
var TxTypes = map[int]string{
	1: "FirstBlock",
}

// ApiPath is the beginning of the api url
var ApiPath = `/api/v2/`

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
