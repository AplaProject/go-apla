package sql

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"

	"github.com/shopspring/decimal"
)

// InsertInLogTx inserts md5 hash and time into log_transaction
func (db *DCDB) InsertInLogTx(binaryTx []byte, time int64) error {
	txMD5 := utils.Md5(binaryTx)
	err := db.ExecSQL("INSERT INTO log_transactions (hash, time) VALUES ([hex], ?)", txMD5, time)
	log.Debug("INSERT INTO log_transactions (hash, time) VALUES ([hex], %s)", txMD5)
	if err != nil {
		return utils.ErrInfo(err)
	}
	return nil
}

// DelLogTx deletes a row with the specified md5 hash in log_transaction
func (db *DCDB) DelLogTx(binaryTx []byte) error {
	txMD5 := utils.Md5(binaryTx)
	affected, err := db.ExecSQLGetAffect("DELETE FROM log_transactions WHERE hex(hash) = ?", txMD5)
	log.Debug("DELETE FROM log_transactions WHERE hex(hash) = %s / affected = %d", txMD5, affected)
	if err != nil {
		return utils.ErrInfo(err)
	}
	return nil
}

// SendTx writes transaction info to transactions_status & queue_tx
func (db *DCDB) SendTx(txType int64, adminWallet int64, data []byte) (err error) {
	md5 := utils.Md5(data)
	err = db.ExecSQL(`INSERT INTO transactions_status (
			hash, time,	type, wallet_id, citizen_id	) VALUES (
			[hex], ?, ?, ?, ? )`, md5, time.Now().Unix(), txType, adminWallet, adminWallet)
	if err != nil {
		return err
	}
	err = db.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, hex.EncodeToString(data))
	if err != nil {
		return err
	}
	return
}

// GetLastBlockData returns the data of the latest block
func (db *DCDB) GetLastBlockData() (map[string]int64, error) {
	result := make(map[string]int64)
	confirmedBlockID, err := db.GetConfirmedBlockID()
	if err != nil {
		return result, utils.ErrInfo(err)
	}
	if confirmedBlockID == 0 {
		confirmedBlockID = 1
	}
	log.Debug("%v", "confirmedBlockId", confirmedBlockID)
	// получим время из последнего подвержденного блока
	// obtain the time of the last affected block
	lastBlockBin, err := db.Single("SELECT data FROM block_chain WHERE id = ?", confirmedBlockID).Bytes()
	if err != nil || len(lastBlockBin) == 0 {
		return result, utils.ErrInfo(err)
	}
	// ID блока
	result["blockId"] = int64(utils.BinToDec(lastBlockBin[1:5]))
	// Время последнего блока
	// the time of the last block
	result["lastBlockTime"] = int64(utils.BinToDec(lastBlockBin[5:9]))
	return result, nil
}

// GetNodePrivateKey returns the private key from my_nodes_key
func (db *DCDB) GetNodePrivateKey() (string, error) {
	var key string
	key, err := db.Single("SELECT private_key FROM my_node_keys WHERE block_id = (SELECT max(block_id) FROM my_node_keys)").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return key, nil
}

// GetMyStateIDAndWalletID returns state id and wallet id from config
func (db *DCDB) GetMyStateIDAndWalletID() (int64, int64, error) {
	myStateID, err := db.GetMyStateID()
	if err != nil {
		return 0, 0, err
	}
	myWalletID, err := db.GetMyWalletID()
	if err != nil {
		return 0, 0, err
	}
	return myStateID, myWalletID, nil
}

// GetHosts returns the list of hosts
func (db *DCDB) GetHosts() ([]string, error) {
	q := ""
	if db.ConfigIni["db_type"] == "postgresql" {
		q = "SELECT DISTINCT ON (host) host FROM full_nodes"
	} else {
		q = "SELECT host FROM full_nodes GROUP BY host"
	}
	hosts, err := db.GetList(q).String()
	if err != nil {
		return nil, err
	}
	return hosts, nil
}

// CheckDelegateCB checks if the state is delegated
func (db *DCDB) CheckDelegateCB(myStateID int64) (bool, error) {
	delegate, err := db.OneRow("SELECT delegate_wallet_id, delegate_state_id FROM system_recognized_states WHERE state_id = ?", myStateID).Int64()
	if err != nil {
		return false, err
	}
	// Если мы - государство и у нас указан delegate, т.е. мы делегировали полномочия по поддержанию ноды другому юзеру или государству, то выходим.
	// If we are the state and we have the delegate specified (we delegated the authority to maintain the node to another user or state, then we leave).
	if delegate["delegate_wallet_id"] > 0 || delegate["delegate_state_id"] > 0 {
		return true, nil
	}
	return false, nil
}

