package network

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/utils"

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
	t, err := readUint(r, 2)
	if err != nil {
		return err
	}

	rt.Type = uint16(t)
	return nil
}

func (rt *RequestType) Write(w io.Writer) error {
	_, err := w.Write(converter.DecToBin(int64(rt.Type), 2))
	return err
}

// MaxBlockRequest is max block request
type MaxBlockRequest struct{}

// MaxBlockResponse is max block response
type MaxBlockResponse struct {
	BlockID uint32
}

func (resp *MaxBlockResponse) Read(r io.Reader) error {
	t, err := readUint(r, 4)
	if err != nil {
		return err
	}

	resp.BlockID = uint32(t)
	return nil
}

func (resp *MaxBlockResponse) Write(w io.Writer) error {
	_, err := w.Write(converter.DecToBin(int64(resp.BlockID), 4))
	return err
}

// GetBodiesRequest contains BlockID
type GetBodiesRequest struct {
	BlockID      uint32
	ReverseOrder bool
}

func (req *GetBodiesRequest) Read(r io.Reader) error {
	t, err := readUint(r, 4)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading getBodiesRequest blockID")
		return err
	}

	req.BlockID = uint32(t)

	req.ReverseOrder, err = readBool(r)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading GetBodiesRequest reverse order")
	}
	return nil
}

func (req *GetBodiesRequest) Write(w io.Writer) error {
	_, err := w.Write(converter.DecToBin(int64(req.BlockID), 4))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending GetBodiesRequest blockID")
		return err
	}

	err = writeBool(w, req.ReverseOrder)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending GetBodiesRequest reverse order")
		return err
	}

	return err
}

// GetBodyResponse is Data []bytes
type GetBodyResponse struct {
	Data []byte
}

func (resp *GetBodyResponse) Read(r io.Reader) error {
	slice, err := readByteSlice(r, -1)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on reading GetBodyResponse")
		return err
	}

	resp.Data = slice
	return nil
}

func (resp *GetBodyResponse) Write(w io.Writer) error {
	return writeByteSlice(w, resp.Data, -1)
}

// ConfirmRequest contains request data
type ConfirmRequest struct {
	BlockID uint32
}

func (req *ConfirmRequest) Read(r io.Reader) error {
	t, err := readUint(r, 4)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading ConfirmRequest blockID")
		return err
	}

	req.BlockID = uint32(t)
	return nil
}

func (req *ConfirmRequest) Write(w io.Writer) error {
	_, err := w.Write(converter.DecToBin(int64(req.BlockID), 4))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending ConfirmRequest blockID")
		return err
	}

	return nil
}

// ConfirmResponse contains response data
type ConfirmResponse struct {
	// ConfType uint8
	Hash []byte `size:"32"`
}

func (resp *ConfirmResponse) Read(r io.Reader) error {
	h, err := readByteSlice(r, consts.HashSize)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading ConfirmResponse reverse order")
		return err
	}
	resp.Hash = h
	return nil
}

func (resp *ConfirmResponse) Write(w io.Writer) error {
	if err := writeByteSlice(w, resp.Hash, consts.HashSize); err != nil {
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
	slice, err := readByteSlice(r, -1)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on reading disseminator request")
		return err
	}

	req.Data = slice
	return nil
}

func (req *DisRequest) Write(w io.Writer) error {
	err := writeByteSlice(w, req.Data, -1)
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
	slice, err := readByteSlice(r, -1)
	if err != nil {
		return err
	}

	resp.Data = slice
	return nil
}

func (resp *DisHashResponse) Write(w io.Writer) error {
	return writeByteSlice(w, resp.Data, -1)
}

type StopNetworkRequest struct {
	Data []byte
}

func (req *StopNetworkRequest) Read(r io.Reader) error {
	slice, err := readByteSlice(r, -1)
	if err != nil {
		return err
	}

	req.Data = slice
	return nil
}

func (req *StopNetworkRequest) Write(w io.Writer) error {
	return writeByteSlice(w, req.Data, -1)
}

type StopNetworkResponse struct {
	Hash []byte
}

func (resp *StopNetworkResponse) Read(r io.Reader) error {
	slice, err := readByteSlice(r, -1)
	if err != nil {
		return err
	}

	resp.Hash = slice
	return nil
}

func (resp *StopNetworkResponse) Write(w io.Writer) error {
	return writeByteSlice(w, resp.Hash, -1)
}

func readUint(r io.Reader, byteCount int) (uint64, error) {
	buf, err := readBytes(r, uint64(byteCount))
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	return uint64(converter.BinToDec(buf)), nil
}

func readBytes(r io.Reader, size uint64) ([]byte, error) {
	var maxSize uint64 = 10485760
	if size > maxSize {
		log.WithFields(log.Fields{"size": size, "max_size": maxSize, "type": consts.ParameterExceeded}).Error("bytes size to read exceeds max allowed size")
		return nil, errors.New("bad size")
	}
	value := make([]byte, int(size))
	_, err := io.ReadFull(r, value)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError}).Warn("cannot read bytes")
	}
	return value, err
}

func readBool(r io.Reader) (bool, error) {
	boolByte, err := readBytes(r, 1)
	if err != nil {
		return false, err
	}

	return string(boolByte[0]) == "1", nil
}

func writeBool(w io.Writer, val bool) error {
	var bs []byte
	if val {
		bs = []byte("1")
	} else {
		bs = []byte("0")
	}
	_, err := w.Write(bs)
	return err
}

func readSliceSizeFromTag(r io.Reader, tagSize string) (size uint64, err error) {
	if len(tagSize) > 0 {
		size, err = strconv.ParseUint(tagSize, 10, 0)
		if err != nil {
			log.WithFields(log.Fields{"value": tagSize, "type": consts.ConversionError, "error": err}).Error("parsing uint")
		}
		return
	}
	return readUint(r, 4)
}

// if bytesLen < 0 then slice length reads before reading slice body
func readByteSlice(r io.Reader, bytesLen int) ([]byte, error) {
	if bytesLen < 0 {
		size, err := readUint(r, 4)
		if err != nil && err == io.EOF {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Warn("on reading slice size")
			return nil, err
		}
		bytesLen = int(size)
	}

	slice, err := readBytes(r, uint64(bytesLen))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading slice")
		return nil, err
	}

	return slice, nil
}

func writeByteSlice(w io.Writer, value []byte, bytesLen int) error {
	if bytesLen < 0 {
		_, err := w.Write(converter.DecToBin(len(value), 4))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on writing slice size")
			return err
		}
	} else {
		if bytesLen != len(value) {
			log.WithFields(log.Fields{"size": bytesLen, "len": len(value), "type": consts.ProtocolError}).Error("bad slice len")
			return fmt.Errorf("bug, bad slice len, want: %d, got %d", bytesLen, len(value))
		}
	}

	_, err := w.Write(value)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on writing slice data")
		return err
	}

	return nil
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
