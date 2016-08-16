package parser

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

var (
	log = logging.MustGetLogger("daemons")
)

func init() {
	flag.Parse()
}


type txMapsType struct {
	Int64   map[string]int64
	String  map[string]string
	Bytes   map[string][]byte
	Float64 map[string]float64
	Money   map[string]float64
}
type Parser struct {
	*utils.DCDB
	TxMaps           *txMapsType
	TxMap            map[string][]byte
	TxMapS           map[string]string
	TxIds            []string
	TxMapArr         []map[string][]byte
	TxMapsArr        []*txMapsType
	BlockData        *utils.BlockData
	PrevBlock        *utils.BlockData
	BinaryData       []byte
	blockHashHex     []byte
	dataType         int
	blockHex         []byte
	CurrentBlockId   int64
	fullTxBinaryData []byte
	halfRollback     bool // уже не актуально, т.к. нет ни одной половинной фронт-проверки
	TxHash           string
	TxSlice          [][]byte
	MerkleRoot       []byte
	GoroutineName    string
	CurrentVersion   string
	MrklRoot         []byte
	PublicKeys       [][]byte
	AdminUserId      int64
	TxUserID         int64
	TxCitizenID         int64
	TxWalletID         int64
	TxTime           int64
	nodePublicKey    []byte
	newPublicKeysHex [3][]byte
}

type MinerData struct {
	adminUserId     int64
	myMinersIds      map[int]int
	minersIds        map[int]int
	votes0           int64
	votes1           int64
	minMinersKeepers int64
}


func ClearTmp(blocks map[int64]string) {
	for _, tmpFileName := range blocks {
		os.Remove(tmpFileName)
	}
}

/*
 * $get_block_script_name, $add_node_host используется только при работе в защищенном режиме и только из blocks_collection.php
 * */
func (p *Parser) GetOldBlocks(walletId,CBID, blockId int64, host string, goroutineName string, dataTypeBlockBody int64) error {
	log.Debug("walletId", walletId,"CBID", CBID, "blockId", blockId)
	err := p.GetBlocks(blockId, host, "rollback_blocks_2", goroutineName, dataTypeBlockBody)
	if err != nil {
		log.Error("v", err)
		return err
	}
	return nil
}

