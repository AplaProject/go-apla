package tx

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

type Request struct {
	ID       string
	Time     time.Time
	Contract string
	values   map[string]string
	files    map[string]*FileField
}

func (r *Request) SetValue(key, value string) {
	r.values[key] = value
}

func (r *Request) GetValue(key string) string {
	return r.values[key]
}

func (r *Request) WriteFile(key, mimeType string, reader io.ReadCloser) (*FileHeader, error) {
	filePath := path.Join(conf.Config.TempDir, utils.UUID())
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
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

	r.files[key] = &FileField{
		FileHeader: fileHeader,
		Path:       filePath,
	}

	return &fileHeader, nil
}

func (r *Request) ReadFile(key string) (*File, error) {
	fileField, ok := r.files[key]
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

func (rb *RequestBuffer) NewRequest(contract string) *Request {
	r := &Request{
		ID:       utils.UUID(),
		Time:     time.Now(),
		Contract: contract,
		values:   make(map[string]string),
		files:    make(map[string]*FileField),
	}

	rb.AddRequest(r)

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
			for _, fileField := range r.files {
				os.Remove(fileField.Path)
			}
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
