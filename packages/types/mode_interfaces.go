package types

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// ClientTxPreprocessor procees tx from client
type ClientTxPreprocessor interface {
	ProcessClientTranstaction([]byte, int64, *log.Entry) (string, error)
}

// SmartContractRunner run serialized contract
type SmartContractRunner interface {
	RunContract(data, hash []byte, keyID int64, le *log.Entry) error
}

type DaemonListFactory interface {
	GetDaemonsList() []string
}

type EcosystemLookupGetter interface {
	GetEcosystemLookup() ([]int64, []string, error)
}

type EcosystemIDValidator interface {
	Validate(id, clientID int64, le *log.Entry) (int64, error)
}

// DaemonLoader allow implement different ways for loading daemons
type DaemonLoader interface {
	Load(context.Context) error
}

type EcosystemNameGetter interface {
	GetEcosystemName(id int64) (string, error)
}
