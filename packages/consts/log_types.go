package consts

type LogEventType int

const (
	StrToIntError          LogEventType = iota
	StrToFloatError                     = iota
	StrToDecimalError                   = iota
	PanicRecoveredError                 = iota
	RouteError                          = iota
	SessionError                        = iota
	FuncStarted                         = iota
	DBError                             = iota
	RecordNotFoundError                 = iota
	CryptoError                         = iota
	GetHeaderError                      = iota
	SendEmbeddedTxError                 = iota
	ConfigError                         = iota
	JustWaiting                         = iota
	BlockError                          = iota
	PrivateKeyError                     = iota
	ParserError                         = iota
	DebugMessage                        = iota
	BlockchainLoadError                 = iota
	ContextError                        = iota
	ConnectionError                     = iota
	NodeBan                             = iota
	IOError                             = iota
	DaemonError                         = iota
	ConverterError                      = iota
	CommandError                        = iota
	JSONError                           = iota
	ContractError                       = iota
	RollbackError                       = iota
	SystemParamsError                   = iota
	SystemError                         = iota
	InputError                          = iota
	APIParamsError                      = iota
	IncompatibleTypesError              = iota
	RequestConditionError               = iota
	InnerError                          = iota
	TCPCserverError                     = iota
	TransactionError                    = iota
	TemplateError                       = iota
	SignError                           = iota
	VMEvent                             = iota
	VMError                             = iota
)

var LogEventsMap = map[LogEventType]string{
	StrToIntError:          "can't convert to int",
	StrToFloatError:        "can't convert to float",
	StrToDecimalError:      "str to decimal error",
	PanicRecoveredError:    "recovered after panic",
	RouteError:             "incorrect route parameters",
	SessionError:           "session is undefined",
	DBError:                "database error",
	RecordNotFoundError:    "record not found",
	CryptoError:            "crypto error",
	GetHeaderError:         "can't get request header",
	SendEmbeddedTxError:    "send embedded tx error",
	ConfigError:            "config error",
	BlockError:             "block error",
	PrivateKeyError:        "private key error",
	ParserError:            "parser error",
	BlockchainLoadError:    "blockchain load error",
	ContextError:           "context error",
	ConnectionError:        "connection error",
	IOError:                "io error",
	DaemonError:            "daemon error",
	ConverterError:         "converter error",
	CommandError:           "Command error",
	JSONError:              "JSON error",
	ContractError:          "Contract error",
	RollbackError:          "Rollback error",
	SystemParamsError:      "system params error",
	SystemError:            "System error",
	InputError:             "Input error",
	APIParamsError:         "API params error",
	IncompatibleTypesError: "Incompatible types error",
	RequestConditionError:  "Request conditions error",
	InnerError:             "error in inner function",
	TCPCserverError:        "tcp server error",
	TransactionError:       "transaction error",
	TemplateError:          "template error",
	SignError:              "sign error",
	VMEvent:                "VM event",
	VMError:                "VMError",
	FuncStarted:            "function started",
	JustWaiting:            "just waiting",
	DebugMessage:           "debug message",
	NodeBan:                "node banned",
}

const (
	NetworkError          = "Network error"
	JSONMarshallError     = "JSON marshall error"
	JSONUnmarshallError   = "JSON unmarshall error"
	CommandExecutionError = "Command execution error"
	ConvertionError       = "Convertion error"
	TypeError             = "Type error"
	ProtocolError         = "Protocol error"
	MarshallingError      = "Marshalling error"
	UnmarshallingError    = "Unmarshalling error"
	ParseError            = "Parse error"
)
