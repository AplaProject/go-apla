package consts

type LogEventType int

const (
	NetworkError          = "Network"
	JSONMarshallError     = "JSONMarshall"
	JSONUnmarshallError   = "JSONUnmarshall"
	CommandExecutionError = "CommandExecution"
	ConvertionError       = "Convertion"
	TypeError             = "Type"
	ProtocolError         = "Protocol"
	MarshallingError      = "Marshall"
	UnmarshallingError    = "Unmarshall"
	ParseError            = "Parse"
	IOError               = "IO"
	CryptoError           = "Crypto"
	DBError               = "DB"
	PanicRecoveredError   = "Panic"
	ConnectionError       = "Connection"
	ConfigError           = "Config"
	VMError               = "VM"
	JustWaiting           = "JustWaiting"
	BlockError            = "Block"
	ParserError           = "Parser"
	ContextError          = "Context"
	SessionError          = "Session"
	RouteError            = "Route"
)