// GetMyStateID returns state id from config
func (db *DCDB) GetMyStateID() (int64, error) {
	return db.Single("SELECT state_id FROM config").Int64()
}

// GetNodeConfig returns config parameters
func (db *DCDB) GetNodeConfig() (map[string]string, error) {
	return db.OneRow("SELECT * FROM config").String()
}

// GetConfirmedBlockID returns the maximal block id from confirmations
func (db *DCDB) GetConfirmedBlockID() (int64, error) {
	result, err := db.Single("SELECT max(block_id) FROM confirmations WHERE good >= ?", consts.MIN_CONFIRMED_NODES).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil

}

// GetBlockID return the latest block id from info_block
func (db *DCDB) GetBlockID() (int64, error) {
	return db.Single("SELECT block_id FROM info_block").Int64()
}

// GetWalletIDByPublicKey converts public key to wallet id
func (db *DCDB) GetWalletIDByPublicKey(publicKey []byte) (int64, error) {
	key, _ := hex.DecodeString(string(publicKey))
	return int64(lib.Address(key)), nil
}

// GetMyWalletID returns wallet id from config
func (db *DCDB) GetMyWalletID() (int64, error) {
	walletID, err := db.Single("SELECT dlt_wallet_id FROM config").Int64()
	if err != nil {
		return 0, err
	}
	if walletID == 0 {
		//		walletId, err = db.Single("SELECT wallet_id FROM dlt_wallets WHERE address = ?", *WalletAddress).Int64()
		walletID = lib.StringToAddress(*utils.WalletAddress)
	}
	return walletID, nil
}

// GetInfoBlock returns the information about the latest block
func (db *DCDB) GetInfoBlock() (map[string]string, error) {
	var result map[string]string
	result, err := db.OneRow("SELECT * FROM info_block").String()
	if err != nil {
		return result, utils.ErrInfo(err)
	}
	if len(result) == 0 {
		return result, fmt.Errorf("empty info_block")
	}
	return result, nil
}

// GetNodePublicKey returns the node public key of the wallet id
func (db *DCDB) GetNodePublicKey(waletID int64) ([]byte, error) {
	result, err := db.Single("SELECT node_public_key FROM dlt_wallets WHERE wallet_id = ?", waletID).Bytes()
	if err != nil {
		return []byte(""), err
	}
	return result, nil
}

// GetNodePublicKeyWalletOrCB returns node public key of wallet id or state id
func (db *DCDB) GetNodePublicKeyWalletOrCB(walletID, stateID int64) ([]byte, error) {
	var result []byte
	var err error
	if walletID != 0 {
		log.Debug("wallet_id %v state_id %v", walletID, stateID)
		result, err = db.Single("SELECT node_public_key FROM dlt_wallets WHERE wallet_id = ?", walletID).Bytes()
		if err != nil {
			return []byte(""), err
		}
	} else {
		result, err = db.Single("SELECT node_public_key FROM system_recognized_states WHERE state_id = ?", stateID).Bytes()
		if err != nil {
			return []byte(""), err
		}
	}
	return result, nil
}

// GetPublicKeyWalletOrCitizen returns public key of the wallet id or citizen id
func (db *DCDB) GetPublicKeyWalletOrCitizen(walletID, citizenID int64) ([]byte, error) {
	var result []byte
	var err error
	if walletID != 0 {
		result, err = db.Single("SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?", walletID).Bytes()
		if err != nil {
			return []byte(""), err
		}
	} else {
		result, err = db.Single("SELECT public_key_0 FROM ea_citizens WHERE citizen_is = ?", citizenID).Bytes()
		if err != nil {
			return []byte(""), err
		}
	}
	return result, nil
}

// DeleteQueueBlock deletes a row from queue_blocks with the specified hash
func (db *DCDB) DeleteQueueBlock(hashHex string) error {
	return db.ExecSQL("DELETE FROM queue_blocks WHERE hex(hash) = ?", hashHex)
}

