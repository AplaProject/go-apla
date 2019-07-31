// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

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
	OBSManagerError          = "OBSManagerError"
	TCPClientError           = "TCPClientError"
	BadTxError               = "BadTxError"
	TimeCalcError            = "BlockTimeCounterError"
)
