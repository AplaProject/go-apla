package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

type Contract struct {
	*smart.Contract
	req *http.Request
}

type getter interface {
	Get(string) string
}

func (c *Contract) ValidateParams(params getter) (err error) {
	if c.Block.Info.(*script.ContractInfo).Tx == nil {
		return
	}

	logger := getLogger(c.req)

	for _, fitem := range *c.Block.Info.(*script.ContractInfo).Tx {
		if fitem.ContainsTag(script.TagFile) || fitem.ContainsTag(script.TagCrypt) || fitem.ContainsTag(script.TagSignature) {
			continue
		}

		val := strings.TrimSpace(params.Get(fitem.Name))
		if fitem.Type.String() == "[]interface {}" {
			count := params.Get(fitem.Name + "[]")
			if converter.StrToInt(count) > 0 || len(val) > 0 {
				continue
			}
			val = ""
		}
		if len(val) == 0 && !fitem.ContainsTag(script.TagOptional) {
			logger.WithFields(log.Fields{"type": consts.EmptyObject, "item_name": fitem.Name}).Error("route item is empty")
			err = fmt.Errorf("%s is empty", fitem.Name)
			break
		}
		if fitem.ContainsTag(script.TagAddress) {
			addr := converter.StringToAddress(val)
			if addr == 0 {
				logger.WithFields(log.Fields{"type": consts.ConversionError, "value": val}).Error("converting string to address")
				err = fmt.Errorf("Address %s is not valid", val)
				break
			}
		}
		if fitem.Type.String() == script.Decimal {
			re := regexp.MustCompile(`^\d+([\.\,]\d+)?$`)
			if !re.Match([]byte(val)) {
				logger.WithFields(log.Fields{"type": consts.InvalidObject, "value": val}).Error("The value of money is not valid")
				err = fmt.Errorf("The value of money %s is not valid", val)
				break
			}
		}
	}

	return
}

func (c *Contract) ForSign(req *tx.Request, smartTx *tx.SmartContract, params prepareRequestItem) (string, error) {
	contReq := req.NewContract(params.Contract)
	forSignParams, err := c.ForSingParams(contReq, params.Params)
	if err != nil {
		return "", err
	}
	req.AddContract(contReq)

	info := c.Info()
	smartTx.Header.Type = int(info.ID)
	forSign := append([]string{smartTx.ForSign()}, forSignParams...)
	return strings.Join(forSign, ","), nil
}

func (c *Contract) ForSignTx(req *tx.Request, smartTx *tx.SmartContract) []string {
	info := c.Info()
	client := getClient(c.req)
	smartTx.Header = tx.Header{
		Type:        int(info.ID),
		Time:        req.Time.Unix(),
		EcosystemID: client.EcosystemID,
		KeyID:       client.KeyID,
		RoleID:      client.RoleID,
		NetworkID:   consts.NETWORK_ID,
	}
	return []string{smartTx.ForSign()}
}