// SetAI sets serial sequence for the table
func (db *DCDB) SetAI(table string, AI int64) error {
	AiID, err := db.GetAiID(table)
	if err != nil {
		return utils.ErrInfo(err)
	}

	if db.ConfigIni["db_type"] == "postgresql" {
		pgGetSerialSequence, err := db.Single("SELECT pg_get_serial_sequence('" + table + "', '" + AiID + "')").String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = db.ExecSQL("ALTER SEQUENCE " + pgGetSerialSequence + " RESTART WITH " + utils.Int64ToStr(AI))
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

// GetAiID returns auto increment column
func (db *DCDB) GetAiID(table string) (string, error) {
	exists := ""
	column := "id"
	if table == "users" {
		column = "user_id"
	} else if table == "miners" {
		column = "miner_id"
	} else {
		switch db.ConfigIni["db_type"] {
		case "postgresql":
			exists = ""
			err := db.QueryRow("SELECT column_name FROM information_schema.columns WHERE table_name=$1 and column_name=$2", table, "id").Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				return "", err
			}
			if len(exists) == 0 {
				err := db.QueryRow("SELECT column_name FROM information_schema.columns WHERE table_name=$1 and column_name=$2", table, "rb_id").Scan(&exists)
				if err != nil {
					return "", err
				}
				if len(exists) == 0 {
					return "", fmt.Errorf("no id, rb_id")
				}
				column = "rb_id"
			}
		}
	}
	return column, nil
}

// NodesBan is reserved
func (db *DCDB) NodesBan(info string) error {

	return nil
}

// GetBlockDataFromBlockChain returns the block information from the blockchain
func (db *DCDB) GetBlockDataFromBlockChain(blockID int64) (*utils.BlockData, error) {
	BlockData := new(utils.BlockData)
	data, err := db.OneRow("SELECT * FROM block_chain WHERE id = ?", blockID).String()
	if err != nil {
		return BlockData, utils.ErrInfo(err)
	}
	log.Debug("data: %x\n", data["data"])
	if len(data["data"]) > 0 {
		binaryData := []byte(data["data"])
		utils.BytesShift(&binaryData, 1) // не нужно. 0 - блок, >0 - тр-ии
		BlockData = utils.ParseBlockHeader(&binaryData)
		BlockData.Hash = utils.BinToHex([]byte(data["hash"]))
	}
	return BlockData, nil
}

// GetTxTypeAndUserID returns tx type, wallet and citizen id from the block data
func GetTxTypeAndUserID(binaryBlock []byte) (txType int64, walletID int64, citizenID int64) {
	tmp := binaryBlock[:]
	txType = utils.BinToDecBytesShift(&binaryBlock, 1)
	if consts.IsStruct(int(txType)) {
		var txHead consts.TxHeader
		lib.BinUnmarshal(&tmp, &txHead)
		walletID = txHead.WalletID
		citizenID = txHead.CitizenID
	} else if txType > 127 {
		header := consts.TXHeader{}
		err := lib.BinUnmarshal(&tmp, &header)
		if err == nil {
			if header.StateID > 0 {
				citizenID = int64(header.WalletID)
			} else {
				walletID = int64(header.WalletID)
			}
		}
	} else {
		utils.BytesShift(&binaryBlock, 4) // уберем время
		walletID = utils.BytesToInt64(utils.BytesShift(&binaryBlock, utils.DecodeLength(&binaryBlock)))
		citizenID = utils.BytesToInt64(utils.BytesShift(&binaryBlock, utils.DecodeLength(&binaryBlock)))
	}
	return
}

// DecryptData decrypts tx data
func (db *DCDB) DecryptData(binaryTx *[]byte) ([]byte, []byte, []byte, error) {
	if len(*binaryTx) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(binaryTx) == 0")
	}
	// вначале пишется user_id, чтобы в режиме пула можно было понять, кому шлется и чей ключ использовать
	// at the beginning the user ID is written to know in the pool mode to whom it is sent and what key to use
	myUserID := utils.BinToDecBytesShift(&*binaryTx, 5)
	log.Debug("myUserId: %d", myUserID)

	// изымем зашифрванный ключ, а всё, что останется в $binary_tx - сами зашифрованные хэши тр-ий/блоков
	// remove the encrypted key, and all that stay in $binary_tx will be encrypted keys of the transactions/blocks
	encryptedKey := utils.BytesShift(&*binaryTx, utils.DecodeLength(&*binaryTx))
	log.Debug("encryptedKey: %x", encryptedKey)
	log.Debug("encryptedKey: %s", encryptedKey)

	// далее идет 16 байт IV
	// 16 bytes IV go further
	iv := utils.BytesShift(&*binaryTx, 16)
	log.Debug("iv: %s", iv)
	log.Debug("iv: %x", iv)

	if len(encryptedKey) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(encryptedKey) == 0")
	}

	if len(*binaryTx) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(*binaryTx) == 0")
	}

	nodePrivateKey, err := db.GetNodePrivateKey()
	if len(nodePrivateKey) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(nodePrivateKey) == 0")
	}

	block, _ := pem.Decode([]byte(nodePrivateKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, nil, nil, utils.ErrInfo("No valid PEM data found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}

	decKey, err := rsa.DecryptPKCS1v15(crand.Reader, privateKey, encryptedKey)
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}
	log.Debug("decrypted Key: %s", decKey)
	if len(decKey) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(decKey)")
	}

	log.Debug("binaryTx %x", *binaryTx)
	log.Debug("iv %s", iv)
	decrypted, err := utils.DecryptCFB(iv, *binaryTx, decKey)
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}

	return decKey, iv, decrypted, nil
}

