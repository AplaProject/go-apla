package parser

import (
	"errors"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/**
Обработка данных (блоков или транзакций), пришедших с гейта. Только проверка.
Processing data (blocks or transactions) gotten from a gate. Just checking.
*/
func (p *Parser) ParseDataGate(onlyTx bool) error {

	var err error
	p.dataPre()
	p.TxIds = []string{}

	p.Variables, err = p.GetAllVariables()
	if err != nil {
		return utils.ErrInfo(err)
	}

	transactionBinaryData := p.BinaryData
	var transactionBinaryDataFull []byte

	log.Debug("p.dataType: %d", p.dataType)
	// если это транзакции (type>0), а не блок (type==0)
	// if it's transactions, but block
	if p.dataType > 0 {

		// проверим, есть ли такой тип тр-ий
		// check if the transaction's type exist
		if len(consts.TxTypes[p.dataType]) == 0 {
			return p.ErrInfo("Incorrect tx type " + utils.IntToStr(p.dataType))
		}

		log.Debug("p.dataType: %d", p.dataType)
		transactionBinaryData = append(utils.DecToBin(int64(p.dataType), 1), transactionBinaryData...)
		transactionBinaryDataFull = transactionBinaryData

		// нет ли хэша этой тр-ии у нас в БД?
		// Does the transaction's hash exist?
		err = p.CheckLogTx(transactionBinaryDataFull)
		if err != nil {
			return p.ErrInfo(err)
		}

		p.TxHash = string(utils.Md5(transactionBinaryData))

		// преобразуем бинарные данные транзакции в массив
		// transforming binary data of the transaction to an array
		log.Debug("transactionBinaryData: %x", transactionBinaryData)
		p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("p.TxSlice", p.TxSlice)
		if len(p.TxSlice) < 3 {
			return p.ErrInfo(errors.New("len(p.TxSlice) < 3"))
		}

		// время транзакции может быть немного больше, чем время на ноде.
		// у нода может быть просто не настроено время.
		// время транзакции используется только для борьбы с атаками вчерашними транзакциями.
		// А т.к. мы храним хэши в log_transaction за 36 часов, то боятся нечего.

		// Time of transaction can be slightly longer than time of a node.
		// A node can use wrong time
		// Time of a transaction used only for fighting off attacks of yesterday transactions
		curTime := utils.Time()
		if utils.BytesToInt64(p.TxSlice[2])-consts.MAX_TX_FORW > curTime || utils.BytesToInt64(p.TxSlice[2]) < curTime-consts.MAX_TX_BACK {
			return p.ErrInfo(errors.New("incorrect tx time"))
		}
		// $this->transaction_array[3] могут подсунуть пустой
		if !utils.CheckInputData(p.TxSlice[3], "bigint") {
			return p.ErrInfo(errors.New("incorrect user id"))
		}
	}

	// если это блок
	// if it's a block
	if p.dataType == 0 {

		txCounter := make(map[int64]int64)

		// если есть $only_tx=true, то значит идет восстановление уже проверенного блока и заголовок не требуется
		// if $only_tx=true, there is a recovery of already checked block and no need in a header
		if !onlyTx {
			err = p.ParseBlock()
			if err != nil {
				return p.ErrInfo(err)
			}

			// проверим данные, указанные в заголовке блока
			err = p.CheckBlockHeader()
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		log.Debug("onlyTx", onlyTx)

		// если в ходе проверки тр-ий возникает ошибка, то вызываем откатчик всех занесенных тр-ий. Эта переменная для него
		// if an error occur during the checking, call 'rollbacker' for all written transactions. This is a variable for it.
		p.fullTxBinaryData = p.BinaryData
		var txForRollbackTo []byte
		if len(p.BinaryData) > 0 {
			for {
				transactionSize := utils.DecodeLength(&p.BinaryData)
				if len(p.BinaryData) == 0 {
					return utils.ErrInfo(fmt.Errorf("empty BinaryData"))
				}

				// отчекрыжим одну транзакцию от списка транзакций
				// get rid of one transaction from the list
				transactionBinaryData := utils.BytesShift(&p.BinaryData, transactionSize)
				transactionBinaryDataFull = transactionBinaryData

				// добавляем взятую тр-ию в набор тр-ий для RollbackTo, в котором пойдем в обратном порядке
				// add taken transaction to a set of transactions for RollbackTo. There we will go in opposite direction
				txForRollbackTo = append(txForRollbackTo, utils.EncodeLengthPlusData(transactionBinaryData)...)

				// нет ли хэша этой тр-ии у нас в БД?
				// Is there a hash of the transaction in DB?
				err = p.CheckLogTx(transactionBinaryDataFull)
				if err != nil {
					p.RollbackTo(txForRollbackTo, true, false)
					return p.ErrInfo(err)
				}

				p.TxHash = string(utils.Md5(transactionBinaryData))
				p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
				log.Debug("p.TxSlice %s", p.TxSlice)
				if err != nil {
					p.RollbackTo(txForRollbackTo, true, false)
					return p.ErrInfo(err)
				}

				var userId int64
				// txSlice[3] могут подсунуть пустой
				// txSlice[3] can be empty
				if len(p.TxSlice) > 3 {
					if !utils.CheckInputData(p.TxSlice[3], "int64") {
						return utils.ErrInfo(fmt.Errorf("empty user_id"))
					} else {
						userId = utils.BytesToInt64(p.TxSlice[3])
					}
				} else {
					return utils.ErrInfo(fmt.Errorf("empty user_id"))
				}

				// считаем по каждому юзеру, сколько в блоке от него транзакций
				// count how many user's transactions in the block
				txCounter[userId]++

				// чтобы 1 юзер не смог прислать дос-блок размером в 10гб, который заполнит своими же транзакциями
				// for not letting to send a DOS-block (for instance 10 GB, which fill out all transactions by itself)
				if txCounter[userId] > consts.MAX_BLOCK_USER_TXS {
					p.RollbackTo(txForRollbackTo, true, false)
					return utils.ErrInfo(fmt.Errorf("max_block_user_transactions"))
				}

				// проверим, есть ли такой тип тр-ий
				// check if there is such a type of transactions
				_, ok := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
				if !ok {
					return utils.ErrInfo(fmt.Errorf("nonexistent type"))
				}

				p.TxMap = map[string][]byte{}

				// для статы
				// for statistics
				p.TxIds = append(p.TxIds, string(p.TxSlice[1]))

				MethodName := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
				log.Debug("MethodName", MethodName+"Init")
				err_ := utils.CallMethod(p, MethodName+"Init")
				if _, ok := err_.(error); ok {
					log.Debug("error: %v", err)
					p.RollbackTo(txForRollbackTo, true, true)
					return utils.ErrInfo(err_.(error))
				}

				log.Debug("MethodName", MethodName+"Front")
				err_ = utils.CallMethod(p, MethodName+"Front")
				if _, ok := err_.(error); ok {
					log.Debug("error: %v", err)
					p.RollbackTo(txForRollbackTo, true, true)
					return utils.ErrInfo(err_.(error))
				}

				// пишем хэш тр-ии в лог
				// write the hash of the transaction to logs
				err = p.InsertInLogTx(transactionBinaryDataFull, utils.BytesToInt64(p.TxMap["time"]))
				if err != nil {
					return utils.ErrInfo(err)
				}

				if len(p.BinaryData) == 0 {
					break
				}
			}
		}
	} else {

		// Оперативные транзакции
		// Operative transactions
		MethodName := consts.TxTypes[p.dataType]
		log.Debug("MethodName", MethodName+"Init")
		err_ := utils.CallMethod(p, MethodName+"Init")
		if _, ok := err_.(error); ok {
			log.Error("%v", utils.ErrInfo(err_.(error)))
			return utils.ErrInfo(err_.(error))
		}

		log.Debug("MethodName", MethodName+"Front")
		err_ = utils.CallMethod(p, MethodName+"Front")
		if _, ok := err_.(error); ok {
			log.Error("%v", utils.ErrInfo(err_.(error)))
			return utils.ErrInfo(err_.(error))
		}

		// пишем хэш тр-ии в лог
		// write the hash of the transaction to logs
		err = p.InsertInLogTx(transactionBinaryDataFull, utils.BytesToInt64(p.TxMap["time"]))
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	return nil
}
