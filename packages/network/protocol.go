package network

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"

	log "github.com/sirupsen/logrus"
)

// Types of requests
const (
	RequestTypeFullNode        = 1
	RequestTypeNotFullNode     = 2
	RequestTypeStopNetwork     = 3
	RequestTypeConfirmation    = 4
	RequestTypeBlockCollection = 7
	RequestTypeMaxBlock        = 10

	// BlocksPerRequest contains count of blocks per request
	BlocksPerRequest int32 = 1000

	MaxBlockSize = 10485760
)

var ErrNotAccepted = errors.New("Not accepted")

// SelfReaderWriter read from Reader to himself and write to io.Writer from himself
type SelfReaderWriter interface {
	Read(io.Reader) error
	Write(io.Writer) error
}

// RequestType is type of request
type RequestType struct {
	Type uint16
}

// Read read first 2 bytes to uint16
func (rt *RequestType) Read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, &rt.Type)
}

func (rt *RequestType) Write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, rt.Type)
}

// MaxBlockRequest is max block request
type MaxBlockRequest struct{}

// MaxBlockResponse is max block response
type MaxBlockResponse struct {
	BlockID uint32
}

func (resp *MaxBlockResponse) Read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, &resp.BlockID)
}

func (resp *MaxBlockResponse) Write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, resp.BlockID)
}

// GetBodiesRequest contains BlockID
type GetBodiesRequest struct {
	BlockID      uint32
	ReverseOrder bool
}

func (req *GetBodiesRequest) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &req.BlockID); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading getBodiesRequest blockID")
		return err
	}

	order, err := readBool(r)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading GetBodiesRequest reverse order")
	}

	req.ReverseOrder = order
	return nil
}

func (req *GetBodiesRequest) Write(w io.Writer) error {

	if err := binary.Write(w, binary.LittleEndian, req.BlockID); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending GetBodiesRequest blockID")
		return err
	}

	if err := writeBool(w, req.ReverseOrder); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending GetBodiesRequest reverse order")
		return err
	}

	return nil
}

// GetBodyResponse is Data []bytes
type GetBodyResponse struct {
	Data []byte
}

func (resp *GetBodyResponse) Read(r io.Reader) error {
	slice, err := ReadSlice(r)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on reading GetBodyResponse")
		return err
	}

	resp.Data = slice
	return nil
}

func (resp *GetBodyResponse) Write(w io.Writer) error {
	return writeSlice(w, resp.Data)
}

// ConfirmRequest contains request data
type ConfirmRequest struct {
	BlockID uint32
}

func (req *ConfirmRequest) Read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, &req.BlockID)
}

func (req *ConfirmRequest) Write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, req.BlockID)
}

// ConfirmResponse contains response data
type ConfirmResponse struct {
	// ConfType uint8
	Hash []byte `size:"32"`
}

func (resp *ConfirmResponse) Read(r io.Reader) error {
	h, err := readSliceWithSize(r, consts.HashSize)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading ConfirmResponse reverse order")
		return err
	}
	resp.Hash = h
	return nil
}

func (resp *ConfirmResponse) Write(w io.Writer) error {
	if err := writeSliceWithSize(w, resp.Hash, consts.HashSize); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending ConfiremResponse hash")
		return err
	}

	return nil
}

// DisRequest contains request data
type DisRequest struct {
	Data []byte
}

func (req *DisRequest) Read(r io.Reader) error {
	slice, err := ReadSlice(r)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on reading disseminator request")
		return err
	}

	req.Data = slice
	return nil
}

func (req *DisRequest) Write(w io.Writer) error {
	err := writeSlice(w, req.Data)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on sending disseminator request")
	}

	return err
}

// DisTrResponse contains response data
type DisTrResponse struct{}

// DisHashResponse contains response data
type DisHashResponse struct {
	Data []byte
}

func (resp *DisHashResponse) Read(r io.Reader) error {
	slice, err := ReadSlice(r)
	if err != nil {
		return err
	}

	resp.Data = slice
	return nil
}

func (resp *DisHashResponse) Write(w io.Writer) error {
	return writeSlice(w, resp.Data)
}

type StopNetworkRequest struct {
	Data []byte
}

func (req *StopNetworkRequest) Read(r io.Reader) error {
	slice, err := ReadSlice(r)
	if err != nil {
		return err
	}

	req.Data = slice
	return nil
}

func (req *StopNetworkRequest) Write(w io.Writer) error {
	return writeSlice(w, req.Data)
}

type StopNetworkResponse struct {
	Hash []byte
}

func (resp *StopNetworkResponse) Read(r io.Reader) error {
	slice, err := ReadSlice(r)
	if err != nil {
		return err
	}

	resp.Hash = slice
	return nil
}

func (resp *StopNetworkResponse) Write(w io.Writer) error {
	return writeSlice(w, resp.Hash)
}

func readBool(r io.Reader) (bool, error) {
	var val uint8
	if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
		return false, err
	}

	return val > 0, nil
}

func writeBool(w io.Writer, val bool) error {
	var intVal int8
	if val {
		intVal = 1
	}

	return binary.Write(w, binary.LittleEndian, intVal)
}

func ReadSlice(r io.Reader) ([]byte, error) {
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, sizeBuf); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading bytes slice size")
		return nil, err
	}

	size, errInt := binary.Uvarint(sizeBuf)
	if errInt <= 0 {
		log.WithFields(log.Fields{"type": consts.ConversionError, "errInt": errInt}).Error("on convirt sizeBuf to value")
		return nil, fmt.Errorf("wrong sizebuf")
	}

	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading block body")
		return nil, err
	}

	return data, nil
}

func readSliceToBuf(r io.Reader, buf []byte) ([]byte, error) {
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, sizeBuf); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading bytes slice size")
		return nil, err
	}

	size, errInt := binary.Uvarint(sizeBuf)
	if errInt <= 0 {
		log.WithFields(log.Fields{"type": consts.ConversionError, "errInt": errInt}).Error("on convirt sizeBuf to value")
		return nil, fmt.Errorf("wrong sizebuf")
	}

	if cap(buf) < int(size) {
		buf = make([]byte, size)
	}

	_, err := io.ReadFull(r, buf[:size])
	return buf, err
}

func writeSlice(w io.Writer, slice []byte) error {
	byteSize := make([]byte, 4)
	binary.PutUvarint(byteSize, uint64(len(slice)))

	w.Write(byteSize)
	_, err := w.Write(slice)
	return err
}

// if bytesLen < 0 then slice length reads before reading slice body
func readSliceWithSize(r io.Reader, size int) ([]byte, error) {
	slice := make([]byte, size)
	_, err := io.ReadFull(r, slice)
	return slice, err
}

func writeSliceWithSize(w io.Writer, value []byte, size int) error {
	if err := binary.Write(w, binary.LittleEndian, size); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on writing size")
		return err
	}

	_, err := w.Write(value)
	return err
}
func SendRequestType(reqType int64, w io.Writer) error {
	_, err := w.Write(converter.DecToBin(reqType, 2))
	return err
}

func ReadInt(r io.Reader) (int64, error) {
	var value int64
	err := binary.Read(r, binary.LittleEndian, &value)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading integer from network")
		return 0, err
	}

	return value, nil
}

func WriteInt(value int64, w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, value); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending integer to network")
		return err
	}

	return nil
}
