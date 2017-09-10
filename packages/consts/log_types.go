package consts

type LogEventType int

const (
	StrToIntError       LogEventType = iota
	StrToFloatError                  = iota
	PanicRecoveredError              = iota
	RouteError                       = iota
	SessionError                     = iota
	FuncStarted                      = iota
	DBError                          = iota
	RecordNotFoundError              = iota
	CryptoError                      = iota
	GetHeaderError                   = iota
	SendEmbeddedTxError              = iota
	ConfigError                      = iota
	JustWaiting                      = iota
	BlockError                       = iota
	PrivateKeyError                  = iota
	ParserError                      = iota
	DebugMessage                     = iota
	BlockchainLoadError              = iota
	ContextError                     = iota
	ConnectionError                  = iota
	NodeBan                          = iota
	IOError                          = iota
	DaemonError                      = iota
	ConverterError                   = iota
)

var LogEventsMap = map[LogEventType]string{
	StrToIntError:       "can't convert to int",
	StrToFloatError:     "can't convert to float",
	PanicRecoveredError: "recovered after panic",
	RouteError:          "incorrect route parameters",
	SessionError:        "session is undefined",
	DBError:             "database error",
	RecordNotFoundError: "record not found",
	CryptoError:         "crypto error",
	GetHeaderError:      "can't get request header",
	SendEmbeddedTxError: "send embedded tx error",
	ConfigError:         "config error",
	BlockError:          "block error",
	PrivateKeyError:     "private key error",
	ParserError:         "parser error",
	BlockchainLoadError: "blockchain load error",
	ContextError:        "context error",
	ConnectionError:     "connection error",
	IOError:             "io error",
	DaemonError:         "daemon error",
	ConverterError:      "converter error",

	FuncStarted:  "function started",
	JustWaiting:  "just waiting",
	DebugMessage: "debug message",
	NodeBan:      "node banned",
}
