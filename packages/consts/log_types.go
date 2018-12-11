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

// LogEventType is storing numeric event type
type LogEventType int

// Types of log errors
const (
	NetworkError             = "Network"
	JSONMarshallError        = "JSONMarshall"
	JSONUnmarshallError      = "JSONUnmarshall"
	CommandExecutionError    = "CommandExecution"
	ConversionError          = "Conversion"
	TypeError                = "Type"
	ProtocolError            = "Protocol"
	MarshallingError         = "Marshall"
	UnmarshallingError       = "Unmarshall"
	ParseError               = "Parse"
	IOError                  = "IO"
	CryptoError              = "Crypto"
	ContractError            = "Contract"
	DBError                  = "DB"
	PanicRecoveredError      = "Panic"
	ConnectionError          = "Connection"
	ConfigError              = "Config"
	VMError                  = "VM"
	JustWaiting              = "JustWaiting"
	BlockError               = "Block"
	ParserError              = "Parser"
	ContextError             = "Context"
	SessionError             = "Session"
	RouteError               = "Route"
	NotFound                 = "NotFound"
	Found                    = "Found"
	EmptyObject              = "EmptyObject"
	InvalidObject            = "InvalidObject"
	DuplicateObject          = "DuplicateObject"
	UnknownObject            = "UnknownObject"
	ParameterExceeded        = "ParameterExceeded"
	DivisionByZero           = "DivisionByZero"
	EvalError                = "Eval"
	JWTError                 = "JWT"
	AccessDenied             = "AccessDenied"
	SizeDoesNotMatch         = "SizeDoesNotMatch"
	NoIndex                  = "NoIndex"
	NoFunds                  = "NoFunds"
	BlockIsFirst             = "BlockIsFirst"
	IncorrectCallingContract = "IncorrectCallingContract"
	WritingFile              = "WritingFile"
	CentrifugoError          = "CentrifugoError"
	StatsdError              = "StatsdError"
	MigrationError           = "MigrationError"
	AutoupdateError          = "AutoupdateError"
	BCRelevanceError         = "BCRelevanceError"
	BCActualizationError     = "BCActualizationError"
	SchedulerError           = "SchedulerError"
	SyncProcess              = "SyncProcess"
	WrongModeError           = "WrongModeError"
	VDEManagerError          = "VDEManagerError"
	QueueError               = "QueueError"
	LevelDBError             = "LevelDBError"
	TCPClientError           = "TCPClientError"
	BadTxError               = "BadTxError"
	TimeCalcError            = "BlockTimeCounterError"
)
