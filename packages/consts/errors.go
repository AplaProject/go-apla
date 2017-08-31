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
)

var LogEventsMap = map[LogEventType]string{
	StrToIntError:       "can't convert to int",
	StrToFloatError:     "can't convert to float",
	PanicRecoveredError: "recovered after panic",
	RouteError:          "incorrect route parameters",
	SessionError:        "session is undefined",
	FuncStarted:         "function started",
	DBError:             "database error",
	RecordNotFoundError: "record not found",
	CryptoError:         "crypto error",
	GetHeaderError:      "can't get request header",
	SendEmbeddedTxError: "send embedded tx error",
}