func (p *Parser) GetBlocks(blockId int64, host string, rollbackBlocks, goroutineName string, dataTypeBlockBody int64) error {

	log.Debug("blockId", blockId)

	parser := new(Parser)
	parser.DCDB = p.DCDB
	var count int64
	blocks := make(map[int64]string)
	for {
		/*
			// отметимся в БД, что мы живы.
			upd_deamon_time($db);
			// отметимся, чтобы не спровоцировать очистку таблиц
			upd_main_lock($db);
			// проверим, не нужно нам выйти, т.к. обновилась версия скрипта
			if (check_deamon_restart($db)){
				main_unlock();
				exit;
			}*/
		if blockId < 2 {
			return utils.ErrInfo(errors.New("block_id < 2"))
		}
		// если превысили лимит кол-ва полученных от нода блоков
		var rollback = consts.RB_BLOCKS_1
		if rollbackBlocks == "rollback_blocks_2" {
			rollback = consts.RB_BLOCKS_2
		}
		if count > int64(rollback) {
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New("count > variables[rollback_blocks]"))
		}

		// качаем тело блока с хоста host
		binaryBlock, err := utils.GetBlockBody(host, blockId, dataTypeBlockBody)

		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}
		log.Debug("binaryBlock: %x\n", binaryBlock)
		binaryBlockFull := binaryBlock
		if len(binaryBlock) == 0 {
			log.Debug("len(binaryBlock) == 0")
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New("len(binaryBlock) == 0"))
		}
		utils.BytesShift(&binaryBlock, 1) // уберем 1-й байт - тип (блок/тр-я)
		// распарсим заголовок блока
		blockData := utils.ParseBlockHeader(&binaryBlock)
		log.Debug("blockData", blockData)

		// если существуют глючная цепочка, тот тут мы её проигнорируем
		badBlocks_, err := p.Single("SELECT bad_blocks FROM config").Bytes()
		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}
		badBlocks := make(map[int64]string)
		if len(badBlocks_) > 0 {
			err = json.Unmarshal(badBlocks_, &badBlocks)
			if err != nil {
				ClearTmp(blocks)
				return utils.ErrInfo(err)
			}
		}
		if badBlocks[blockData.BlockId] == string(utils.BinToHex(blockData.Sign)) {
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New("bad block"))
		}
		if blockData.BlockId != blockId {
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New("bad block_data['block_id']"))
		}

		// размер блока не может быть более чем max_block_size
		if int64(len(binaryBlock)) > consts.MAX_BLOCK_SIZE {
			ClearTmp(blocks)
			return utils.ErrInfo(errors.New(`len(binaryBlock) > variables.Int64["max_block_size"]`))
		}

		// нам нужен хэш предыдущего блока, чтобы найти, где началась вилка
		prevBlockHash, err := p.Single("SELECT hash FROM block_chain WHERE id  =  ?", blockId-1).String()
		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}

		// нам нужен меркель-рут текущего блока
		mrklRoot, err := utils.GetMrklroot(binaryBlock, false)
		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}

		// публичный ключ того, кто этот блок сгенерил
		nodePublicKey, err := p.GetNodePublicKeyWalletOrCB(blockData.WalletId, blockData.CBID)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// SIGN от 128 байта до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, CB_ID, MRKL_ROOT
		forSign := fmt.Sprintf("0,%v,%x,%v,%v,%v,%s", blockData.BlockId, prevBlockHash, blockData.Time, blockData.WalletId, blockData.CBID, mrklRoot)
		log.Debug("forSign", forSign)

		// проверяем подпись
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey}, forSign, blockData.Sign, true)
		log.Debug("okSignErr", okSignErr)

		// сам блок сохраняем в файл, чтобы не нагружать память
		file, err := ioutil.TempFile(*utils.Dir, "DC")
		defer os.Remove(file.Name())
		_, err = file.Write(binaryBlockFull)
		if err != nil {
			ClearTmp(blocks)
			return utils.ErrInfo(err)
		}
		blocks[blockId] = file.Name()
		blockId--
		count++

		// качаем предыдущие блоки до тех пор, пока отличается хэш предыдущего.
		// другими словами, пока подпись с prevBlockHash будет неверной, т.е. пока что-то есть в okSignErr
		if okSignErr == nil {
			log.Debug("plug found blockId=%v\n", blockData.BlockId)
			break
		}
	}

	// чтобы брать блоки по порядку
	blocksSorted := utils.SortMap(blocks)
	log.Debug("blocks", blocksSorted)

	// получим наши транзакции в 1 бинарнике, просто для удобства
	var transactions []byte
	utils.WriteSelectiveLog(`SELECT data FROM transactions WHERE verified = 1 AND used = 0`)
	all, err := p.GetAll(`SELECT data FROM transactions WHERE verified = 1 AND used = 0`, -1)
	if err != nil {
		utils.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	for _, data := range all {
		utils.WriteSelectiveLog(utils.BinToHex(data["data"]))
		log.Debug("data", data)
		transactions = append(transactions, utils.EncodeLengthPlusData([]byte(data["data"]))...)
	}
	if len(transactions) > 0 {
		// отмечаем, что эти тр-ии теперь нужно проверять по новой
		utils.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
		affect, err := p.ExecSqlGetAffect("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
		// откатываем по фронту все свежие тр-ии
		parser.GoroutineName = goroutineName
		parser.BinaryData = transactions
		err = parser.ParseDataRollbackFront(false)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}


	// откатываем наши блоки до начала вилки
	rows, err := p.Query(p.FormatQuery(`
			SELECT data
			FROM block_chain
			WHERE id > ?
			ORDER BY id DESC`), blockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var data []byte
		err = rows.Scan(&data)
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("We roll away blocks before plug", blockId)
		parser.GoroutineName = goroutineName
		parser.BinaryData = data
		err = parser.ParseDataRollback()
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	log.Debug("blocks", blocksSorted)

	prevBlock := make(map[int64]*utils.BlockData)

	// проходимся по новым блокам
	for _, data := range blocksSorted {
		for intBlockId, tmpFileName := range data {
			log.Debug("Go on new blocks", intBlockId, tmpFileName)

			// проверяем и заносим данные
			binaryBlock, err := ioutil.ReadFile(tmpFileName)
			if err != nil {
				return utils.ErrInfo(err)
			}
			log.Debug("binaryBlock: %x\n", binaryBlock)
			parser.GoroutineName = goroutineName
			parser.BinaryData = binaryBlock
			// передаем инфу о предыдущем блоке, т.к. это новые блоки, то инфа о предыдущих блоках в block_chain будет всё еще старая, т.к. обновление block_chain идет ниже
			if prevBlock[intBlockId-1] != nil {
				log.Debug("prevBlock[intBlockId-1] != nil : %v", prevBlock[intBlockId-1])
				parser.PrevBlock.Hash = prevBlock[intBlockId-1].Hash
				parser.PrevBlock.Time = prevBlock[intBlockId-1].Time
				parser.PrevBlock.BlockId = prevBlock[intBlockId-1].BlockId
			}

			// если вернулась ошибка, значит переданный блок уже откатился
			// info_block и config.my_block_id обновляются только если ошибки не было
			err = parser.ParseDataFull()
			// для последующей обработки получим хэши и time
			if err == nil {
				prevBlock[intBlockId] = parser.GetBlockInfo()
				log.Debug("prevBlock[%d] = %v", intBlockId, prevBlock[intBlockId])
			}
			// если есть ошибка, то откатываем все предыдущие блоки из новой цепочки
			if err != nil {
				log.Debug("there is an error is rolled back all previous blocks of a new chain: %v", err)

				// баним на 1 час хост, который дал нам ложную цепочку
				err = p.NodesBan(fmt.Sprintf("%s", err))
				if err != nil {
					return utils.ErrInfo(err)
				}
				// обязательно проходимся по блокам в обратном порядке
				blocksSorted := utils.RSortMap(blocks)
				for _, data := range blocksSorted {
					for int2BlockId, tmpFileName := range data {
						log.Debug("int2BlockId", int2BlockId)
						if int2BlockId >= intBlockId {
							continue
						}
						binaryBlock, err := ioutil.ReadFile(tmpFileName)
						if err != nil {
							return utils.ErrInfo(err)
						}
						parser.GoroutineName = goroutineName
						parser.BinaryData = binaryBlock
						err = parser.ParseDataRollback()
						if err != nil {
							return utils.ErrInfo(err)
						}
					}
				}
				// заносим наши данные из block_chain, которые были ранее
				log.Debug("We push data from our block_chain, which were previously")
				rows, err := p.Query(p.FormatQuery(`
					SELECT data
					FROM block_chain
					WHERE id > ?
					ORDER BY id ASC`), blockId)
				if err != nil {
					return p.ErrInfo(err)
				}
				defer rows.Close()
				for rows.Next() {
					var data []byte
					err = rows.Scan(&data)
					if err != nil {
						return p.ErrInfo(err)
					}
					log.Debug("blockId", blockId, "intBlockId", intBlockId)
					parser.GoroutineName = goroutineName
					parser.BinaryData = data
					err = parser.ParseDataFull()
					if err != nil {
						return utils.ErrInfo(err)
					}
				}
				// т.к. в предыдущем запросе к block_chain могло не быть данных, т.к. $block_id больше чем наш самый большой id в block_chain
				// то значит info_block мог не обновится и остаться от занесения новых блоков, что приведет к пропуску блока в block_chain
				lastMyBlock, err := p.OneRow("SELECT * FROM block_chain ORDER BY id DESC").String()
				if err != nil {
					return utils.ErrInfo(err)
				}
				binary := []byte(lastMyBlock["data"])
				utils.BytesShift(&binary, 1) // уберем 1-й байт - тип (блок/тр-я)
				lastMyBlockData := utils.ParseBlockHeader(&binary)
				err = p.ExecSql(`
					UPDATE info_block
					SET   hash = [hex],
							block_id = ?,
							time = ?,
							sent = 0
					`, utils.BinToHex(lastMyBlock["hash"]), lastMyBlockData.BlockId, lastMyBlockData.Time)
				if err != nil {
					return utils.ErrInfo(err)
				}
				err = p.ExecSql(`UPDATE config SET my_block_id = ?`, lastMyBlockData.BlockId)
				if err != nil {
					return utils.ErrInfo(err)
				}
				ClearTmp(blocks)
				return utils.ErrInfo(err) // переходим к следующему блоку в queue_blocks
			}
		}
	}
	log.Debug("remove the blocks and enter new block_chain")

	// если всё занеслось без ошибок, то удаляем блоки из block_chain и заносим новые
	affect, err := p.ExecSqlGetAffect("DELETE FROM block_chain WHERE id > ?", blockId)
	if err != nil {
		return utils.ErrInfo(err)
	}
	log.Debug("affect", affect)
	log.Debug("prevblock", prevBlock)
	log.Debug("blocks", blocks)

	// для поиска бага
	maxBlockId, err := p.Single("SELECT id FROM block_chain ORDER BY id DESC LIMIT 1").Int64()
	if err != nil {
		return utils.ErrInfo(err)
	}
	log.Debug("maxBlockId", maxBlockId)

	// проходимся по новым блокам
	for blockId, tmpFileName := range blocks {

		block, err := ioutil.ReadFile(tmpFileName)
		if err != nil {
			return utils.ErrInfo(err)
		}
		blockHex := utils.BinToHex(block)

		// пишем в цепочку блоков
		err = p.ExecSql("UPDATE info_block SET hash = [hex], block_id = ?, time = ?, wallet_id = ?, cb_id = ?, sent = 0", prevBlock[blockId].Hash, prevBlock[blockId].BlockId, prevBlock[blockId].Time, prevBlock[blockId].WalletId, prevBlock[blockId].CBID)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = p.ExecSql(`UPDATE config SET my_block_id = ?`, prevBlock[blockId].BlockId)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// т.к. эти данные создали мы сами, то пишем их сразу в таблицу проверенных данных, которые будут отправлены другим нодам
		exists, err := p.Single("SELECT id FROM block_chain WHERE id = ?", blockId).Int64()
		if err != nil {
			return utils.ErrInfo(err)
		}
		if exists == 0 {
			affect, err := p.ExecSqlGetAffect("INSERT INTO  block_chain (id, hash, data) VALUES (?, [hex], [hex])", blockId, prevBlock[blockId].Hash, blockHex)
			if err != nil {
				return utils.ErrInfo(err)
			}
			log.Debug("affect", affect)
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("tmpFileName %v", tmpFileName)
		// для поиска бага
		maxBlockId, err := p.Single("SELECT id FROM block_chain ORDER BY id DESC LIMIT 1").Int64()
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("maxBlockId", maxBlockId)
	}

	log.Debug("HAPPY END")

	return nil
}

func (p *Parser) GetBlockInfo() *utils.BlockData {
	return &utils.BlockData{Hash: p.BlockData.Hash, Time: p.BlockData.Time,  WalletId: p.BlockData.WalletId,  CBID: p.BlockData.CBID, BlockId: p.BlockData.BlockId}
}

func (p *Parser) limitRequest(limit_ interface{}, txType string, period_ interface{}) error {

	var limit int
	switch limit_.(type) {
	case string:
		limit = utils.StrToInt(limit_.(string))
	case int:
		limit = limit_.(int)
	case int64:
		limit = int(limit_.(int64))
	}

	var period int
	switch period_.(type) {
	case string:
		period = utils.StrToInt(period_.(string))
	case int:
		period = period_.(int)
	}

	time := utils.BytesToInt(p.TxMap["time"])
	num, err := p.Single("SELECT count(time) FROM rb_time_"+txType+" WHERE user_id = ? AND time > ?", p.TxUserID, (time - period)).Int()
	if err != nil {
		return err
	}
	if num >= limit {
		return utils.ErrInfo(fmt.Errorf("[limit_requests] rb_time_%v %v >= %v", txType, num, limit))
	} else {
		err := p.ExecSql("INSERT INTO rb_time_"+txType+" (user_id, time) VALUES (?, ?)", p.TxUserID, time)
		if err != nil {
			return err
		}
	}
	return nil
}


// общая проверка для всех _front
func (p *Parser) generalCheck() error {
	log.Debug("%s", p.TxMap)
	if !utils.CheckInputData(p.TxMap["wallet_id"], "int64") {
		return utils.ErrInfoFmt("incorrect wallet_id")
	}
	if !utils.CheckInputData(p.TxMap["citizen_id"], "int64") {
		return utils.ErrInfoFmt("incorrect citizen_id")
	}
	if !utils.CheckInputData(p.TxMap["time"], "int") {
		return utils.ErrInfoFmt("incorrect time")
	}

	// проверим, есть ли такой юзер и заодно получим public_key
	if p.TxMaps.Int64["type"] == utils.TypeInt("DLTTransfer") || p.TxMaps.Int64["type"] == utils.TypeInt("DLTChangeHostVote") || p.TxMaps.Int64["type"] == utils.TypeInt("DLTCitizenRequest") {
			data, err := p.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM dlt_wallets WHERE wallet_id = ?", utils.BytesToInt64(p.TxMap["wallet_id"])).String()
			if err != nil {
				return utils.ErrInfo(err)
			}
			log.Debug("datausers", data)
			if len(data["public_key_0"]) == 0 {
				if len(p.TxMap["public_key"]) == 0 {
					return utils.ErrInfoFmt("incorrect public_key")
				}
				// возможно юзер послал ключ с тр-ией
				log.Debug("lower(hex(address) %s", string(utils.HashSha1Hex([]byte(p.TxMap["public_key"]))))
				walletId, err := p.Single(`SELECT wallet_id FROM dlt_wallets WHERE address = [hex]`, string(utils.HashSha1Hex([]byte(p.TxMap["public_key"])))).Int64()
				if err != nil {
					return utils.ErrInfo(err)
				}
				if walletId == 0 {
					return utils.ErrInfoFmt("incorrect wallet_id or public_key")
				}
				p.PublicKeys = append(p.PublicKeys, []byte(data["public_key"]))
			}
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_0"]))
			if len(data["public_key_1"]) > 10 {
				p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_1"]))
			}
			if len(data["public_key_2"]) > 10 {
				p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_2"]))
			}
	} else {
		data, err := p.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM citizens WHERE citizen_id = ?", utils.BytesToInt64(p.TxMap["citizen_id"])).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("datausers", data)
		if len(data["public_key_0"]) == 0 {
			return utils.ErrInfoFmt("incorrect user_id")
		}
		p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_0"]))
		if len(data["public_key_1"]) > 10 {
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_1"]))
		}
		if len(data["public_key_2"]) > 10 {
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_2"]))
		}
	}
	// чтобы не записали слишком длинную подпись
	// 128 - это нод-ключ
	if len(p.TxMap["sign"]) < 128 || len(p.TxMap["sign"]) > 5120 {
		return utils.ErrInfoFmt("incorrect sign size %d", len(p.TxMap["sign"]))
	}
	return nil
}

func (p *Parser) dataPre() {
	p.blockHashHex = utils.DSha256(p.BinaryData)
	p.blockHex = utils.BinToHex(p.BinaryData)
	// определим тип данных
	p.dataType = int(utils.BinToDec(utils.BytesShift(&p.BinaryData, 1)))
	log.Debug("dataType", p.dataType)
}

func (p *Parser) ParseBlock() error {
	/*
		Заголовок
		TYPE (0-блок, 1-тр-я)     1
		BLOCK_ID   				       4
		TIME       					       4
		WALLET_ID                         1-8
		CB_ID                         1
		SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, CB_ID, MRKL_ROOT
		Далее - тело блока (Тр-ии)
	*/
	p.BlockData = utils.ParseBlockHeader(&p.BinaryData)
	log.Debug("%v", p.BlockData)

	p.CurrentBlockId = p.BlockData.BlockId

	return nil
}

func (p *Parser) CheckBlockHeader() error {
	var err error
	// инфа о предыдущем блоке (т.е. последнем занесенном).
	// в GetBlocks p.PrevBlock определяется снаружи, поэтому тут важно не перезаписать данными из block_chain
	if p.PrevBlock == nil || p.PrevBlock.BlockId != p.BlockData.BlockId-1 {
		p.PrevBlock, err = p.GetBlockDataFromBlockChain(p.BlockData.BlockId - 1)
		log.Debug("PrevBlock 0", p.PrevBlock)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	log.Debug("PrevBlock.BlockId: %v / PrevBlock.Time: %v / PrevBlock.WalletId: %v / PrevBlock.CBID: %v / PrevBlock.Sign: %v", p.PrevBlock.BlockId, p.PrevBlock.Time, p.PrevBlock.WalletId, p.PrevBlock.CBID, p.PrevBlock.Sign)

	log.Debug("p.PrevBlock.BlockId", p.PrevBlock.BlockId)
	// для локальных тестов
	if p.PrevBlock.BlockId == 1 {
		if *utils.StartBlockId != 0 {
			p.PrevBlock.BlockId = *utils.StartBlockId
		}
	}

	var first bool
	if p.BlockData.BlockId == 1 {
		first = true
	} else {
		first = false
	}
	log.Debug("%v", first)

	// меркель рут нужен для проверки подписи блока, а также проверки лимитов MAX_TX_SIZE и MAX_TX_COUNT
	//log.Debug("p.Variables: %v", p.Variables)
	p.MrklRoot, err = utils.GetMrklroot(p.BinaryData, first)
	log.Debug("p.MrklRoot: %s", p.MrklRoot)
	if err != nil {
		return utils.ErrInfo(err)
	}

	// проверим время
	if !utils.CheckInputData(p.BlockData.Time, "int") {
		log.Debug("p.BlockData.Time", p.BlockData.Time)
		return utils.ErrInfo(fmt.Errorf("incorrect time"))
	}


	// не слишком ли рано прислан этот блок. допустима погрешность = error_time
	if !first {
		if p.PrevBlock.Time+consts.GAPS_BETWEEN_BLOCKS-p.BlockData.Time > consts.ERROR_TIME {
			return utils.ErrInfo(fmt.Errorf("incorrect block time %d + %d - %d > %d", p.PrevBlock.Time, consts.GAPS_BETWEEN_BLOCKS,  p.BlockData.Time, consts.ERROR_TIME))
		}
	}

	// исключим тех, кто сгенерил блок с бегущими часами
	if p.BlockData.Time > time.Now().Unix() {
		utils.ErrInfo(fmt.Errorf("incorrect block time"))
	}

	// проверим ID блока
	if !utils.CheckInputData(p.BlockData.BlockId, "int") {
		return utils.ErrInfo(fmt.Errorf("incorrect block_id"))
	}

	// проверим, верный ли ID блока
	if !first {
		if p.BlockData.BlockId != p.PrevBlock.BlockId+1 {
			return utils.ErrInfo(fmt.Errorf("incorrect block_id %d != %d +1", p.BlockData.BlockId, p.PrevBlock.BlockId))
		}
	}

	// проверим, есть ли такой майнер и заодно получим public_key
	nodePublicKey, err := p.GetNodePublicKeyWalletOrCB(p.BlockData.WalletId, p.BlockData.CBID)
	if err != nil {
		return utils.ErrInfo(err)
	}

	if !first {
		if len(nodePublicKey) == 0 {
			return utils.ErrInfo(fmt.Errorf("empty nodePublicKey"))
		}
		// SIGN от 128 байта до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
		forSign := fmt.Sprintf("0,%d,%s,%d,%d,%d,%s", p.BlockData.BlockId, p.PrevBlock.Hash, p.BlockData.Time, p.BlockData.WalletId, p.BlockData.CBID, p.MrklRoot)
		log.Debug(forSign)
		// проверим подпись
		resultCheckSign, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.BlockData.Sign, true)
		if err != nil {
			return utils.ErrInfo(fmt.Errorf("err: %v / p.PrevBlock.BlockId: %d", err, p.PrevBlock.BlockId))
		}
		if !resultCheckSign {
			return utils.ErrInfo(fmt.Errorf("incorrect signature / p.PrevBlock.BlockId: %d", p.PrevBlock.BlockId))
		}
	}
	return nil
}

// Это защита от dos, когда одну транзакцию можно было бы послать миллион раз,
// и она каждый раз успешно проходила бы фронтальную проверку
func (p *Parser) CheckLogTx(tx_binary []byte) error {
	hash, err := p.Single(`SELECT hash FROM rb_transactions WHERE hex(hash) = ?`, utils.Md5(tx_binary)).String()
	log.Debug("SELECT hash FROM rb_transactions WHERE hex(hash) = %s", utils.Md5(tx_binary))
	if err != nil {
		log.Error("%s", utils.ErrInfo(err))
		return utils.ErrInfo(err)
	}
	log.Debug("hash %x", hash)
	if len(hash) > 0 {
		return utils.ErrInfo(fmt.Errorf("double rb_transactions %s", utils.Md5(tx_binary)))
	}
	return nil
}

func (p *Parser) GetInfoBlock() error {

	// последний успешно записанный блок
	p.PrevBlock = new(utils.BlockData)
	var q string
	if p.ConfigIni["db_type"] == "mysql" || p.ConfigIni["db_type"] == "sqlite" {
		q = "SELECT LOWER(HEX(hash)) as hash, block_id, time FROM info_block"
	} else if p.ConfigIni["db_type"] == "postgresql" {
		q = "SELECT encode(hash, 'HEX')  as hash, block_id, time FROM info_block"
	}
	err := p.QueryRow(q).Scan(&p.PrevBlock.Hash, &p.PrevBlock.BlockId, &p.PrevBlock.Time)

	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	return nil
}

/**
 * Откат таблиц rb_time_, которые были изменены транзакциями
 */
func (p *Parser) ParseDataRollbackFront(txcandidateBlock bool) error {

	// вначале нужно получить размеры всех тр-ий, чтобы пройтись по ним в обратном порядке
	binForSize := p.BinaryData
	var sizesSlice []int64
	for {
		txSize := utils.DecodeLength(&binForSize)
		if txSize == 0 {
			break
		}
		sizesSlice = append(sizesSlice, txSize)
		// удалим тр-ию
		utils.BytesShift(&binForSize, txSize)
		if len(binForSize) == 0 {
			break
		}
	}
	sizesSlice = utils.SliceReverse(sizesSlice)
	for i := 0; i < len(sizesSlice); i++ {
		// обработка тр-ий может занять много времени, нужно отметиться
		p.UpdDaemonTime(p.GoroutineName)
		// отделим одну транзакцию
		transactionBinaryData := utils.BytesShiftReverse(&p.BinaryData, sizesSlice[i])
		// узнаем кол-во байт, которое занимает размер
		size_ := len(utils.EncodeLength(sizesSlice[i]))
		// удалим размер
		utils.BytesShiftReverse(&p.BinaryData, size_)
		p.TxHash = string(utils.Md5(transactionBinaryData))

		// инфа о предыдущем блоке (т.е. последнем занесенном)
		err := p.GetInfoBlock()
		if err != nil {
			return p.ErrInfo(err)
		}
		if txcandidateBlock {
			utils.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE hex(hash) = " + string(p.TxHash))
			affect, err := p.ExecSqlGetAffect("UPDATE transactions SET verified = 0 WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				utils.WriteSelectiveLog(err)
				return p.ErrInfo(err)
			}
			utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
		}
		affected, err := p.ExecSqlGetAffect("DELETE FROM rb_transactions WHERE hex(hash) = ?", p.TxHash)
		log.Debug("DELETE FROM rb_transactions WHERE hex(hash) = %s / affected = %d", p.TxHash, affected)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.dataType = utils.BytesToInt(p.TxSlice[1])
		//userId := p.TxSlice[3]
		MethodName := consts.TxTypes[p.dataType]
		err_ := utils.CallMethod(p, MethodName+"Init")
		if _, ok := err_.(error); ok {
			return p.ErrInfo(err_.(error))
		}
		err_ = utils.CallMethod(p, MethodName+"RollbackFront")
		if _, ok := err_.(error); ok {
			return p.ErrInfo(err_.(error))
		}
	}

	return nil
}

/**
 * Откат БД по блокам
 */
func (p *Parser) ParseDataRollback() error {

	p.dataPre()
	if p.dataType != 0 { // парсим только блоки
		return utils.ErrInfo(fmt.Errorf("incorrect dataType"))
	}
	var err error

	err = p.ParseBlock()
	if err != nil {
		return utils.ErrInfo(err)
	}
	if len(p.BinaryData) > 0 {
		// вначале нужно получить размеры всех тр-ий, чтобы пройтись по ним в обратном порядке
		binForSize := p.BinaryData
		var sizesSlice []int64
		for {
			txSize := utils.DecodeLength(&binForSize)
			if txSize == 0 {
				break
			}
			sizesSlice = append(sizesSlice, txSize)
			// удалим тр-ию
			utils.BytesShift(&binForSize, txSize)
			if len(binForSize) == 0 {
				break
			}
		}
		sizesSlice = utils.SliceReverse(sizesSlice)
		for i := 0; i < len(sizesSlice); i++ {
			// обработка тр-ий может занять много времени, нужно отметиться
			p.UpdDaemonTime(p.GoroutineName)
			// отделим одну транзакцию
			transactionBinaryData := utils.BytesShiftReverse(&p.BinaryData, sizesSlice[i])
			// узнаем кол-во байт, которое занимает размер
			size_ := len(utils.EncodeLength(sizesSlice[i]))
			// удалим размер
			utils.BytesShiftReverse(&p.BinaryData, size_)
			p.TxHash = string(utils.Md5(transactionBinaryData))

			utils.WriteSelectiveLog("UPDATE transactions SET used=0, verified = 0 WHERE hex(hash) = " + string(p.TxHash))
			affect, err := p.ExecSqlGetAffect("UPDATE transactions SET used=0, verified = 0 WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				utils.WriteSelectiveLog(err)
				return p.ErrInfo(err)
			}
			utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			affected, err := p.ExecSqlGetAffect("DELETE FROM rb_transactions WHERE hex(hash) = ?", p.TxHash)
			log.Debug("DELETE FROM rb_transactions WHERE hex(hash) = %s / affected = %d", p.TxHash, affected)
			if err != nil {
				return p.ErrInfo(err)
			}
			// даем юзеру понять, что его тр-ия не в блоке
			err = p.ExecSql("UPDATE transactions_status SET block_id = 0 WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				return p.ErrInfo(err)
			}
			// пишем тр-ию в очередь на проверку, авось пригодится
			dataHex := utils.BinToHex(transactionBinaryData)
			err = p.ExecSql("DELETE FROM queue_tx  WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", p.TxHash, dataHex)
			if err != nil {
				return p.ErrInfo(err)
			}

			p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
			if err != nil {
				return p.ErrInfo(err)
			}
			p.dataType = utils.BytesToInt(p.TxSlice[1])
			MethodName := consts.TxTypes[p.dataType]
			err_ := utils.CallMethod(p, MethodName+"Init")
			if _, ok := err_.(error); ok {
				return p.ErrInfo(err_.(error))
			}
			err_ = utils.CallMethod(p, MethodName+"Rollback")
			if _, ok := err_.(error); ok {
				return p.ErrInfo(err_.(error))
			}
			err_ = utils.CallMethod(p, MethodName+"RollbackFront")
			if _, ok := err_.(error); ok {
				return p.ErrInfo(err_.(error))
			}
		}
	}
	return nil
}

func (p *Parser) RollbackToBlockId(blockId int64) error {

	/*err := p.ExecSql("SET GLOBAL net_read_timeout = 86400")
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("SET GLOBAL max_connections  = 86400")
	if err != nil {
		return p.ErrInfo(err)
	}*/
	err := p.RollbackTransactions()
	if err != nil {
		return p.ErrInfo(err)
	}

	// откатываем наши блоки
	var blocks []map[string][]byte
	rows, err := p.Query(p.FormatQuery("SELECT id, data FROM block_chain WHERE id > ? ORDER BY id DESC"), blockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	parser := new(Parser)
	parser.DCDB = p.DCDB
	for rows.Next() {
		var data, id []byte
		err = rows.Scan(&id, &data)
		if err != nil {
			rows.Close()
			return p.ErrInfo(err)
		}
		blocks = append(blocks, map[string][]byte{"id": id, "data": data})
	}
	rows.Close()
	for _, block := range blocks {
		// Откатываем наши блоки до блока blockId
		parser.BinaryData = block["data"]
		err = parser.ParseDataRollback()
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.ExecSql("DELETE FROM block_chain WHERE id = ?", block["id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	var hash, data []byte
	err = p.QueryRow(p.FormatQuery("SELECT hash, data FROM block_chain WHERE id  =  ?"), blockId).Scan(&hash, &data)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	utils.BytesShift(&data, 1)
	block_id := utils.BinToDecBytesShift(&data, 4)
	time := utils.BinToDecBytesShift(&data, 4)
	size := utils.DecodeLength(&data)
	walletId := utils.BinToDecBytesShift(&data, size)
	CBID := utils.BinToDecBytesShift(&data, 1)
	err = p.ExecSql("UPDATE info_block SET hash = [hex], block_id = ?, time = ?, wallet_id = ?, cb_id = ?", utils.BinToHex(hash), block_id, time, walletId, CBID)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("UPDATE config SET my_block_id = ?", block_id)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RollbackTransactions() error {

	var blockBody []byte

	utils.WriteSelectiveLog("SELECT data, hash FROM transactions WHERE verified = 1 AND used = 0")
	rows, err := p.Query("SELECT data, hash FROM transactions WHERE verified = 1 AND used = 0")
	if err != nil {
		utils.WriteSelectiveLog(err)
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var data, hash []byte
		err = rows.Scan(&data, &hash)
		if err != nil {
			utils.WriteSelectiveLog(err)
			return p.ErrInfo(err)
		}
		utils.WriteSelectiveLog(utils.BinToHex(hash))
		blockBody = append(blockBody, utils.EncodeLengthPlusData(data)...)
		utils.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE hex(hash) = " + string(utils.BinToHex(hash)))
		affect, err := p.ExecSqlGetAffect("UPDATE transactions SET verified = 0 WHERE hex(hash) = ?", utils.BinToHex(hash))
		if err != nil {
			utils.WriteSelectiveLog(err)
			return p.ErrInfo(err)
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
	}

	// нужно откатить наши транзакции
	if len(blockBody) > 0 {
		parser := new(Parser)
		parser.DCDB = p.DCDB
		parser.BinaryData = blockBody
		err = parser.ParseDataRollbackFront(false)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

//  если в ходе проверки тр-ий возникает ошибка, то вызываем откатчик всех занесенных тр-ий
func (p *Parser) RollbackTo(binaryData []byte, skipCurrent bool, onlyFront bool) error {
	var err error
	if len(binaryData) > 0 {
		// вначале нужно получить размеры всех тр-ий, чтобы пройтись по ним в обратном порядке
		binForSize := binaryData
		var sizesSlice []int64
		for {
			txSize := utils.DecodeLength(&binForSize)
			if txSize == 0 {
				break
			}
			sizesSlice = append(sizesSlice, txSize)
			// удалим тр-ию
			log.Debug("txSize", txSize)
			//log.Debug("binForSize", binForSize)
			utils.BytesShift(&binForSize, txSize)
			if len(binForSize) == 0 {
				break
			}
		}
		sizesSlice = utils.SliceReverse(sizesSlice)
		for i := 0; i < len(sizesSlice); i++ {
			// обработка тр-ий может занять много времени, нужно отметиться
			p.UpdDaemonTime(p.GoroutineName)
			// отделим одну транзакцию
			transactionBinaryData := utils.BytesShiftReverse(&binaryData, sizesSlice[i])
			transactionBinaryData_ := transactionBinaryData
			// узнаем кол-во байт, которое занимает размер
			size_ := len(utils.EncodeLength(sizesSlice[i]))
			// удалим размер
			utils.BytesShiftReverse(&binaryData, size_)
			p.TxHash = string(utils.Md5(transactionBinaryData))
			p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
			if err != nil {
				return utils.ErrInfo(err)
			}
			MethodName := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
			p.TxMap = map[string][]byte{}
			err_ := utils.CallMethod(p, MethodName+"Init")
			if _, ok := err_.(error); ok {
				return utils.ErrInfo(err_.(error))
			}

			// если дошли до тр-ии, которая вызвала ошибку, то откатываем только фронтальную проверку
			if i == 0 {
				if skipCurrent { // тр-ия, которая вызвала ошибку закончилась еще до фронт. проверки, т.е. откатывать по ней вообще нечего
					continue
				}
				// если успели дойти только до половины фронтальной функции
				MethodNameRollbackFront := ""
				if p.halfRollback {
					MethodNameRollbackFront = MethodName + "RollbackFront0"
				} else {
					MethodNameRollbackFront = MethodName + "RollbackFront"
				}
				// откатываем только фронтальную проверку
				err_ = utils.CallMethod(p, MethodNameRollbackFront)
				if _, ok := err_.(error); ok {
					return utils.ErrInfo(err_.(error))
				}
			} else if onlyFront {
				err_ = utils.CallMethod(p, MethodName+"RollbackFront")
				if _, ok := err_.(error); ok {
					return utils.ErrInfo(err_.(error))
				}
			} else {
				err_ = utils.CallMethod(p, MethodName+"RollbackFront")
				if _, ok := err_.(error); ok {
					return utils.ErrInfo(err_.(error))
				}
				err_ = utils.CallMethod(p, MethodName+"Rollback")
				if _, ok := err_.(error); ok {
					return utils.ErrInfo(err_.(error))
				}
			}
			err = p.DelLogTx(transactionBinaryData_)
			if err!=nil{
				log.Error("error: %v", err)
			}
			// =================== ради эксперимента =========
			if onlyFront {
				utils.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE hex(hash) = " + string(p.TxHash))
				affect, err := p.ExecSqlGetAffect("UPDATE transactions SET verified = 0 WHERE hex(hash) = ?", p.TxHash)
				if err != nil {
					utils.WriteSelectiveLog(err)
					return utils.ErrInfo(err)
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			} else { // ====================================
				utils.WriteSelectiveLog("UPDATE transactions SET used = 0 WHERE hex(hash) = " + string(p.TxHash))
				affect, err := p.ExecSqlGetAffect("UPDATE transactions SET used = 0 WHERE hex(hash) = ?", p.TxHash)
				if err != nil {
					utils.WriteSelectiveLog(err)
					return utils.ErrInfo(err)
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			}
		}
	}
	return err
}

func (p *Parser) ParseTransaction(transactionBinaryData *[]byte) ([][]byte, error) {

	var returnSlice [][]byte
	var transSlice [][]byte
	var merkleSlice [][]byte
	log.Debug("transactionBinaryData: %x", *transactionBinaryData)
	log.Debug("transactionBinaryData: %s", *transactionBinaryData)

	if len(*transactionBinaryData) > 0 {

		// хэш транзакции
		transSlice = append(transSlice, utils.DSha256(*transactionBinaryData))

		// первый байт - тип транзакции
		transSlice = append(transSlice, utils.Int64ToByte(utils.BinToDecBytesShift(transactionBinaryData, 1)))
		if len(*transactionBinaryData) == 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect tx"))
		}

		// следующие 4 байта - время транзакции
		transSlice = append(transSlice, utils.Int64ToByte(utils.BinToDecBytesShift(transactionBinaryData, 4)))
		if len(*transactionBinaryData) == 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect tx"))
		}
		log.Debug("%s", transSlice)

		// преобразуем бинарные данные транзакции в массив
		i := 0
		for {
			length := utils.DecodeLength(transactionBinaryData)
			log.Debug("length: %d\n", length)
			if length > 0 && length < consts.MAX_TX_SIZE {
				data := utils.BytesShift(transactionBinaryData, length)
				returnSlice = append(returnSlice, data)
				merkleSlice = append(merkleSlice, utils.DSha256(data))
				log.Debug("%x", data)
				log.Debug("%s", data)
			}
			i++
			if length == 0 || i >= 20 { // у нас нет тр-ий с более чем 20 элементами
				break
			}
		}
		if len(*transactionBinaryData) > 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect transactionBinaryData %x", transactionBinaryData))
		}
	} else {
		merkleSlice = append(merkleSlice, []byte("0"))
	}
	log.Debug("merkleSlice", merkleSlice)
	if len(merkleSlice) == 0 {
		merkleSlice = append(merkleSlice, []byte("0"))
	}
	p.MerkleRoot = utils.MerkleTreeRoot(merkleSlice)
	log.Debug("MerkleRoot %s\n", p.MerkleRoot)
	return append(transSlice, returnSlice...), nil
}

func (p *Parser) InsertIntoBlockchain() error {
	//var mutex = &sync.Mutex{}
	// для локальных тестов
	if p.BlockData.BlockId == 1 {
		if *utils.StartBlockId != 0 {
			p.BlockData.BlockId = *utils.StartBlockId
		}
	}

	TxIdsJson, _ := json.Marshal(p.TxIds)

	//mutex.Lock()
	// пишем в цепочку блоков
	err := p.ExecSql("DELETE FROM block_chain WHERE id = ?", p.BlockData.BlockId)
	if err != nil {
		return err
	}
	err = p.ExecSql("INSERT INTO block_chain (id, hash, data, time, tx) VALUES (?, [hex], [hex], ?, ?)",
		p.BlockData.BlockId, p.BlockData.Hash, p.blockHex, p.BlockData.Time, TxIdsJson)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//mutex.Unlock()
	return nil
}

func (p *Parser) UpdBlockInfo() {

	blockId := p.BlockData.BlockId
	// для локальных тестов
	if p.BlockData.BlockId == 1 {
		if *utils.StartBlockId != 0 {
			blockId = *utils.StartBlockId
		}
	}
	forSha := fmt.Sprintf("%d,%s,%s,%d,%d,%d", blockId, p.PrevBlock.Hash, p.MrklRoot, p.BlockData.Time, p.BlockData.WalletId, p.BlockData.CBID)
	log.Debug("forSha", forSha)
	p.BlockData.Hash = utils.DSha256(forSha)
	log.Debug("%v", p.BlockData.Hash)
	log.Debug("%v", blockId)
	log.Debug("%v", p.BlockData.Time)
	log.Debug("%v", p.CurrentVersion)

	if p.BlockData.BlockId == 1 {
		err := p.ExecSql("INSERT INTO info_block (hash, block_id, time, cb_id, wallet_id, current_version) VALUES ([hex], ?, ?, ?, ?, ?)",
			p.BlockData.Hash, blockId, p.BlockData.Time, p.BlockData.CBID, p.BlockData.WalletId, p.CurrentVersion)
		if err!=nil {
			log.Error("%v", err)
		}
	} else {
		err := p.ExecSql("UPDATE info_block SET hash = [hex], block_id = ?, time = ?, cb_id = ?, wallet_id = ?, sent = 0",
			p.BlockData.Hash, blockId, p.BlockData.Time, p.BlockData.CBID, p.BlockData.WalletId)
		if err!=nil {
			log.Error("%v", err)
		}
		err = p.ExecSql("UPDATE config SET my_block_id = ? WHERE my_block_id < ?", blockId, blockId)
		if err!=nil {
			log.Error("%v", err)
		}
	}
}

func (p *Parser) GetTxMaps(fields []map[string]string) error {
	log.Debug("p.TxSlice %s", p.TxSlice)
	if len(p.TxSlice) != len(fields)+5 {
		return fmt.Errorf("bad transaction_array %d != %d (type=%d)", len(p.TxSlice), len(fields)+4, p.TxSlice[0])
	}
	//log.Debug("p.TxSlice", p.TxSlice)
	p.TxMap = make(map[string][]byte)
	p.TxMaps = new(txMapsType)
	p.TxMaps.Float64 = make(map[string]float64)
	p.TxMaps.Money = make(map[string]float64)
	p.TxMaps.Int64 = make(map[string]int64)
	p.TxMaps.Bytes = make(map[string][]byte)
	p.TxMaps.String = make(map[string]string)
	p.TxMaps.Bytes["hash"] = p.TxSlice[0]
	p.TxMaps.Int64["type"] = utils.BytesToInt64(p.TxSlice[1])
	p.TxMaps.Int64["time"] = utils.BytesToInt64(p.TxSlice[2])
	p.TxMaps.Int64["wallet_id"] = utils.BytesToInt64(p.TxSlice[3])
	p.TxMaps.Int64["citizen_id"] = utils.BytesToInt64(p.TxSlice[4])
	p.TxMaps.Int64["_id"] = utils.BytesToInt64(p.TxSlice[4])
	p.TxMap["hash"] = p.TxSlice[0]
	p.TxMap["type"] = p.TxSlice[1]
	p.TxMap["time"] = p.TxSlice[2]
	p.TxMap["wallet_id"] = p.TxSlice[3]
	p.TxMap["citizen_id"] = p.TxSlice[4]
	for i := 0; i < len(fields); i++ {
		for field, fType := range fields[i] {
			p.TxMap[field] = p.TxSlice[i+5]
			switch fType {
			case "int64":
				p.TxMaps.Int64[field] = utils.BytesToInt64(p.TxSlice[i+5])
			case "float64":
				p.TxMaps.Float64[field] = utils.BytesToFloat64(p.TxSlice[i+5])
			case "money":
				p.TxMaps.Money[field] = utils.StrToMoney(string(p.TxSlice[i+5]))
			case "bytes":
				p.TxMaps.Bytes[field] = p.TxSlice[i+5]
			case "string":
				p.TxMaps.String[field] = string(p.TxSlice[i+5])
			}
		}
	}
	log.Debug("%s", p.TxMaps)
	p.TxCitizenID = p.TxMaps.Int64["citizen_id"]
	p.TxWalletID = p.TxMaps.Int64["wallet_id"]
	p.TxTime = p.TxMaps.Int64["time"]
	p.PublicKeys = nil
	//log.Debug("p.TxMaps", p.TxMaps)
	//log.Debug("p.TxMap", p.TxMap)
	return nil
}

// старое
func (p *Parser) GetTxMap(fields []string) (map[string][]byte, error) {
	if len(p.TxSlice) != len(fields)+4 {
		return nil, fmt.Errorf("bad transaction_array %d != %d (type=%d)", len(p.TxSlice), len(fields)+4, p.TxSlice[0])
	}
	TxMap := make(map[string][]byte)
	TxMap["hash"] = p.TxSlice[0]
	TxMap["type"] = p.TxSlice[1]
	TxMap["time"] = p.TxSlice[2]
	TxMap["user_id"] = p.TxSlice[3]
	for i, field := range fields {
		TxMap[field] = p.TxSlice[i+4]
	}
	p.TxUserID = utils.BytesToInt64(TxMap["user_id"])
	p.TxTime = utils.BytesToInt64(TxMap["time"])
	p.PublicKeys = nil
	//log.Debug("TxMap", TxMap)
	//log.Debug("TxMap[hash]", TxMap["hash"])
	//log.Debug("p.TxSlice[0]", p.TxSlice[0])
	return TxMap, nil
}

// старое
func (p *Parser) GetTxMapStr(fields []string) (map[string]string, error) {
	//log.Debug("p.TxSlice", p.TxSlice)
	//log.Debug("fields", fields)
	if len(p.TxSlice) != len(fields)+4 {
		return nil, fmt.Errorf("bad transaction_array %d != %d (type=%d)", len(p.TxSlice), len(fields)+4, p.TxSlice[0])
	}
	TxMapS := make(map[string]string)
	TxMapS["hash"] = string(p.TxSlice[0])
	TxMapS["type"] = string(p.TxSlice[1])
	TxMapS["time"] = string(p.TxSlice[2])
	TxMapS["user_id"] = string(p.TxSlice[3])
	for i, field := range fields {
		TxMapS[field] = string(p.TxSlice[i+4])
	}
	p.TxUserID = utils.StrToInt64(TxMapS["user_id"])
	p.TxTime = utils.StrToInt64(TxMapS["time"])
	p.PublicKeys = nil
	log.Debug("TxMapS", TxMapS)
	//log.Debug("TxMap[hash]", TxMap["hash"])
	//log.Debug("p.TxSlice[0]", p.TxSlice[0])
	return TxMapS, nil
}

func (p *Parser) CheckInputData(data map[string]string) error {
	for k, v := range data {
		if !utils.CheckInputData(p.TxMap[k], v) {
			return fmt.Errorf("incorrect " + k)
		}
	}
	return nil
}

func (p *Parser) limitRequestsRollback(txType string) error {
	time := p.TxMap["time"]
	if p.ConfigIni["db_type"] == "mysql" {
		return p.ExecSql("DELETE FROM rb_time_"+txType+" WHERE user_id = ? AND time = ? LIMIT 1", p.TxUserID, time)
	} else if p.ConfigIni["db_type"] == "postgresql" {
		return p.ExecSql("DELETE FROM rb_time_"+txType+" WHERE ctid IN (SELECT ctid FROM rb_time_"+txType+" WHERE  user_id = ? AND time = ? LIMIT 1)", p.TxUserID, time)
	} else {
		return p.ExecSql("DELETE FROM rb_time_"+txType+" WHERE id IN (SELECT id FROM rb_time_"+txType+" WHERE  user_id = ? AND time = ? LIMIT 1)", p.TxUserID, time)
	}
	return nil
}

// откатываем ID на кол-во затронутых строк
func (p *Parser) rollbackAI(table string, num int64) error {

	if num == 0 {
		return nil
	}

	AiId, err := p.GetAiId(table)
	if err != nil {
		return utils.ErrInfo(err)
	}
	log.Debug("AiId: %s", AiId)
	// если табла была очищена, то тут будет 0, поэтому нелья чистить таблы под нуль
	current, err := p.Single("SELECT " + AiId + " FROM " + table + " ORDER BY " + AiId + " DESC LIMIT 1").Int64()
	if err != nil {
		return utils.ErrInfo(err)
	}
	NewAi := current + num
	log.Debug("NewAi: %d", NewAi)

	if p.ConfigIni["db_type"] == "postgresql" {
		pg_get_serial_sequence, err := p.Single("SELECT pg_get_serial_sequence('" + table + "', '" + AiId + "')").String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = p.ExecSql("ALTER SEQUENCE " + pg_get_serial_sequence + " RESTART WITH " + utils.Int64ToStr(NewAi))
		if err != nil {
			return utils.ErrInfo(err)
		}
	} else if p.ConfigIni["db_type"] == "mysql" {
		err := p.ExecSql("ALTER TABLE " + table + " AUTO_INCREMENT = " + utils.Int64ToStr(NewAi))
		if err != nil {
			return utils.ErrInfo(err)
		}
	} else if p.ConfigIni["db_type"] == "sqlite" {
		NewAi--
		err := p.ExecSql("UPDATE SQLITE_SEQUENCE SET seq = ? WHERE name = ?", NewAi, table)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) generalRollback(table string, whereUserId_ interface{}, addWhere string, AI bool) error {
	var whereUserId int64
	switch whereUserId_.(type) {
	case string:
		whereUserId = utils.StrToInt64(whereUserId_.(string))
	case []byte:
		whereUserId = utils.BytesToInt64(whereUserId_.([]byte))
	case int:
		whereUserId = int64(whereUserId_.(int))
	case int64:
		whereUserId = whereUserId_.(int64)
	}

	where := ""
	if whereUserId > 0 {
		where = fmt.Sprintf(" WHERE user_id = %d ", whereUserId)
	}
	// получим rb_id, по которому можно найти данные, которые были до этого
	logId, err := p.Single("SELECT rb_id FROM " + table + " " + where + addWhere).Int64()
	if err != nil {
		return utils.ErrInfo(err)
	}
	// если $rb_id = 0, значит восстанавливать нечего и нужно просто удалить запись
	if logId == 0 {
		err = p.ExecSql("DELETE FROM " + table + " " + where + addWhere)
		if err != nil {
			return utils.ErrInfo(err)
		}
	} else {
		// данные, которые восстановим
		data, err := p.OneRow("SELECT * FROM rb_"+table+" WHERE rb_id = ?", logId).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		addSql := ""
		for k, v := range data {
			// block_id т.к. в rb_ он нужен для удаления старых данных, а в обычной табле не нужен
			if k == "rb_id" || k == "prev_rb_id" || k == "block_id" {
				continue
			}
			if k == "node_public_key" {
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					addSql += fmt.Sprintf("%v='%x',", k, v)
				case "postgresql":
					addSql += fmt.Sprintf("%v=decode('%x','HEX'),", k, v)
				case "mysql":
					addSql += fmt.Sprintf("%v=UNHEX('%x'),", k, v)
				}
			} else {
				addSql += fmt.Sprintf("%v = '%v',", k, v)
			}
		}
		// всегда пишем предыдущий rb_id
		addSql += fmt.Sprintf("rb_id = %v,", data["prev_rb_id"])
		addSql = addSql[0 : len(addSql)-1]
		err = p.ExecSql("UPDATE " + table + " SET " + addSql + where + addWhere)
		if err != nil {
			return utils.ErrInfo(err)
		}
		// подчищаем log
		err = p.ExecSql("DELETE FROM rb_"+table+" WHERE rb_id= ?", logId)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = p.rollbackAI("rb_"+table, 1)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

func arrayIntersect(arr1, arr2 map[int]int) bool {
	for _, v := range arr1 {
		for _, v2 := range arr2 {
			if v == v2 {
				return true
			}
		}
	}
	return false
}

func (p *Parser) minersCheckMyMinerIdAndVotes0(data *MinerData) bool {
	log.Debug("data.myMinersIds", data.myMinersIds)
	log.Debug("data.minersIds", data.minersIds)
	log.Debug("data.votes0", data.votes0)
	log.Debug("data.minMinersKeepers", data.minMinersKeepers)
	log.Debug("int(data.votes0)", int(data.votes0))
	log.Debug("len(data.minersIds)", len(data.minersIds))
	if (arrayIntersect(data.myMinersIds, data.minersIds)) && (data.votes0 > data.minMinersKeepers || int(data.votes0) == len(data.minersIds)) {
		return true
	} else {
		return false
	}
}

func (p *Parser) minersCheckVotes1(data *MinerData) bool {
	log.Debug("data.votes1", data.votes1)
	log.Debug("data.minMinersKeepers", data.minMinersKeepers)
	log.Debug("data.minersIds", len(data.minersIds))
	if data.votes1 >= data.minMinersKeepers || int(data.votes1) == len(data.minersIds) /*|| data.adminUiserId == p.TxUserID Админская нода не решающая*/ {
		log.Debug("true")
		return true
	} else {
		return false
	}
}

func (p *Parser) FormatBlockData() string {
	result := ""
	if p.BlockData != nil {
		v := reflect.ValueOf(*p.BlockData)
		typeOfT := v.Type()
		if typeOfT.Kind() == reflect.Ptr {
			typeOfT = typeOfT.Elem()
		}
		for i := 0; i < v.NumField(); i++ {
			name := typeOfT.Field(i).Name
			switch name {
			case "BlockId", "Time", "UserId", "Level":
				result += "[" + name + "] = " + fmt.Sprintf("%d\n", v.Field(i).Interface())
			case "Sign", "Hash", "HeadHash":
				result += "[" + name + "] = " + fmt.Sprintf("%x\n", v.Field(i).Interface())
			default:
				result += "[" + name + "] = " + fmt.Sprintf("%s\n", v.Field(i).Interface())
			}
		}
	}
	return result
}

func (p *Parser) FormatTxMap() string {
	result := ""
	for k, v := range p.TxMap {
		switch k {
		case "sign":
			result += "[" + k + "] = " + fmt.Sprintf("%x\n", v)
		default:
			result += "[" + k + "] = " + fmt.Sprintf("%s\n", v)
		}
	}
	return result
}

func (p *Parser) ErrInfo(err_ interface{}) error {
	var err error
	switch err_.(type) {
	case error:
		err = err_.(error)
	case string:
		err = fmt.Errorf(err_.(string))
	}
	return fmt.Errorf("[ERROR] %s (%s)\n%s\n%s", err, utils.Caller(1), p.FormatBlockData(), p.FormatTxMap())
}



// не использовать для комментов
func (p *Parser) selectiveLoggingAndUpd(fields []string, values_ []interface{}, table string, whereFields, whereValues []string) error {

	values := utils.InterfaceSliceToStr(values_)

	addSqlFields := ""
	for _, field := range fields {
		addSqlFields += field + ","
	}

	addSqlWhere := ""
	if whereFields!=nil && whereValues!=nil {
		for i := 0; i < len(whereFields); i++ {
			addSqlWhere += whereFields[i] + "=" + whereValues[i] + " AND "
		}
	}
	if len(addSqlWhere) > 0 {
		addSqlWhere = " WHERE " + addSqlWhere[0:len(addSqlWhere)-5]
	}
	// если есть, что логировать
	logData, err := p.OneRow("SELECT " + addSqlFields + " rb_id FROM " + table + " " + addSqlWhere).String()
	if err != nil {
		return err
	}
	if len(logData) > 0 {
		addSqlValues := ""
		addSqlFields := ""
		for k, v := range logData {
			if utils.InSliceString(k, []string{"hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}) && v != "" {
				v := string(utils.BinToHex([]byte(v)))
				query := ""
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					query = `x'` + v + `',`
				case "postgresql":
					query = `decode('` + v + `','HEX'),`
				case "mysql":
					query = `UNHEX("` + v + `"),`
				}
				addSqlValues += query
			} else {
				addSqlValues += `'` + v + `',`
			}
			if k == "rb_id" {
				k = "prev_rb_id"
			}
			addSqlFields += k + ","
		}
		addSqlValues = addSqlValues[0 : len(addSqlValues)-1]
		addSqlFields = addSqlFields[0 : len(addSqlFields)-1]

		logId, err := p.ExecSqlGetLastInsertId("INSERT INTO rb_"+table+" ( "+addSqlFields+", block_id ) VALUES ( "+addSqlValues+", ? )", "rb_id", p.BlockData.BlockId)
		if err != nil {
			return err
		}
		addSqlUpdate := ""
		for i := 0; i < len(fields); i++ {
			if utils.InSliceString(fields[i], []string{"hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}) && len(values[i]) != 0 {
				query := ""
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					query = fields[i] + `=x'` + values[i] + `',`
				case "postgresql":
					query = fields[i] + `=decode('` + values[i] + `','HEX'),`
				case "mysql":
					query = fields[i] + `=UNHEX("` + values[i] + `"),`
				}
				addSqlUpdate += query
			} else {
				addSqlUpdate += fields[i] + `='` + values[i] + `',`
			}
		}
		err = p.ExecSql("UPDATE "+table+" SET "+addSqlUpdate+" rb_id = ? "+addSqlWhere, logId)
		//log.Debug("UPDATE "+table+" SET "+addSqlUpdate+" rb_id = ? "+addSqlWhere)
		//log.Debug("logId", logId)
		if err != nil {
			return err
		}
	} else {
		addSqlIns0 := ""
		addSqlIns1 := ""
		for i := 0; i < len(fields); i++ {
			addSqlIns0 += `` + fields[i] + `,`
			if utils.InSliceString(fields[i], []string{"hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}) && len(values[i]) != 0 {
				query := ""
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					query = `x'` + values[i] + `',`
				case "postgresql":
					query = `decode('` + values[i] + `','HEX'),`
				case "mysql":
					query = `UNHEX("` + values[i] + `"),`
				}
				addSqlIns1 += query
			} else {
				addSqlIns1 += `'` + values[i] + `',`
			}
		}
		for i := 0; i < len(whereFields); i++ {
			addSqlIns0 += `` + whereFields[i] + `,`
			addSqlIns1 += `'` + whereValues[i] + `',`
		}
		addSqlIns0 = addSqlIns0[0 : len(addSqlIns0)-1]
		addSqlIns1 = addSqlIns1[0 : len(addSqlIns1)-1]
		err = p.ExecSql("INSERT INTO " + table + " (" + addSqlIns0 + ") VALUES (" + addSqlIns1 + ")")
		if err != nil {
			return err
		}
	}
	return nil
}


func (p *Parser) limitRequestsMoneyOrdersRollback() error {
	err := p.ExecSql("DELETE FROM rb_time_money_orders WHERE hex(tx_hash) = ?", p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// откат не всех полей, а только указанных, либо 1 строку, если нет where
func (p *Parser) selectiveRollback(fields []string, table string, where string, rollback bool) error {
	if len(where) > 0 {
		where = " WHERE " + where
	}
	addSqlFields := ""
	for _, field := range fields {
		addSqlFields += field + ","
	}
	// получим rb_id, по которому можно найти данные, которые были до этого
	logId, err := p.Single("SELECT rb_id FROM " + table + " " + where).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if logId > 0 {
		// данные, которые восстановим
		logData, err := p.OneRow("SELECT "+addSqlFields+" prev_rb_id FROM rb_"+table+" WHERE rb_id  =  ?", logId).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		//log.Debug("logData",logData)
		addSqlUpdate := ""
		for _, field := range fields {
			if utils.InSliceString(field, []string{"hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2"}) && len(logData[field]) != 0 {
				query := ""
				logData[field] = string(utils.BinToHex([]byte(logData[field])))
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					query = field + `=x'` + logData[field] + `',`
				case "postgresql":
					query = field + `=decode('` + logData[field] + `','HEX'),`
				case "mysql":
					query = field + `=UNHEX("` + logData[field] + `"),`
				}
				addSqlUpdate += query
			} else {
				addSqlUpdate += field + `='` + logData[field] + `',`
			}
		}
		err = p.ExecSql("UPDATE "+table+" SET "+addSqlUpdate+" rb_id = ? "+where, logData["prev_rb_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		// подчищаем _log
		err = p.ExecSql("DELETE FROM rb_"+table+" WHERE rb_id = ?", logId)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.rollbackAI("rb_"+table, 1)
	} else {
		err = p.ExecSql("DELETE FROM " + table + " " + where)
		if err != nil {
			return p.ErrInfo(err)
		}
		if rollback {
			p.rollbackAI(table, 1)
		}
	}

	return nil
}

func (p *Parser) getMyNodeCommission(currencyId, userId int64, amount float64) (float64, error) {
	return consts.COMMISSION, nil

}

func (p *Parser) getWalletsBufferAmount(currencyId int64) (float64, error) {
	return p.Single("SELECT sum(amount) FROM dlt_wallets_buffer WHERE user_id = ? AND currency_id = ? AND del_block_id = 0", p.TxUserID, currencyId).Float64()
}

func (p *Parser) updateWalletsBuffer(amount float64, currencyId int64) error {
	// добавим нашу сумму в буфер кошельков, чтобы юзер не смог послать запрос на вывод всех DC с кошелька.
	hash, err := p.Single("SELECT hash FROM dlt_wallets_buffer WHERE hex(hash) = ?", p.TxHash).String()
	if len(hash) > 0 {
		err = p.ExecSql("UPDATE wallets_buffer SET user_id = ?, currency_id = ?, amount = ? WHERE hex(hash) = ?", p.TxUserID, currencyId, utils.Round(amount, 2), p.TxHash)
	} else {
		err = p.ExecSql("INSERT INTO wallets_buffer ( hash, user_id, currency_id, amount ) VALUES ( [hex], ?, ?, ? )", p.TxHash, p.TxUserID, currencyId, utils.Round(amount, 2))
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RollbackIncompatibleTx(typesArr []string) error {

	var whereType string
	for _, txType := range typesArr {
		whereType += utils.Int64ToStr(utils.TypeInt(txType)) + ","
	}
	whereType = whereType[:len(whereType)-1]

	utils.WriteSelectiveLog(`SELECT data FROM transactions WHERE type IN (` + whereType + `) AND verified=1 AND used = 0`)
	transactions, err := p.GetList(`SELECT data FROM transactions WHERE type IN (` + whereType + `) AND verified=1 AND used = 0`).String()
	if err != nil {
		utils.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	for _, txData := range transactions {

		md5 := utils.Md5(txData)
		utils.WriteSelectiveLog("md5: " + string(md5))
		// откатим фронтальные записи
		p.BinaryData = utils.EncodeLengthPlusData([]byte(txData))
		err = p.ParseDataRollback()
		if err != nil {
			return utils.ErrInfo(err)
		}
		// Удаляем уже записанные тр-ии.

		utils.WriteSelectiveLog("DELETE FROM transactions WHERE hex(hash) = " + string(md5))
		affect, err := p.ExecSqlGetAffect("DELETE FROM transactions WHERE hex(hash) = ?", md5)
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))


		// создаем тр-ию, которую потом заново проверим
		err = p.ExecSql("DELETE FROM queue_tx  WHERE hex(hash) = ?", md5)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = p.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, utils.BinToHex([]byte(txData)))
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) ClearIncompatibleTx(binaryTx []byte, myTx bool) (string, string, int64, int64, int64, int64, int64) {

	var fatalError, waitError string
	var toUserId int64

	// получим тип тр-ии и юзера
	txType, walletId, citizenId, thirdVar := utils.GetTxTypeAndUserId(binaryTx)

	if !utils.CheckInputData(txType, "int") {
		fatalError = "error type"
	}
	if !utils.CheckInputData(walletId, "int") {
		fatalError = "error walletId"
	}
	if !utils.CheckInputData(citizenId, "int") {
		fatalError = "error citizenId"
	}
	if !utils.CheckInputData(thirdVar, "int") {
		fatalError = "error thirdVar"
	}


	var forSelfUse int64
	if utils.InSliceInt64(txType, utils.TypesToIds([]string{"NewPct", "NewReduction", "NewMaxPromisedAmounts", "NewMaxOtherCurrencies"})) {
		//  чтобы никому не слать эту тр-ю
		forSelfUse = 1
		// $my_tx == true - это значит функция вызвана из pct_generator reduction_generator
		// если же false, то она была спаршена query_tx или tesblock_generator и имела verified=0
		// а т.к. new_pct/NewReduction актуальны только 1 блок, то нужно её удалять
		if !myTx {
			fatalError = "old new_pct/NewReduction/NewMaxPromisedAmounts/NewMaxOtherCurrencies"
			return fatalError, waitError, forSelfUse, txType, walletId, citizenId, toUserId
		}
	} else {
		forSelfUse = 0
	}

	// две тр-ии одного типа от одного юзера не должны попасть в один блок
	// исключение - перевод DC между юзерами
	if len(fatalError) == 0 {
		p.ClearIncompatibleTxSql(txType, walletId, citizenId, &waitError)


		// нельзя удалять CF-проект и в этом же блоке изменить его описание/профинансировать
		if txType == utils.TypeInt("DelCfProject") {
			p.ClearIncompatibleTxSqlSet([]string{"CfSendDc"}, 0, 0, &waitError, thirdVar)
		}
		if utils.InSliceInt64(txType, utils.TypesToIds([]string{"CfSendDc"})) {
			p.ClearIncompatibleTxSqlSet([]string{"DelCfProject"}, 0, 0, &waitError, thirdVar)
		}

		// потом нужно сделать более тонко. но пока так. Если есть удаление проекта, тогда откатываем все тр-ии del_cf_funding
		if txType == utils.TypeInt("DelCfProject") {
			p.RollbackIncompatibleTx([]string{"DelCfFunding"})
		}

		// Если есть смена коммиссий арбитров, то нельзя делать перевод монет, т.к. там может быть указана комиссия арбитра
		if utils.InSliceInt64(txType, utils.TypesToIds([]string{"SendDc"})) {
			p.RollbackIncompatibleTx([]string{"ChangeArbitratorConditions"})
		}
		if txType == utils.TypeInt("ChangeArbitratorConditions") {
			p.ClearIncompatibleTxSqlSet([]string{"SendDc"}, 0, 0, &waitError, "")
		}


		// на всякий случай не даем попасть в один блок тр-ии отправки в CF-проект монет и другим тр-ям связанным с этим CF-проектом. Т.к. проект может завершиться и 2-я тр-я вызовет ошибку
		if txType == utils.TypeInt("CfSendDc") {
			p.ClearIncompatibleTxSqlSet([]string{"DelCfProject"}, 0, 0, &waitError, thirdVar)
		}
		if utils.InSliceInt64(txType, utils.TypesToIds([]string{"DelCfProject"})) {
			p.ClearIncompatibleTxSqlSet([]string{"CfSendDc"}, 0, 0, &waitError, thirdVar)
		}

		// в один блок должен попасть только один голос за один объект голосования. thirdVar - объект голосования
		if utils.InSliceInt64(txType, utils.TypesToIds([]string{"VotesPromisedAmount", "VotesMiner", "VotesNodeNewMiner", "VotesComplex"})) {
			num, err := p.Single(`
			  			  SELECT count(*)
				            FROM (
					            SELECT citizen_id
					            FROM transactions
					            WHERE  type IN (?, ?, ?, ?) AND
					                          third_var = ? AND
					                          verified=1 AND
					                          used = 0
							)  AS x
							`, utils.TypeInt("VotesPromisedAmount"), utils.TypeInt("VotesMiner"), utils.TypeInt("VotesNodeNewMiner"), utils.TypeInt("VotesComplex"), thirdVar, utils.TypeInt("VotesPromisedAmount"), utils.TypeInt("VotesMiner"), utils.TypeInt("VotesNodeNewMiner"), utils.TypeInt("VotesComplex"), thirdVar).Int64()
			if err != nil {
				fatalError = fmt.Sprintf("%s", err)
			}
			if num > 0 {
				waitError = "only 1 vote"
			}
		}

		// если новая тр-я - это смена праймари ключа, то не должно быть никаких других тр-ий от этого юзера
		if txType == utils.TypeInt("ChangePrimaryKey") {
			num, err := p.Single(`
						  SELECT count(*)
				            FROM (
					            SELECT citizen_id
					            FROM transactions
					            WHERE  user_id = ? AND
					                         verified=1 AND
					                         used = 0
							)  AS x
							`, citizenId, citizenId).Int64()
			if err != nil {
				fatalError = fmt.Sprintf("%s", err)
			}
			if num > 0 {
				waitError = "there are other tr-s"
			}
		}

		// любая тр-я от юзера не должна проходить, если уже есть тр-я со сменой праймари ключа или new_pct или NewReduction
		num, err := p.Single(`
						SELECT count(*)
				          FROM (
					            SELECT citizen_id
					            FROM transactions
					            WHERE  (
						                            (type = ? AND citizen_id = ?)
						                            OR
						                            (type IN (?, ?) )
					                          ) AND
					                         verified=1 AND
					                         used = 0

						)  AS x
						`, utils.TypeInt("ChangePrimaryKey"), citizenId, utils.TypeInt("NewPct"), utils.TypeInt("NewReduction"), utils.TypeInt("ChangePrimaryKey"), citizenId, utils.TypeInt("NewPct"), utils.TypeInt("NewReduction")).Int64()
		if err != nil {
			fatalError = fmt.Sprintf("%s", err)
		}
		if num > 0 {
			waitError = "have ChangePrimaryKey tx"
		}


		// временно запрещаем 2 тр-ии любого типа от одного юзера, а то затрахался уже.
		num, err = p.Single(`
				    SELECT count(*)
				    FROM (
							SELECT citizen_id
							FROM transactions
							WHERE  citizen_id = ? AND
				                      verified=1 AND
				                      used = 0
					)  AS x
					`, citizenId, citizenId).Int64()
		if err != nil {
			fatalError = fmt.Sprintf("%s", err)
		}
		if num > 0 {
			waitError = "only 1 tx"
		}
	}
	log.Debug("fatalError: %v, waitError: %v, forSelfUse: %v, txType: %v, walletId: %v, citizenId: %v, thirdVar: %v", fatalError, waitError, forSelfUse, txType, walletId, citizenId, thirdVar)
	return fatalError, waitError, forSelfUse, txType, walletId, citizenId, thirdVar

}

func (p *Parser) TxParser(hash, binaryTx []byte, myTx bool) error {

	// проверим, нет ли несовместимых тр-ий
	// 	&waitError  - значит просто откладываем обработку тр-ии на после того, как сформируются блок
	// $fatal_error - удаляем тр-ию, т.к. она некорректная

	var err error
	fatalError, waitError, forSelfUse, txType, walletId, citizenId, thirdVar := p.ClearIncompatibleTx(binaryTx, myTx)
	if len(fatalError) == 0 && len(waitError) == 0 {
		p.BinaryData = binaryTx
		err = p.ParseDataGate(false)
	}

	hashHex := utils.BinToHex(hash)
	if err != nil || len(fatalError) > 0 {
		p.DeleteQueueTx(hashHex) // удалим тр-ию из очереди
	}
	if err == nil && len(fatalError) > 0 {
		err = errors.New(fatalError)
	}
	if err == nil && len(waitError) > 0 {
		err = errors.New(waitError)
	}
	if err != nil {
		log.Error("err: %v", err)
		errText := fmt.Sprintf("%s", err)
		if len(errText) > 255 {
			errText = errText[:255]
		}
		err = p.ExecSql("UPDATE transactions_status SET error = ? WHERE hex(hash) = ?", errText, hashHex)
		if err != nil {
			return utils.ErrInfo(err)
		}
	} else {

		log.Debug("SELECT counter FROM transactions WHERE hex(hash) = ?", string(hashHex))
		utils.WriteSelectiveLog("SELECT counter FROM transactions WHERE hex(hash) = " + string(hashHex))
		counter, err := p.Single("SELECT counter FROM transactions WHERE hex(hash) = ?", hashHex).Int64()
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("counter: " + utils.Int64ToStr(counter))
		counter++
		utils.WriteSelectiveLog("DELETE FROM transactions WHERE hex(hash) = " + string(hashHex))
		affect, err := p.ExecSqlGetAffect(`DELETE FROM transactions WHERE hex(hash) = ?`, hashHex)
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))

		log.Debug("INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter) VALUES (%s, %s, %v, %v, %v, %v, %v, %v)", hashHex, utils.BinToHex(binaryTx), forSelfUse, txType, walletId, citizenId, thirdVar, counter)
		utils.WriteSelectiveLog("INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter) VALUES ([hex], [hex], ?, ?, ?, ?, ?, ?)")
		// вставляем с verified=1
		err = p.ExecSql(`INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter, verified) VALUES ([hex], [hex], ?, ?, ?, ?, ?, ?, 1)`, hashHex, utils.BinToHex(binaryTx), forSelfUse, txType, walletId, citizenId, thirdVar, counter)
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("result insert")
		log.Debug("INSERT INTO transactions - OK")
		// удалим тр-ию из очереди (с verified=0)
		err = p.DeleteQueueTx(hashHex)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) DeleteQueueTx(hashHex []byte) error {

	err := p.ExecSql("DELETE FROM queue_tx WHERE hex(hash) = ?", hashHex)
	if err != nil {
		return utils.ErrInfo(err)
	}
	// т.к. мы обрабатываем в queue_parser_tx тр-ии с verified=0, то после их обработки их нужно удалять.
	utils.WriteSelectiveLog("DELETE FROM transactions WHERE hex(hash) = " + string(hashHex) + " AND verified=0 AND used = 0")
	affect, err := p.ExecSqlGetAffect("DELETE FROM transactions WHERE hex(hash) = ? AND verified=0 AND used = 0", hashHex)
	if err != nil {
		utils.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
	return nil
}

func (p *Parser) AllTxParser() error {

	// берем тр-ии
	all, err := p.GetAll(`
			SELECT *
			FROM (
	              SELECT data,
	                         hash
	              FROM queue_tx
				UNION
				SELECT data,
							 hash
				FROM transactions
				WHERE verified = 0 AND
							 used = 0
			)  AS x
			`, -1)
	for _, data := range all {

		log.Debug("hash: %x", data["hash"])

		err = p.TxParser([]byte(data["hash"]), []byte(data["data"]), false)
		if err != nil {
			err0 := p.ExecSql(`INSERT INTO incorrect_tx (time, hash, err) VALUES (?, [hex], ?)`, utils.Time(), utils.BinToHex(data["hash"]), fmt.Sprintf("%s", err))
			if err0 != nil {
				log.Error("%v", utils.ErrInfo(err0))
			}
			return utils.ErrInfo(err)
		}
	}
	return nil
}
