// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package consts

import (
	"time"
)

// VERSION is current version
const VERSION = "2.0.3"

const BV_ROLLBACK_HASH = 2
const BV_INCLUDE_ROLLBACK_HASH = 3

// BLOCK_VERSION is block version
const BLOCK_VERSION = BV_INCLUDE_ROLLBACK_HASH

// DEFAULT_TCP_PORT used when port number missed in host addr
const DEFAULT_TCP_PORT = 7078

// FounderAmount is the starting amount of founder
const FounderAmount = 100000000

// MoneyDigits is numbers of digits for tokens 1000000000000000000
const MoneyDigits = 18

// WAIT_CONFIRMED_NODES is used in confirmations
const WAIT_CONFIRMED_NODES = 10

// MIN_CONFIRMED_NODES The number of nodes which should have the same block as we have for regarding this block belongs to the major part of DC-net. For get_confirmed_block_id()
const MIN_CONFIRMED_NODES = 0

// DOWNLOAD_CHAIN_TRY_COUNT is number of attempt
const DOWNLOAD_CHAIN_TRY_COUNT = 10

// MAX_TX_FORW How fast could the time of transaction pass
const MAX_TX_FORW = 600

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

// HashSize is size of hash
const HashSize = 32

const AvailableBCGap = 4

const DefaultNodesConnectDelay = 6

const MaxTXAttempt = 10

const (
	TxTypeFirstBlock  = 1
	TxTypeStopNetwork = 2

	TxTypeParserFirstBlock  = "FirstBlock"
	TxTypeParserStopNetwork = "StopNetwork"
)

// TxTypes is the list of the embedded transactions
var TxTypes = map[int64]string{
	TxTypeFirstBlock:  TxTypeParserFirstBlock,
	TxTypeStopNetwork: TxTypeParserStopNetwork,
}

// ApiPath is the beginning of the api url
var ApiPath = `/api/v2/`

// BuildInfo should be defined through -ldflags
var BuildInfo string

const (
	// DefaultConfigFile name of config file (toml format)
	DefaultConfigFile = "config.toml"

	// DefaultWorkdirName name of working directory
	DefaultWorkdirName = "apla-data"

	// DefaultPidFilename is default filename of pid file
	DefaultPidFilename = "go-apla.pid"

	// DefaultLockFilename is default filename of lock file
	DefaultLockFilename = "go-apla.lock"

	// FirstBlockFilename name of first block binary file
	FirstBlockFilename = "1block"

	// PrivateKeyFilename name of wallet private key file
	PrivateKeyFilename = "PrivateKey"

	// PublicKeyFilename name of wallet public key file
	PublicKeyFilename = "PublicKey"

	// NodePrivateKeyFilename name of node private key file
	NodePrivateKeyFilename = "NodePrivateKey"

	// NodePublicKeyFilename name of node public key file
	NodePublicKeyFilename = "NodePublicKey"

	// KeyIDFilename generated KeyID
	KeyIDFilename = "KeyID"

	// RollbackResultFilename rollback result file
	RollbackResultFilename = "rollback_result"

	// FromToPerDayLimit day limit token transfer between accounts
	FromToPerDayLimit = 10000

	// TokenMovementQtyPerBlockLimit block limit token transfer
	TokenMovementQtyPerBlockLimit = 100

	// TCPConnTimeout timeout of tcp connection
	TCPConnTimeout = 5 * time.Second

	// TxRequestExpire is expiration time for request of transaction
	TxRequestExpire = 1 * time.Minute

	// DefaultTempDirName is default name of temporary directory
	DefaultTempDirName = "apla-temp"

	// DefaultOBS allways is 1
	DefaultOBS = 1

	// MoneyLength is the maximum number of digits in money value
	MoneyLength = 30

	TokenEcosystem = 1

	HTTPServerMaxBodySize = 1 << 20

	// ShiftContractID is the offset of tx identifiers
	ShiftContractID = 5000

	// ContractList is the number of contracts per page on loading
	ContractList = 200

	// Guest key
	GuestPublic  = "489347a1205c818d9a02f285faaedd0122a56138e3d985f5e1b4f6a9470f90f692a00a3453771dd7feea388ceb7aefeaf183e299c70ad1aecb7f870bfada3b86"
	GuestKey     = "8529184409882256911"
	GuestAddress = "0852-9184-4098-8225-6911"

	// StatusMainPage is a status for Main Page
	StatusMainPage = `2`

	NoneVDE     = "none"
	DBFindLimit = 10000
)
