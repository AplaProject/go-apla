package resp

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

// SyncPolicy represents a file's fsync policy.
type SyncPolicy int

const (
	Never       SyncPolicy = iota // The policy 'Never' means fsync is never called, the Operating System will be in charge of your data. This is the fastest and less safe method.
	EverySecond SyncPolicy = iota // The policy 'EverySecond' means that fsync will be called every second or so. This is fast enough, and at most you can lose one second of data if there is a disaster.
	Always      SyncPolicy = iota // The policy 'Always' means fsync is called after every write. This is super duper safe and very incredibly slow.
)

// String returns a string respesentation.
func (policy SyncPolicy) String() string {
	switch policy {
	default:
		return "unknown"
	case Never:
		return "never"
	case EverySecond:
		return "every second"
	case Always:
		return "always"
	}
}

var errClosed = errors.New("closed")

// AOF represents an open file descriptor.
type AOF struct {
	mu     sync.Mutex
	f      *os.File
	closed bool
	rd     *Reader
	policy SyncPolicy
	atEnd  bool
}

// OpenAOF will open and return an AOF file. If the file does not exist a new one will be created.
func OpenAOF(path string) (*AOF, error) {
	var err error
	aof := &AOF{}
	aof.f, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	aof.policy = EverySecond
	go func() {
		for {
			aof.mu.Lock()
			if aof.closed {
				aof.mu.Unlock()
				return
			}
			if aof.policy == EverySecond {
				aof.f.Sync()
			}
			aof.mu.Unlock()
			time.Sleep(time.Second)
		}
	}()
	return aof, nil
}

// SetSyncPolicy set the sync policy of the file.
// The policy 'EverySecond' means that fsync will be called every second or so. This is fast enough, and at most you can lose one second of data if there is a disaster.
// The policy 'Never' means fsync is never called, the Operating System will be in charge of your data. This is the fastest and less safe method.
// The policy 'Always' means fsync is called after every write. This is super duper safe and very incredibly slow.
// EverySecond is the default.
func (aof *AOF) SetSyncPolicy(policy SyncPolicy) {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	if aof.policy == policy {
		return
	}
	switch policy {
	default:
		return
	case Never, EverySecond, Always:
	}
	aof.policy = policy
}

// Close will close the file.
func (aof *AOF) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	if aof.closed {
		return errClosed
	}
	aof.f.Close()
	aof.closed = true
	return nil
}

func (aof *AOF) readValues(iterator func(v Value)) error {
	aof.atEnd = false
	if _, err := aof.f.Seek(0, 0); err != nil {
		return err
	}
	rd := NewReader(aof.f)
	for {
		v, _, err := rd.ReadValue()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if iterator != nil {
			iterator(v)
		}
	}
	if _, err := aof.f.Seek(0, 2); err != nil {
		return err
	}
	aof.atEnd = true
	return nil
}

// Append writes a value to the end of the file.
func (aof *AOF) Append(v Value) error {
	return aof.AppendMulti([]Value{v})
}

// AppendMulti writes multiple values to the end of the file.
// This operation can increase performance over calling multiple Append()s and also has the benefit of transactional writes.
func (aof *AOF) AppendMulti(vs []Value) error {
	var bs []byte
	for _, v := range vs {
		b, err := v.MarshalRESP()
		if err != nil {
			return err
		}
		if bs == nil {
			bs = b
		} else {
			bs = append(bs, b...)
		}
	}
	aof.mu.Lock()
	defer aof.mu.Unlock()
	if aof.closed {
		return errClosed
	}
	if !aof.atEnd {
		if err := aof.readValues(nil); err != nil {
			return err
		}
	}
	_, err := aof.f.Write(bs)
	if err != nil {
		return err
	}
	if aof.policy == Always {
		aof.f.Sync()
	}
	return nil
}

// Scan iterates though all values in the file.
// This operation could take a long time if there lots of values, and the operation cannot be canceled part way through.
func (aof *AOF) Scan(iterator func(v Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	if aof.closed {
		return errClosed
	}
	return aof.readValues(iterator)
}