// FindInFullNodes returns id of the node
func (db *DCDB) FindInFullNodes(myStateID, myWalletID int64) (int64, error) {
	return db.Single("SELECT id FROM full_nodes WHERE final_delegate_state_id = ? OR final_delegate_wallet_id = ? OR state_id = ? OR wallet_id = ?",
		myStateID, myWalletID, myStateID, myWalletID).Int64()
}

// GetBinSign returns a signature made with node private key
func (db *DCDB) GetBinSign(forSign string) ([]byte, error) {
	nodePrivateKey, err := db.GetNodePrivateKey()
	if err != nil {
		return nil, utils.ErrInfo(err)
	}
	return lib.SignECDSA(nodePrivateKey, forSign)
}

// InsertReplaceTxInQueue replaces a row in queue_tx
func (db *DCDB) InsertReplaceTxInQueue(data []byte) error {
	log.Debug("DELETE FROM queue_tx WHERE hex(hash) = %s", utils.Md5(data))
	err := db.ExecSQL("DELETE FROM queue_tx WHERE hex(hash) = ?", utils.Md5(data))
	if err != nil {
		return utils.ErrInfo(err)
	}
	log.Debug("INSERT INTO queue_tx (hash, data) VALUES (%s, %s)", utils.Md5(data), utils.BinToHex(data))
	err = db.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", utils.Md5(data), utils.BinToHex(data))
	if err != nil {
		return utils.ErrInfo(err)
	}
	return nil
}

// GetSleepTime returns the waiting time for wallet id and state id
func (db *DCDB) GetSleepTime(myWalletID, myStateID, prevBlockStateID, prevBlockWalletID int64) (int64, error) {
	// возьмем список всех full_nodes
	// take the list of all full_nodes
	fullNodesList, err := db.GetAll("SELECT id, wallet_id, state_id as state_id FROM full_nodes", -1)
	if err != nil {
		return int64(0), utils.ErrInfo(err)
	}
	log.Debug("fullNodesList %s", fullNodesList)

	// определим full_node_id того, кто должен был генерить блок (но мог это делегировать)
	// determine full_node_id of the one, who had to generate a block (but could delegate this)
	prevBlockFullNodeID, err := db.Single("SELECT id FROM full_nodes WHERE state_id = ? OR wallet_id = ?", prevBlockStateID, prevBlockWalletID).Int64()
	if err != nil {
		return int64(0), utils.ErrInfo(err)
	}
	log.Debug("prevBlockFullNodeId %d", prevBlockFullNodeID)
	log.Debug("%v %v", fullNodesList, prevBlockFullNodeID)
	prevBlockFullNodePosition := func(fullNodesList []map[string]string, prevBlockFullNodeID int64) int {
		for i, fullNodes := range fullNodesList {
			if utils.StrToInt64(fullNodes["id"]) == prevBlockFullNodeID {
				return i
			}
		}
		return -1
	}(fullNodesList, prevBlockFullNodeID)
	log.Debug("prevBlockFullNodePosition %d", prevBlockFullNodePosition)

	// определим свое место (в том числе в delegate)
	// define our place (Including in the 'delegate')
	myPosition := func(fullNodesList []map[string]string, myWalletID, myStateID int64) int {
		log.Debug("%v %v", fullNodesList, myWalletID)
		for i, fullNodes := range fullNodesList {
			if utils.StrToInt64(fullNodes["state_id"]) == myStateID || utils.StrToInt64(fullNodes["wallet_id"]) == myWalletID ||
				utils.StrToInt64(fullNodes["final_delegate_state_id"]) == myWalletID || utils.StrToInt64(fullNodes["final_delegate_wallet_id"]) == myWalletID {
				return i
			}
		}
		return -1
	}(fullNodesList, myWalletID, myStateID)
	log.Debug("myPosition %d", myPosition)

	sleepTime := 0
	if myPosition == prevBlockFullNodePosition {
		sleepTime = ((len(fullNodesList) + myPosition) - int(prevBlockFullNodePosition)) * consts.GAPS_BETWEEN_BLOCKS
	}

	if myPosition > prevBlockFullNodePosition {
		sleepTime = (myPosition - int(prevBlockFullNodePosition)) * consts.GAPS_BETWEEN_BLOCKS
	}

	if myPosition < prevBlockFullNodePosition {
		sleepTime = (len(fullNodesList) - prevBlockFullNodePosition) * consts.GAPS_BETWEEN_BLOCKS
	}
	log.Debug("sleepTime %v / myPosition %v / prevBlockFullNodePosition %v / consts.GAPS_BETWEEN_BLOCKS %v", sleepTime, myPosition, prevBlockFullNodePosition, consts.GAPS_BETWEEN_BLOCKS)

	return int64(sleepTime), nil
}

