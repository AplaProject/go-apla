// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	uuid "github.com/satori/go.uuid"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/theckman/go-flock"
)

// BlockData is a structure of the block's header
type BlockData struct {
	BlockID      int64
	Time         int64
	EcosystemID  int64
	KeyID        int64
	NodePosition int64
	Sign         []byte
	Hash         []byte
	Version      int
}

func (b BlockData) String() string {
	return fmt.Sprintf("BlockID:%d, Time:%d, NodePosition %d", b.BlockID, b.Time, b.NodePosition)
}

var (
	// ReturnCh is chan for returns
	ReturnCh chan string
	// CancelFunc is represents cancel func
	CancelFunc context.CancelFunc
	// DaemonsCount is number of daemons
	DaemonsCount int
)

// GetHTTPTextAnswer returns HTTP answer as a string
func GetHTTPTextAnswer(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError, "url": url}).Error("cannot get url")
		return "", err
	}
	defer resp.Body.Close()
	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError}).Error("cannot read response body")
		return "", err
	}
	if resp.StatusCode == 404 {
		err = fmt.Errorf(`404`)
	}
	return string(htmlData), err
}

// ErrInfoFmt fomats the error message
func ErrInfoFmt(err string, a ...interface{}) error {
	return fmt.Errorf("%s (%s)", fmt.Sprintf(err, a...), Caller(1))
}

// ErrInfo formats the error message
func ErrInfo(verr interface{}, additionally ...string) error {
	var err error
	switch verr.(type) {
	case error:
		err = verr.(error)
	case string:
		err = errors.New(verr.(string))
	}
	if err != nil {
		if len(additionally) > 0 {
			return fmt.Errorf("%s # %s (%s)", err, additionally, Caller(1))
		}
		return fmt.Errorf("%s (%s)", err, Caller(1))
	}
	return err
}

// CallMethod calls the function by its name
func CallMethod(i interface{}, methodName string) interface{} {
	var ptr reflect.Value
	var value reflect.Value
	var finalMethod reflect.Value

	value = reflect.ValueOf(i)

	// if we start with a pointer, we need to get value pointed to
	// if we start with a value, we need to get a pointer to that value
	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(value)
	}

	// check for method on value
	method := value.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	// check for method on pointer
	method = ptr.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}

	if finalMethod.IsValid() {
		return finalMethod.Call([]reflect.Value{})[0].Interface()
	}

	// return or panic, method not found of either type
	log.WithFields(log.Fields{"method_name": methodName, "type": consts.NotFound}).Error("method not found")
	return fmt.Errorf("method %s not found", methodName)
}

// Caller returns the name of the latest function
func Caller(steps int) string {
	name := "?"
	if pc, _, num, ok := runtime.Caller(steps + 1); ok {
		name = fmt.Sprintf("%s :  %d", filepath.Base(runtime.FuncForPC(pc).Name()), num)
	}
	return name
}

// CopyFileContents copy files
func CopyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError, "file_name": src}).Error("opening file")
		return ErrInfo(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError, "file_name": dst}).Error("creating file")
		return ErrInfo(err)
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			log.WithFields(log.Fields{"error": err, "type": consts.IOError, "file_name": dst}).Error("closing file")
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError, "from_file": src, "to_file": dst}).Error("copying from to")
		return ErrInfo(err)
	}
	err = out.Sync()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError, "file_name": dst}).Error("syncing file")
	}
	return ErrInfo(err)
}

// CheckSign checks the signature
func CheckSign(publicKeys [][]byte, forSign string, signs []byte, nodeKeyOrLogin bool) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{"type": consts.PanicRecoveredError, "error": r}).Error("recovered panic in check sign")
		}
	}()

	var signsSlice [][]byte
	if len(forSign) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("for sign is empty")
		return false, ErrInfoFmt("len(forSign) == 0")
	}
	if len(publicKeys) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public keys is empty")
		return false, ErrInfoFmt("len(publicKeys) == 0")
	}
	if len(signs) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signs is empty")
		return false, ErrInfoFmt("len(signs) == 0")
	}

	// node always has olny one signature
	if nodeKeyOrLogin {
		signsSlice = append(signsSlice, signs)
	} else {
		length, err := converter.DecodeLength(&signs)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Fatal("decoding signs length")
			return false, err
		}
		if length > 0 {
			signsSlice = append(signsSlice, converter.BytesShift(&signs, length))
		}
		if len(publicKeys) != len(signsSlice) {
			log.WithFields(log.Fields{"public_keys_length": len(publicKeys), "signs_length": len(signsSlice), "type": consts.SizeDoesNotMatch}).Error("public keys and signs slices lengths does not match")
			return false, fmt.Errorf("sign error %d!=%d", len(publicKeys), len(signsSlice))
		}
	}
	return crypto.CheckSign(publicKeys[0], forSign, signsSlice[0])
}