func (c *Contract) ForSingParams(req *tx.RequestContract, params map[string]string) ([]string, error) {
	var curSize int64

	forSign := []string{}
	limitSize := syspar.GetMaxTxSize()

	info := c.Info()
	for _, fitem := range *info.Tx {
		if fitem.ContainsTag(script.TagSignature) {
			continue
		}

		if fitem.ContainsTag(script.TagFile) {
			fileHeader, err := prepareFormFile(c.req, params[fitem.Name], fitem.Name, req)
			if err != nil {
				return nil, err
			}
			forSign = append(forSign, fileHeader.MimeType, fileHeader.Hash)
			continue
		}

		var val string
		switch fitem.Type.String() {
		case `[]interface {}`:
			for key, values := range params {
				if key == fitem.Name+`[]` && len(values) > 0 {
					count := converter.StrToInt(string(values[0]))
					req.SetParam(key, string(values[0]))
					var list []string
					for i := 0; i < count; i++ {
						k := fmt.Sprintf(`%s[%d]`, fitem.Name, i)
						v := params[k]
						list = append(list, v)
						req.SetParam(k, v)
					}
					val = strings.Join(list, `,`)
				}
			}
			if len(val) == 0 {
				val = params[fitem.Name]
				req.SetParam(fitem.Name, val)
			}

		case script.Decimal:
			d, err := decimal.NewFromString(params[fitem.Name])
			if err != nil {
				getLogger(c.req).WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("converting to decimal")
				return nil, err
			}

			client := getClient(c.req)
			sp := &model.StateParameter{}
			sp.SetTablePrefix(client.Prefix())
			if _, err = sp.Get(nil, model.ParamMoneyDigit); err != nil {
				getLogger(c.req).WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting value from db")
				return nil, err
			}
			exp := int32(converter.StrToInt(sp.Value))

			val = d.Mul(decimal.New(1, exp)).StringFixed(0)
			req.SetParam(fitem.Name, val)

		default:
			val = strings.TrimSpace(params[fitem.Name])
			req.SetParam(fitem.Name, val)
			if strings.Contains(fitem.Tags, `address`) {
				val = converter.Int64ToStr(converter.StringToAddress(val))
			} else if fitem.Type.String() == script.Decimal {
				val = strings.TrimLeft(val, `0`)
			} else if fitem.Type.String() == `int64` && len(val) == 0 {
				val = `0`
			}
		}

		curSize += int64(len(val))
		if curSize > limitSize {
			return nil, errLimitTxSize.Errorf(curSize)
		}

		forSign = append(forSign, val)
	}

	return forSign, nil
}

func (c *Contract) CreateTxFromRequest(contReq *tx.RequestContract, smartTx *tx.SmartContract) (*contractResult, error) {
	smartTx.Data = make([]byte, 0)

	logger := getLogger(c.req)
	client := getClient(c.req)

	info := c.Info()
	if info.Tx != nil {
		var err error
		smartTx.Data, err = packParamsContract(*info.Tx, contReq, logger)
		if err != nil {
			return nil, newError(err, http.StatusBadRequest)
		}
	}

	smartTx.Header.Type = int(info.ID)

	serializedData, err := msgpack.Marshal(smartTx)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
		return nil, err
	}

	if client.IsVDE {
		return callVDEContract(c.req, serializedData)
	}

	hash, err := model.SendTx(int64(info.ID), client.KeyID, append([]byte{128}, serializedData...))
	if err != nil {
		return nil, err
	}

	return &contractResult{
		Hash: hex.EncodeToString(hash),
	}, nil
}

func (c *Contract) Info() *script.ContractInfo {
	return c.Block.Info.(*script.ContractInfo)
}

func (c *Contract) CreateTx(values ...string) error {
	if isVDEMode() {
		return createVDETx(c.req, c.Contract, values...)
	}

	return createTx(c.Contract, values...)
}

func prepareFormFile(r *http.Request, key, reqKey string, req *tx.RequestContract) (*tx.FileHeader, error) {
	logger := getLogger(r)

	file, header, err := r.FormFile(key)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("getting multipart file")
		return nil, newError(err, http.StatusBadRequest)
	}
	defer file.Close()

	fileHeader, err := req.WriteFile(reqKey, header.Header.Get(`Content-Type`), file)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing file")
		return nil, err
	}

	return fileHeader, nil
}

func getContract(r *http.Request, name string) *Contract {
	vm := smart.GetVM()
	if vm == nil {
		return nil
	}

	client := getClient(r)
	contract := smart.VMGetContract(vm, name, uint32(client.EcosystemID))
	if contract == nil {
		return nil
	}

	return &Contract{Contract: contract, req: r}
}

func getContractInfo(contract *smart.Contract) *script.ContractInfo {
	return contract.Block.Info.(*script.ContractInfo)
}

func createTx(contract *smart.Contract, values ...string) error {
	NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) < 1 {
		if err == nil {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		}
		return err
	}

	params := make([]byte, 0)
	for _, v := range values {
		params = append(append(params, converter.EncodeLength(int64(len(v)))...), v...)
	}

	info := contract.Block.Info.(*script.ContractInfo)
	err = tx.BuildTransaction(tx.SmartContract{
		Header: tx.Header{
			Type:        int(info.ID),
			Time:        time.Now().Unix(),
			EcosystemID: 1,
			KeyID:       conf.Config.KeyID,
			NetworkID:   consts.NETWORK_ID,
		},
		SignedBy: smart.PubToID(NodePublicKey),
		Data:     params,
	}, NodePrivateKey, NodePublicKey, values...)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ContractError}).Error("Executing contract")
		return err
	}

	return nil
}