//GetStateName returns the name of the state
func (db *DCDB) GetStateName(stateID int64) (string, error) {
	var err error
	sID, err := db.Single(`SELECT id FROM system_states WHERE id = ?`, stateID).String()
	if err != nil {
		return ``, err
	}
	stateName := ""
	if sID != "0" {
		stateName, err = db.Single(`SELECT value FROM "` + sID + `_state_parameters" WHERE name = 'state_name'`).String()
		if err != nil {
			return ``, err
		}
	}
	return stateName, nil
}

// CheckStateName checks if the state id is valid
func (db *DCDB) CheckStateName(stateID int64) (bool, error) {
	stateID, err := db.Single(`SELECT id FROM system_states WHERE id = ?`, stateID).Int64()
	if err != nil {
		return false, err
	}
	if stateID > 0 {
		return true, nil
	}
	return false, fmt.Errorf("null stateId")
}

// GetFuel returns the fuel rate
func (db *DCDB) GetFuel() decimal.Decimal {
	// fuel = qEGS/F
	/*	fuelMutex.Lock()
		defer fuelMutex.Unlock()
		if cacheFuel <= 0 {*/
	fuel, _ := db.Single(`SELECT value FROM system_parameters WHERE name = ?`, "fuel_rate").String()
	//}
	cacheFuel, _ := decimal.NewFromString(fuel)
	return cacheFuel
}

// IsNodeState checks if the state is specified as node_stat_id in config file
func (db *DCDB) IsNodeState(state int64, host string) bool {
	if strings.HasPrefix(host, `localhost`) {
		return true
	}
	if val, ok := db.ConfigIni[`node_state_id`]; ok {
		if val == `*` {
			return true
		}
		for _, id := range strings.Split(val, `,`) {
			if utils.StrToInt64(id) == state {
				return true
			}
		}
	}
	return false
}

// IsState returns the identifier of the state
func (db *DCDB) IsState(country string) (int64, error) {
	data, err := db.GetList(`SELECT id FROM system_states`).Int64()
	if err != nil {
		return 0, err
	}
	for _, id := range data {
		stateName, err := db.Single(fmt.Sprintf(`SELECT value FROM "%d_state_parameters" WHERE name = 'state_name'`, id)).String()
		if err != nil {
			return 0, err
		}
		if strings.ToLower(stateName) == strings.ToLower(country) {
			return id, nil
		}
	}
	return 0, nil
}

// UpdateFuel is reserved
func (db *DCDB) UpdateFuel() {
	/*	fuelMutex.Lock()
		cacheFuel, _ = db.Single(`SELECT value FROM system_parameters WHERE name = ?`, "fuel_rate").Int64()
		fuelMutex.Unlock()*/
}