// MerkleTreeRoot rertun Merkle value
func MerkleTreeRoot(dataArray [][]byte) []byte {
	log.Debug("dataArray: %s", dataArray)
	result := make(map[int32][][]byte)
	for _, v := range dataArray {
		hash, err := crypto.DoubleHash(v)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Fatal("double hasing value, while calculating merkle tree root")
		}
		hash = converter.BinToHex(hash)
		result[0] = append(result[0], hash)
	}
	var j int32
	for len(result[j]) > 1 {
		for i := 0; i < len(result[j]); i = i + 2 {
			if len(result[j]) <= (i + 1) {
				if _, ok := result[j+1]; !ok {
					result[j+1] = [][]byte{result[j][i]}
				} else {
					result[j+1] = append(result[j+1], result[j][i])
				}
			} else {
				if _, ok := result[j+1]; !ok {
					hash, err := crypto.DoubleHash(append(result[j][i], result[j][i+1]...))
					if err != nil {
						log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Fatal("double hasing value, while calculating merkle tree root")
					}
					hash = converter.BinToHex(hash)
					result[j+1] = [][]byte{hash}
				} else {
					hash, err := crypto.DoubleHash([]byte(append(result[j][i], result[j][i+1]...)))
					if err != nil {
						log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Fatal("double hasing value, while calculating merkle tree root")
					}
					hash = converter.BinToHex(hash)
					result[j+1] = append(result[j+1], hash)
				}
			}
		}
		j++
	}

	ret := result[int32(len(result)-1)]
	return []byte(ret[0])
}

// TypeInt returns the identifier of the embedded transaction
func TypeInt(txType string) int64 {
	for k, v := range consts.TxTypes {
		if v == txType {
			return int64(k)
		}
	}
	return 0
}

// GetCurrentDir returns the current directory
func GetCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Warning("getting current dir")
		return "."
	}
	return dir
}

// ShellExecute runs cmdline
func ShellExecute(cmdline string) {
	time.Sleep(500 * time.Millisecond)
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", cmdline).Start()
	case "windows":
		exec.Command(`rundll32.exe`, `url.dll,FileProtocolHandler`, cmdline).Start()
	case "darwin":
		exec.Command("open", cmdline).Start()
	}
}

// GetParent returns the information where the call of function happened
func GetParent() string {
	parent := ""
	for i := 2; ; i++ {
		var name string
		if pc, _, num, ok := runtime.Caller(i); ok {
			name = filepath.Base(runtime.FuncForPC(pc).Name())
			file, line := runtime.FuncForPC(pc).FileLine(pc)
			if i > 5 || name == "runtime.goexit" {
				break
			} else {
				parent += fmt.Sprintf("%s:%d -> %s:%d / ", filepath.Base(file), line, name, num)
			}
		}
	}
	return parent
}

// GetNodeKeys returns node private key and public key
func GetNodeKeys() (string, string, error) {
	nprivkey, err := ioutil.ReadFile(filepath.Join(conf.Config.KeysDir, consts.NodePrivateKeyFilename))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading node private key from file")
		return "", "", err
	}
	key, err := hex.DecodeString(string(nprivkey))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding private key from hex")
		return "", "", err
	}
	npubkey, err := crypto.PrivateToPublic(key)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("converting node private key to public")
		return "", "", err
	}
	return string(nprivkey), hex.EncodeToString(npubkey), nil
}

func GetHostPort(h string) string {
	if strings.Contains(h, ":") {
		return h
	}
	return fmt.Sprintf("%s:%d", h, consts.DEFAULT_TCP_PORT)
}

func BuildBlockTimeCalculator() (BlockTimeCalculator, error) {
	var btc BlockTimeCalculator
	firstBlock := model.Block{}
	found, err := firstBlock.Get(1)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting first block")
		return btc, err
	}

	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("first block not found")
		return btc, err
	}

	blockGenerationDuration := time.Millisecond * time.Duration(syspar.GetMaxBlockGenerationTime())
	blocksGapDuration := time.Second * time.Duration(syspar.GetGapsBetweenBlocks())

	btc = NewBlockTimeCalculator(time.Unix(firstBlock.Time, 0),
		blockGenerationDuration,
		blocksGapDuration,
		syspar.GetNumberOfNodes(),
	)
	return btc, nil
}

func CreateDirIfNotExists(dir string, mode os.FileMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, mode)
		if err != nil {
			return errors.Wrapf(err, "creating dir %s", dir)
		}
	}
	return nil
}

func LockOrDie(dir string) *flock.Flock {
	f := flock.NewFlock(dir)
	success, err := f.TryLock()
	if err != nil {
		log.WithError(err).Fatal("Locking go-genesis")
	}

	if !success {
		log.Fatal("Go-genesis is locked")
	}

	return f
}

func ShuffleSlice(slice []string) {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func UUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

// MakeDirectory makes directory if is not exists
func MakeDirectory(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(dir, 0775)
		}
		return err
	}
	return nil
}
