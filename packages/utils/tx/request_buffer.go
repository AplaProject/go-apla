package tx

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

type Request struct {
	ID        string
	Time      time.Time
	Contracts []*RequestContract
}

func (r *Request) NewContract(contract string) *RequestContract {
	return &RequestContract{
		contract: contract,
		values:   make(map[string]string),
		files:    make(map[string]*FileField),
	}
}

func (r *Request) AddContract(contract *RequestContract) {
	r.Contracts = append(r.Contracts, contract)
}

func (r *Request) clean() {
	for _, c := range r.Contracts {
		c.clean()
	}
}

type RequestContract struct {
	contract string
	values   map[string]string
	files    map[string]*FileField
}

func (rc *RequestContract) Contract() string {
	return rc.contract
}

func (rc *RequestContract) SetParam(key, value string) {
	rc.values[key] = value
}

func (rc *RequestContract) GetParam(key string) string {
	return rc.values[key]
}

func (rc *RequestContract) WriteFile(key, mimeType string, reader io.ReadCloser) (*FileHeader, error) {
	file, err := ioutil.TempFile(conf.Config.TempDir, "")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err = io.Copy(file, io.TeeReader(reader, hash)); err != nil {
		return nil, err
	}

	fileHeader := FileHeader{
		Hash:     hex.EncodeToString(hash.Sum(nil)),
		MimeType: mimeType,
	}

	rc.files[key] = &FileField{
		FileHeader: fileHeader,
		Path:       file.Name(),
	}

	return &fileHeader, nil
}

func (rc *RequestContract) ReadFile(key string) (*File, error) {
	fileField, ok := rc.files[key]
	if !ok {
		return nil, nil
	}

	data, err := ioutil.ReadFile(fileField.Path)
	if err != nil {
		return nil, err
	}

	return &File{
		FileHeader: FileHeader{
			Hash:     fileField.Hash,
			MimeType: fileField.MimeType,
		},
		Data: data,
	}, nil
}

func (rc *RequestContract) clean() {
	for _, f := range rc.files {
		os.Remove(f.Path)
	}
}

type FileHeader struct {
	Hash     string
	MimeType string
}

type FileField struct {
	FileHeader
	Path string
}

type File struct {
	FileHeader
	Data []byte
}

type RequestBuffer struct {
	mutex sync.Mutex

	timer         *time.Timer
	requestExpire time.Duration

	requests map[string]*Request
}

func (rb *RequestBuffer) ExpireDuration() time.Duration {
	return rb.requestExpire
}

func (rb *RequestBuffer) NewRequest() *Request {
	r := &Request{
		ID:        utils.UUID(),
		Time:      time.Now(),
		Contracts: make([]*RequestContract, 0),
	}

	return r
}

func (rb *RequestBuffer) AddRequest(r *Request) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.requests[r.ID] = r
	rb.timer.Reset(rb.requestExpire)
}

func (rb *RequestBuffer) GetRequest(id string) (*Request, bool) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	r, ok := rb.requests[id]
	if !ok {
		return nil, false
	}

	return r, true
}

func (rb *RequestBuffer) waitForCleaning() {
	for t := range rb.timer.C {
		rb.clean(t)
	}
}

func (rb *RequestBuffer) clean(t time.Time) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	for id, r := range rb.requests {
		if t.Sub(r.Time) > rb.requestExpire {
			r.clean()
			delete(rb.requests, id)
		}
	}
}

func NewRequestBuffer(requestExpire time.Duration) *RequestBuffer {
	rb := &RequestBuffer{
		requests:      make(map[string]*Request),
		timer:         time.NewTimer(-1),
		requestExpire: requestExpire,
	}

	go rb.waitForCleaning()

	return rb
}