func createVDETx(r *http.Request, contract *smart.Contract, values ...string) error {
	NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) < 1 {
		if err == nil {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		}
		return err
	}

	params := make([]byte, 0)
	for _, v := range values {
		params = append(append(params, converter.EncodeLength(int64(len(v)))...), v...)
	}

	info := contract.Block.Info.(*script.ContractInfo)
	smartTx := tx.SmartContract{
		Header: tx.Header{
			Type:        int(info.ID),
			Time:        time.Now().Unix(),
			EcosystemID: 1,
			KeyID:       conf.Config.KeyID,
			NetworkID:   consts.NETWORK_ID,
		},
		SignedBy: smart.PubToID(NodePublicKey),
		Data:     params,
	}

	signPrms := []string{smartTx.ForSign()}
	signPrms = append(signPrms, values...)
	signature, err := crypto.Sign(
		NodePrivateKey,
		strings.Join(signPrms, ","),
	)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing by node private key")
		return err
	}
	smartTx.BinSignatures = converter.EncodeLengthPlusData(signature)

	if smartTx.PublicKey, err = hex.DecodeString(NodePublicKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex")
		return err
	}

	data, err := msgpack.Marshal(smartTx)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
		return err
	}

	if _, err := callVDEContract(r, data); err != nil {
		return err
	}

	return nil
}

func getTestValue(v string) string {
	return smart.GetTestValue(v)
}

func newTxContract() *tx.SmartContract {
	return &tx.SmartContract{}
}

func newTxHeader() tx.Header {
	return tx.Header{}
}

func packParamsContract(fields []*script.FieldInfo, req *tx.RequestContract, logger *log.Entry) ([]byte, error) {
	idata := []byte{}
	var err error
	for _, fitem := range fields {
		if fitem.ContainsTag(script.TagFile) {
			file, err := req.ReadFile(fitem.Name)
			if err != nil {
				return nil, err
			}

			serialFile, err := msgpack.Marshal(file)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling file to msgpack")
				return nil, err
			}

			idata = append(append(idata, converter.EncodeLength(int64(len(serialFile)))...), serialFile...)
			continue
		}

		val := strings.TrimSpace(req.GetParam(fitem.Name))
		if strings.Contains(fitem.Tags, `address`) {
			val = converter.Int64ToStr(converter.StringToAddress(val))
		}
		switch fitem.Type.String() {
		case `[]interface {}`:
			var list []string
			value := req.GetParam(fitem.Name + `[]`)
			if len(value) > 0 {
				count := converter.StrToInt(value)
				for i := 0; i < count; i++ {
					list = append(list, req.GetParam(fmt.Sprintf(`%s[%d]`, fitem.Name, i)))
				}
			}
			if len(list) == 0 && len(val) > 0 {
				list = append(list, val)
			}
			idata = append(idata, converter.EncodeLength(int64(len(list)))...)
			for _, ilist := range list {
				blist := []byte(ilist)
				idata = append(append(idata, converter.EncodeLength(int64(len(blist)))...), blist...)
			}
		case `uint64`:
			converter.BinMarshal(&idata, converter.StrToUint64(val))
		case `int64`:
			converter.EncodeLenInt64(&idata, converter.StrToInt64(val))
		case `float64`:
			converter.BinMarshal(&idata, converter.StrToFloat64(val))
		case `string`, script.Decimal:
			idata = append(append(idata, converter.EncodeLength(int64(len(val)))...), []byte(val)...)
		case `[]uint8`:
			var bytes []byte
			bytes, err = hex.DecodeString(val)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": val}).Error("decoding value from hex")
				return idata, err
			}
			idata = append(append(idata, converter.EncodeLength(int64(len(bytes)))...), bytes...)
		}
	}
	return idata, nil
}

func isVDEMode() bool {
	return conf.Config.IsSupportingVDE()
}

func callVDEContract(r *http.Request, contractData []byte) (result *contractResult, err error) {
	logger := getLogger(r)

	var ret string
	hash, err := crypto.Hash(contractData)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("getting hash of contract data")
		return nil, err
	}
	result = &contractResult{Hash: hex.EncodeToString(hash)}

	sc := smart.SmartContract{VDE: true, TxHash: hash}
	err = initVDEContract(&sc, contractData)
	if err != nil {
		result.Message = &txstatusError{Type: "panic", Error: err.Error()}
		return
	}

	if token := getToken(r); token != nil && token.Valid {
		if auth, err := token.SignedString([]byte(jwtSecret)); err == nil {
			sc.TxData["auth_token"] = auth
		}
	}

	if ret, err = sc.CallContract(smart.CallInit | smart.CallCondition | smart.CallAction); err == nil {
		result.Result = ret
	} else {
		if errResult := json.Unmarshal([]byte(err.Error()), &result.Message); errResult != nil {
			logger.WithFields(log.Fields{
				"type":  consts.JSONUnmarshallError,
				"text":  err.Error(),
				"error": errResult,
			}).Error("unmarshalling contract error")

			result.Message = &txstatusError{Type: "panic", Error: errResult.Error()}
		}
	}
	return
}

// initVDEContract is initializes smart contract
func initVDEContract(sc *smart.SmartContract, data []byte) error {
	if err := msgpack.Unmarshal(data, &sc.TxSmart); err != nil {
		return err
	}

	sc.TxContract = smart.VMGetContractByID(smart.GetVM(), int32(sc.TxSmart.Type))
	if sc.TxContract == nil {
		return fmt.Errorf(`unknown contract %d`, sc.TxSmart.Type)
	}
	forsign := []string{sc.TxSmart.ForSign()}

	input := sc.TxSmart.Data
	sc.TxData = make(map[string]interface{})

	if sc.TxContract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *sc.TxContract.Block.Info.(*script.ContractInfo).Tx {
			var err error
			var v interface{}
			var forv string
			var isforv bool

			if fitem.ContainsTag(script.TagFile) {
				var (
					data []byte
					file *tx.File
				)
				if err := converter.BinUnmarshal(&input, &data); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling file")
					return err
				}
				if err := msgpack.Unmarshal(data, &file); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("unmarshalling file msgpack")
					return err
				}

				sc.TxData[fitem.Name] = file.Data
				sc.TxData[fitem.Name+"MimeType"] = file.MimeType

				forsign = append(forsign, file.MimeType, file.Hash)
				continue
			}

			switch fitem.Type.String() {
			case `uint64`:
				var val uint64
				converter.BinUnmarshal(&input, &val)
				v = val
			case `float64`:
				var val float64
				converter.BinUnmarshal(&input, &val)
				v = val
			case `int64`:
				v, err = converter.DecodeLenInt64(&input)
			case script.Decimal:
				var s string
				if err := converter.BinUnmarshal(&input, &s); err != nil {
					return err
				}
				v, err = decimal.NewFromString(s)
			case `string`:
				var s string
				if err := converter.BinUnmarshal(&input, &s); err != nil {
					return err
				}
				v = s
			case `[]uint8`:
				var b []byte
				if err := converter.BinUnmarshal(&input, &b); err != nil {
					return err
				}
				v = hex.EncodeToString(b)
			case `[]interface {}`:
				count, err := converter.DecodeLength(&input)
				if err != nil {
					return err
				}
				isforv = true
				list := make([]interface{}, 0)
				for count > 0 {
					length, err := converter.DecodeLength(&input)
					if err != nil {
						return err
					}
					if len(input) < int(length) {
						return fmt.Errorf(`input slice is short`)
					}
					list = append(list, string(input[:length]))
					input = input[length:]
					count--
				}
				if len(list) > 0 {
					slist := make([]string, len(list))
					for j, lval := range list {
						slist[j] = lval.(string)
					}
					forv = strings.Join(slist, `,`)
				}
				v = list
			}
			sc.TxData[fitem.Name] = v
			if err != nil {
				return err
			}
			if strings.Index(fitem.Tags, `image`) >= 0 {
				continue
			}
			if isforv {
				v = forv
			}
			forsign = append(forsign, fmt.Sprintf("%v", v))
		}
	}
	sc.TxData["forsign"] = strings.Join(forsign, ",")
	return nil
}
