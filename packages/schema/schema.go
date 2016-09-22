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

package schema

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("schema")

type Recmap map[string]interface{}
type Recmapi map[int]interface{}
type Recmap2 map[string]string
type SchemaStruct struct {
	*utils.DCDB
	DbType       string
	PrefixUserId int
	S            Recmap
	OnlyPrint    bool
	AddColumn    bool
	ChangeType   bool
}

func (schema *SchemaStruct) GetSchema() {

	s := make(Recmap)
	s1 := make(Recmap)
	s2 := make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('dlt_transactions_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "sender_wallet_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "recipient_wallet_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "recipient_wallet_address", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "amount", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "commission", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время, когда транзакцию создал юзер"}
	s2[7] = map[string]string{"name": "comment", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "block_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Блок, в котором данная транзакция была запечатана. При откате блока все транзакции с таким block_id будут удалены"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["dlt_transactions"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('my_keys_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "add_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "для удаления старых my_pending"}
	s2[2] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Нужно для поиска в users"}
	s2[4] = map[string]string{"name": "private_key", "mysql": "varchar(3096) NOT NULL DEFAULT ''", "sqlite": "varchar(3096) NOT NULL DEFAULT ''", "postgresql": "varchar(3096) NOT NULL DEFAULT ''", "comment": "Хранят те, кто не боятся"}
	s2[5] = map[string]string{"name": "password_hash", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": "Хранят те, кто не боятся"}
	s2[6] = map[string]string{"name": "status", "mysql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "comment": ""}
	s2[7] = map[string]string{"name": "my_time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время создания записи"}
	s2[8] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время из блока"}
	s2[9] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для откатов и определения крайнего"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Ключи для авторизации юзера. Используем крайний"
	s["my_keys"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('my_node_keys_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "add_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "для удаления старых my_pending"}
	s2[2] = map[string]string{"name": "public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "private_key", "mysql": "varchar(3096) NOT NULL DEFAULT ''", "sqlite": "varchar(3096) NOT NULL DEFAULT ''", "postgresql": "varchar(3096) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "status", "mysql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "comment": ""}
	s2[5] = map[string]string{"name": "my_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Время создания записи"}
	s2[6] = map[string]string{"name": "time", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["my_node_keys"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "type", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "wallet_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "citizen_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "error", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = "Для удобства незарегенных юзеров на пуле. Показываем им статус их тр-ий"
	s["transactions_status"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "block_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "good", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "bad", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"block_id"}
	s1["comment"] = "Результаты сверки имеющегося у нас блока с блоками у случайных нодов"
	s["confirmations"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш от полного заголовка блока (new_block_id,prev_block_hash,merkle_root,time,user_id,level). Используется как PREV_BLOCK_HASH"}
	s2[2] = map[string]string{"name": "data", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "state_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "wallet_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "tx", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "cur_0l_miner_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Майнер, который должен был сгенерить блок на 0-м уровне. Для отладки"}
	s2[8] = map[string]string{"name": "max_miner_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Макс. miner_id на момент, когда был записан этот блок. Для отладки"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["comment"] = "Главная таблица. Хранит цепочку блоков"
	s["block_chain"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "tinyint(3) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "smallint  NOT NULL  default nextval('currency_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "char(3) NOT NULL DEFAULT ''", "sqlite": "char(3) NOT NULL DEFAULT ''", "postgresql": "char(3) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "full_name", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "rb_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["currency"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "account_id", "mysql": "tinyint(3) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "smallint  NOT NULL  default nextval('currency_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "char(3) NOT NULL DEFAULT ''", "sqlite": "char(3) NOT NULL DEFAULT ''", "postgresql": "char(3) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "full_name", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "rb_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"account_id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["ea_account"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "name", "mysql": "varchar(255) NOT NULL DEFAULT '0.0.1'", "sqlite": "varchar(255) NOT NULL DEFAULT '0.0.1'", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "columns", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	//	s1["PRIMARY"] = []string{"id"}
	s1["comment"] = ""
	s["ea_state_tables"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш от полного заголовка блока (new_block_id,prev_block_hash,merkle_root,time,user_id,level). Используется как prev_hash"}
	s2[1] = map[string]string{"name": "block_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "state_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "wallet_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время создания блока"}
	s2[5] = map[string]string{"name": "level", "mysql": "tinyint(4) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(4)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": "На каком уровне был сгенерирован блок"}
	s2[6] = map[string]string{"name": "current_version", "mysql": "varchar(50) NOT NULL DEFAULT '0.0.1'", "sqlite": "varchar(50) NOT NULL DEFAULT '0.0.1'", "postgresql": "varchar(50) NOT NULL DEFAULT '0.0.1'", "comment": ""}
	s2[7] = map[string]string{"name": "sent", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Был ли блок отправлен нодам, указанным в nodes_connections"}
	s1["fields"] = s2
	s1["comment"] = "Текущий блок, данные из которого мы уже занесли к себе"
	s["info_block"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = "Храним данные за сутки, чтобы избежать дублей."
	s["log_transactions"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "lock_time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "script_name", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "info", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "uniq", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["UNIQ"] = []string{"uniq"}
	s1["comment"] = "Полная блокировка на поступление новых блоков/тр-ий"
	s["main_lock"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "full_node_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('full_nodes_full_node_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "host", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "wallet_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "state_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "final_delegate_wallet_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "final_delegate_state_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "rb_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"full_node_id"}
	s1["AI"] = "full_node_id"
	s1["comment"] = ""
	s["full_nodes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('rb_full_nodes_rb_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "full_nodes_wallet_json", "mysql": "varbinary(1024) NOT NULL DEFAULT ''", "sqlite": "varbinary(1024) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[3] = map[string]string{"name": "prev_rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"rb_id"}
	s1["AI"] = "rb_id"
	s1["comment"] = ""
	s["rb_full_nodes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"rb_id"}
	s1["AI"] = "rb_id"
	s1["comment"] = ""
	s["upd_full_nodes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('rb_upd_full_nodes_rb_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[3] = map[string]string{"name": "prev_rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"rb_id"}
	s1["AI"] = "rb_id"
	s1["comment"] = ""
	s["rb_upd_full_nodes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "full_node_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = ""
	s["queue_blocks"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "md5 от тр-ии"}
	s2[1] = map[string]string{"name": "data", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "_tmp_node_user_id", "mysql": "VARCHAR(255) DEFAULT ''", "sqlite": "VARCHAR(255) DEFAULT ''", "postgresql": "VARCHAR(255) DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = "Тр-ии, которые мы должны проверить"
	s["queue_tx"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Все хэши из этой таблы шлем тому, у кого хотим получить блок (т.е. недостающие тр-ии для составления блока)"}
	s2[1] = map[string]string{"name": "data", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": "Само тело тр-ии"}
	s2[2] = map[string]string{"name": "verified", "mysql": "tinyint(1) NOT NULL DEFAULT '1'", "sqlite": "tinyint(1) NOT NULL DEFAULT '1'", "postgresql": "smallint NOT NULL DEFAULT '1'", "comment": "Оставшиеся после прихода нового блока тр-ии отмечаются как \"непроверенные\" и их нужно проверять по новой"}
	s2[3] = map[string]string{"name": "used", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "После того как попадют в блок, ставим 1, а те, у которых уже стояло 1 - удаляем"}
	s2[4] = map[string]string{"name": "high_rate", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "1 - админские, 0 - другие"}
	s2[5] = map[string]string{"name": "for_self_use", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "для new_pct(pct_generator), т.к. эта тр-ия валидна только вместе с блоком, который сгенерил тот, кто сгенерил эту тр-ию"}
	s2[6] = map[string]string{"name": "type", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Тип тр-ии. Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера"}
	s2[7] = map[string]string{"name": "wallet_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера"}
	s2[8] = map[string]string{"name": "citizen_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера"}
	s2[9] = map[string]string{"name": "third_var", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для исключения пересения в одном блоке удаления обещанной суммы и запроса на её обмен на DC. И для исключения голосования за один и тот же объект одним и тем же юзеров и одном блоке"}
	s2[10] = map[string]string{"name": "counter", "mysql": "tinyint(3) NOT NULL DEFAULT '0'", "sqlite": "tinyint(3) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Чтобы избежать зацикливания при проверке тр-ии: verified=1, новый блок, verified=0. При достижении 10-и - удаляем тр-ию "}
	s2[11] = map[string]string{"name": "sent", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Была отправлена нодам, указанным в nodes_connections"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = "Все незанесенные в блок тр-ии, которые у нас есть"
	s["transactions"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "wallet_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('dlt_wallets_wallet_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "address", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "public_key_0", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Открытый ключ которым проверяются все транзакции от юзера"}
	s2[3] = map[string]string{"name": "public_key_1", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "2-й ключ, если есть"}
	s2[4] = map[string]string{"name": "public_key_2", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "3-й ключ, если есть"}
	s2[5] = map[string]string{"name": "node_public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "amount", "mysql": "decimal(30) NOT NULL DEFAULT '0'", "sqlite": "decimal(30) NOT NULL DEFAULT '0'", "postgresql": "decimal(30) NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "host", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "addressVote", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[9] = map[string]string{"name": "rb_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"wallet_id"}
	s1["AI"] = "wallet_id"
	s1["comment"] = ""
	s["dlt_wallets"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "citizen_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('ea_citizens_citizen_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "public_key_0", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "The public key that verifies all transactions from the user"}
	s2[2] = map[string]string{"name": "public_key_1", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "2nd key, if there is"}
	s2[3] = map[string]string{"name": "public_key_2", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "3rd key, if there is"}
	s2[4] = map[string]string{"name": "block_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "rb_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"citizen_id"}
	s1["AI"] = "citizen_id"
	s1["comment"] = ""
	s["ea_citizens"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "request_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('ea_citizenship_requests_request_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "dlt_wallet_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"citizen_id"}
	s1["AI"] = "request_id"
	s1["comment"] = ""
	s["ea_citizenship_requests"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "citizen_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('ea_citizens_private_citizen_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "main_citizen_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "name", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "fields", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "JSON of request fields"}
	s2[4] = map[string]string{"name": "binary", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"citizen_id"}
	s1["AI"] = "citizen_id"
	s1["comment"] = ""
	s["ea_citizens_private"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('ea_citizenship_requests_private_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "request_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "INTEGER NOT NULL", "postgresql": "bigint  NOT NULL  default '0'", "comment": ""}
	s2[2] = map[string]string{"name": "fields", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "JSON of request fields"}
	s2[3] = map[string]string{"name": "binary", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "public", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "approved", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "INTEGER NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["ea_citizenship_requests_private"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "state_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('states_state_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "state_code", "mysql": "varchar(2) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(2) NOT NULL DEFAULT ''", "postgresql": "varchar(2) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "node_public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "delegate_wallet_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "delegate_state_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"state_id"}
	s1["AI"] = "state_id"
	s1["comment"] = ""
	s["states"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "progress", "mysql": "varchar(10) NOT NULL DEFAULT ''", "sqlite": "varchar(10) NOT NULL DEFAULT ''", "postgresql": "varchar(10) NOT NULL DEFAULT ''", "comment": "На каком шаге остановились"}
	s1["fields"] = s2
	s1["comment"] = "Используется только в момент установки"
	s["install"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "my_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Параллельно с info_block пишем и сюда. Нужно при обнулении рабочих таблиц, чтобы знать до какого блока не трогаем таблы my_"}
	s2[1] = map[string]string{"name": "dlt_wallet_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "state_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "citizen_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "bad_blocks", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Номера и sign плохих блоков. Нужно, чтобы не подцепить более длинную, но глючную цепочку блоков"}
	s2[5] = map[string]string{"name": "pool_tech_works", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "auto_reload", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Если произойдет сбой и в main_lock будет висеть запись более auto_reload секунд, тогда будет запущен сбор блоков с чистого листа"}
	s2[7] = map[string]string{"name": "setup_password", "mysql": "varchar(255)  NOT NULL DEFAULT ''", "sqlite": "varchar(255)  NOT NULL DEFAULT ''", "postgresql": "varchar(255)  NOT NULL DEFAULT ''", "comment": "После установки и после сбора блоков, появляется окно, когда кто-угодно может ввести главный ключ"}
	s2[8] = map[string]string{"name": "sqlite_db_url", "mysql": "varchar(255)  NOT NULL DEFAULT ''", "sqlite": "varchar(255)  NOT NULL DEFAULT ''", "postgresql": "varchar(255)  NOT NULL DEFAULT ''", "comment": "Если не пусто, значит качаем с сервера sqlite базу данных"}
	s2[9] = map[string]string{"name": "first_load_blockchain_url", "mysql": "varchar(255)  NOT NULL DEFAULT ''", "sqlite": "varchar(255)  NOT NULL DEFAULT ''", "postgresql": "varchar(255)  NOT NULL DEFAULT ''", "comment": ""}
	s2[10] = map[string]string{"name": "first_load_blockchain", "mysql": "enum('nodes','file','null') DEFAULT 'null'", "sqlite": "varchar(100)  DEFAULT 'null'", "postgresql": "enum('nodes','file','null') DEFAULT 'null'", "comment": ""}
	s2[11] = map[string]string{"name": "current_load_blockchain", "mysql": "enum('nodes','file','null') DEFAULT 'null'", "sqlite": "varchar(100)  DEFAULT 'null'", "postgresql": "enum('nodes','file','null') DEFAULT 'null'", "comment": "Откуда сейчас собирается база данных"}
	s2[12] = map[string]string{"name": "http_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "адрес, по которому будет висеть панель юзера.  Если это майнер, то адрес должен совпадать с my_table.http_host"}
	s2[13] = map[string]string{"name": "auto_update", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[14] = map[string]string{"name": "auto_update_url", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[15] = map[string]string{"name": "analytics_disabled", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[16] = map[string]string{"name": "stat_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["config"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "stop_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = "Сигнал демонам об остановке"
	s["stop_daemons"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "md5 от тр-ии"}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "err", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = ""
	s["incorrect_tx"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('migration_history_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "version", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "date_applied", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["migration_history"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш транзакции. Нужно для удаления данных из буфера, после того, как транзакция была обработана в блоке, либо анулирована из-за ошибок при повторной проверке"}
	s2[1] = map[string]string{"name": "del_block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Т.к. удалять нельзя из-за возможного отката блока, приходится делать delete=1, а через сутки - чистить"}
	s2[2] = map[string]string{"name": "wallet_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "amount", "mysql": "decimal(15,2) unsigned NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Может быть = 0. Номер блока, в котором была занесена запись. Если блок в процессе фронт. проверки окажется невалдиным, то просто удалим все данные по block_id"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = "Суммируем все списания, которые еще не в блоке"
	s["dlt_wallets_buffer"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('president_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "state_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "citizen_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "start_time", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["president"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('cb_head_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "state_code", "mysql": "varchar(2) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(2) NOT NULL DEFAULT ''", "postgresql": "varchar(2) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "citizen_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["cb_head"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "parameter", "mysql": "varchar(100) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "value", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "bytecode", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "conditions_parameter", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "text_description", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "last_update", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "parent", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"parameter"}
	s1["comment"] = ""
	s["ea_state_parameters"] = s1
	schema.S = s
	schema.PrintSchema()



	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "name", "mysql": "varchar(100) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "value", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "visibility", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"name"}
	s1["comment"] = ""
	s["ea_state_variables"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('ea_state_smart_laws_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "varchar(100) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "value", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "bytecode", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "conditions_parameter", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "text_description", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "last_update", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "parent", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["ea_state_smart_laws"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "name", "mysql": "varchar(100) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "text_description", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "author", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "status", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "last_update", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"name"}
	s1["comment"] = ""
	s["ea_state_laws"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('rollback_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "tx_hash", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "table_name", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "table_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["rollback_tx"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"rb_id"}
	s1["AI"] = "rb_id"
	s1["comment"] = ""
	s["upd_full_nodes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('rb_upd_full_nodes_rb_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[3] = map[string]string{"name": "prev_rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"rb_id"}
	s1["AI"] = "rb_id"
	s1["comment"] = ""
	s["rb_upd_full_nodes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('rollback_rb_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "data", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"rb_id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["rollback"] = s1
	schema.S = s
	schema.PrintSchema()

	if !schema.OnlyPrint {

		err := schema.DCDB.ExecSql(`INSERT INTO states (state_id, state_code, node_public_key, delegate_wallet_id) VALUES (1, "EA", [hex], 1)`, "a4cb8cb9d55dcb4b2fc44cf84dee6761fcf12fd1c502fce1102d60da17fe32b329227118cce092230d6f555024e3874bee5beb5587c943e142c5c11105ceba91")
		if err != nil {
			log.Error("%v", err)
		}

		err = schema.DCDB.ExecSql(`INSERT INTO ea_citizens (citizen_id, public_key_0) VALUES (1, [hex])`, "a4cb8cb9d55dcb4b2fc44cf84dee6761fcf12fd1c502fce1102d60da17fe32b329227118cce092230d6f555024e3874bee5beb5587c943e142c5c11105ceba91")
		if err != nil {
			log.Error("%v", err)
		}

		err = schema.DCDB.ExecSql(`INSERT INTO ea_state_parameters (parameter, value) VALUES ("registration_of_citizens", "president")`)
		if err != nil {
			log.Error("%v", err)
		}
		err = schema.DCDB.ExecSql(`INSERT INTO ea_state_parameters (parameter, value) VALUES ("citizen_fields", ?)`,
			`[{"name":"name", "htmlType":"textinput", "txType":"string", "title":"First Name"},
{"name":"lastname", "htmlType":"textinput", "txType":"string", "title":"Last Name"},
{"name":"birthday", "htmlType":"calendar", "txType":"string", "title":"Birthday"},
{"name":"photo", "htmlType":"file", "txType":"binary", "title":"Photo"}
]`)
		if err != nil {
			log.Error("%v", err)
		}
		err = schema.DCDB.ExecSql(`INSERT INTO ea_state_parameters (parameter, value)  VALUES ("citizen_dlt_price", "1000000")`)
		if err != nil {
			log.Error("%v", err)
		}
		err = schema.DCDB.ExecSql(`INSERT INTO ea_state_parameters (parameter, value)  VALUES ("new_state_table", ?)`, `{"name":"name", "htmlType":"textinput", "txType":"string", "title":"Name"}]`)
		if err != nil {
			log.Error("%v", err)
		}
		err = schema.DCDB.ExecSql(`INSERT INTO president (state_id, citizen_id) VALUES (1, 1)`)
		if err != nil {
			log.Error("%v", err)
		}
		/*
			// President can only be a citizen of this state
			err = schema.DCDB.ExecSql(`INSERT INTO ea_state_parameters VALUES ("president_candidate", "if (CurrentTime - citizen.start_time > 86400) && count(president.citizen_id) < 2)", "referendum")`)
			if err!=nil {
				log.Error(err)
			}
			err = schema.DCDB.ExecSql(`INSERT INTO ea_state_parameters VALUES ("change_head_cb", "president", "parliament")`)
			if err!=nil {
				log.Error(err)
			}
			err = schema.DCDB.ExecSql(`INSERT INTO ea_state_parameters VALUES ("president_period_years", "5", "parliament")`)
			if err!=nil {
				log.Error(err)
			}
			err = schema.DCDB.ExecSql(`INSERT INTO ea_state_parameters VALUES ("parliament", "unicameralism", "parliament+president")`)
			if err!=nil {
				log.Error(err)
			}
		*/
	}
}

func (schema *SchemaStruct) typePostgresql() {
	var result string
	var err error
	for table_name, v := range schema.S {
		result = ""
		primaryKey := ""
		uniqKey := ""
		AI := ""
		AI_START := "1"
		var tableSlice []string
		for k, v1 := range v.(Recmap) {
			if k == "fields" {
				for i := 0; i < len(v1.(Recmapi)); i++ {
					var enumSlice []string
					dType := v1.(Recmapi)[i].(map[string]string)["postgresql"]
					if ok, _ := regexp.MatchString(`enum`, dType); ok {
						//enum('normal','refund') NOT NULL DEFAULT 'normal'
						r, _ := regexp.Compile(`'([\w]+)'`)
						//fmt.Println(dType)
						for _, match := range r.FindAllStringSubmatch(dType, -1) {
							//fmt.Printf("==>%s\n",  match[1])
							if ok, _ := regexp.MatchString(`^([\w]+)$`, match[1]); ok {
								if !utils.InSliceString(match[1], enumSlice) {
									enumSlice = append(enumSlice, match[1])
								}
							}
						}
						name := v1.(Recmapi)[i].(map[string]string)["name"]
						result += fmt.Sprintf("DROP TYPE IF EXISTS \"%s_enum_%s\" CASCADE;\n", table_name, name)
						result += fmt.Sprintf("CREATE TYPE \"%s_enum_%s\" AS ENUM ('%s');\n", table_name, name, strings.Join(enumSlice, "','"))
					}
				}

				for i := 0; i < len(v1.(Recmapi)); i++ {
					dType := v1.(Recmapi)[i].(map[string]string)["postgresql"]
					if ok, _ := regexp.MatchString(`enum`, dType); ok {
						//NOT NULL DEFAULT 'user'
						r, _ := regexp.Compile(`^enum\(.*?\)(.*)$`)
						rest := r.FindStringSubmatch(dType)
						dType = fmt.Sprintf("%s_enum_%s %s", table_name, v1.(Recmapi)[i].(map[string]string)["name"], rest[1])
					}
					if ok, _ := regexp.MatchString(`nextval\('\[my_prefix\]`, dType); ok {
						if schema.PrefixUserId == 0 {
							dType = strings.Replace(dType, "[my_prefix]", "", -1)
						} else {
							dType = strings.Replace(dType, "[my_prefix]", utils.IntToStr(schema.PrefixUserId)+"_", -1)
						}
					}

					tableSlice = append(tableSlice, fmt.Sprintf("\"%s\" %s", v1.(Recmapi)[i].(map[string]string)["name"], dType))
				}
			} else if k == "PRIMARY" {
				primaryKey = fmt.Sprintf("ALTER TABLE ONLY \"%[1]s\" ADD CONSTRAINT %[1]s_pkey PRIMARY KEY (%[2]s);", table_name, strings.Join(v1.([]string), ","))
			} else if k == "UNIQ" {
				uniqKey = fmt.Sprintf("CREATE UNIQUE INDEX %[1]s_%[2]s ON \"%[1]s\" USING btree (%[3]s);", table_name, v1.([]string)[0], strings.Join(v1.([]string), ","))
			} else if k == "AI" {
				AI = v1.(string)
			} else if k == "AI_START" {
				AI_START = v1.(string)
			}
		}

		if len(AI) > 0 {
			result += fmt.Sprintf("DROP SEQUENCE IF EXISTS %[3]s_%[1]s_seq CASCADE;\nCREATE SEQUENCE %[3]s_%[1]s_seq START WITH %[2]s;\n", AI, AI_START, table_name)
		}

		if schema.ChangeType {
			result += fmt.Sprintf("ALTER TABLE \"%[1]s\" RENAME TO tmp;\n", table_name)
		}
		if !schema.AddColumn {
			result += fmt.Sprintf("DROP TABLE IF EXISTS \"%[1]s\"; CREATE TABLE \"%[1]s\" (\n", table_name)
		} else {
			result += fmt.Sprintf("ALTER TABLE \"%[1]s\"\n", table_name)
		}
		//fmt.Println(tableSlice)
		for i, line := range tableSlice {
			if schema.AddColumn {
				result += "ADD COLUMN "
			}
			if i == len(tableSlice)-1 {
				result += fmt.Sprintf("%s\n", line)
			} else {
				result += fmt.Sprintf("%s,\n", line)
			}
		}
		if !schema.AddColumn {
			result += fmt.Sprintln(");")
		} else {
			result += fmt.Sprintln(";")
		}

		if len(uniqKey) > 0 {
			result += fmt.Sprintln(uniqKey)
		}

		if len(AI) > 0 {
			result += fmt.Sprintf("ALTER SEQUENCE %[2]s_%[1]s_seq owned by %[2]s.%[1]s;\n", AI, table_name)
		}

		if len(primaryKey) > 0 {
			result += fmt.Sprintln(primaryKey)
		}

		if schema.ChangeType {
			result += fmt.Sprintf("INSERT INTO \"%[1]s\" SELECT * FROM tmp;\nDROP TABLE tmp;\n", table_name)
		}
		result += fmt.Sprintln("\n\n")
		if !schema.OnlyPrint {
			err = schema.DCDB.ExecSql(result)
		} else {
			fmt.Println(result)
		}
		if err != nil {
			log.Error("%v", err)
		}
	}
}

func (schema *SchemaStruct) typeSqlite() {
	var result string
	for table_name, v := range schema.S {
		log.Debug("table_name", table_name)
		result = ""
		if schema.ChangeType {
			result += fmt.Sprintf("ALTER TABLE \"%[1]s\" RENAME TO tmp;\n", table_name)
		}
		if !schema.AddColumn {
			result += fmt.Sprintf("DROP TABLE IF EXISTS \"%[1]s\"; CREATE TABLE \"%[1]s\" (\n", table_name)
		}
		//var tableComment string
		primaryKey := ""
		uniqKey := ""
		AI := ""
		var tableSlice []string
		for k, v1 := range v.(Recmap) {
			/*if k=="comment" {
				tableComment = v1.(string)
				//fmt.Println(k, v1.(string), v1)
			} else*/if k == "fields" {
				//fmt.Println(k, v1)
				//i:=0
				//end:=""
				for i := 0; i < len(v1.(Recmapi)); i++ {
					/*if i == len(v1.(Recmap)) - 1 {
						end = ""
					} else {
						end = ","
					}*/
					tableSlice = append(tableSlice, fmt.Sprintf("\"%s\" %s", v1.(Recmapi)[i].(map[string]string)["name"], v1.(Recmapi)[i].(map[string]string)["sqlite"]))
					//fmt.Println(i)
					//i++
				}
			} else if k == "PRIMARY" {
				primaryKey = fmt.Sprintf("PRIMARY KEY (`%s`)", strings.Join(v1.([]string), "`,`"))
			} else if k == "UNIQ" {
				uniqKey = fmt.Sprintf("UNIQUE (`%v`)", strings.Join(v1.([]string), "`,`"))
			} else if k == "AI" {
				AI = v1.(string)
			}
		}
		if len(uniqKey) > 0 {
			tableSlice = append(tableSlice, uniqKey)
			//fmt.Printf("%s,\n", uniqKey)
		}
		if len(primaryKey) > 0 && len(AI) == 0 {
			tableSlice = append(tableSlice, primaryKey)
			//fmt.Printf("%s\n", primaryKey)
		}
		//fmt.Println(tableSlice)
		for i, line := range tableSlice {
			if schema.AddColumn {
				result += fmt.Sprintf("ALTER TABLE \"%[1]s\" ", table_name) + " ADD COLUMN "
				result += fmt.Sprintf("%s;\n", line)
			} else {
				if i == len(tableSlice)-1 {
					result += fmt.Sprintf("%s\n", line)
				} else {
					result += fmt.Sprintf("%s,\n", line)
				}
			}
		}
		if !schema.AddColumn {
			result += fmt.Sprintln(");\n\n")
		}

		if schema.ChangeType {
			result += fmt.Sprintf("INSERT INTO \"%[1]s\" SELECT * FROM tmp;\nDROP TABLE tmp;\n", table_name)
		}

		//log.Println(result)
		if !schema.OnlyPrint {
			log.Debug("result", result)
			err := schema.DCDB.ExecSql(result)
			if err != nil {
				log.Error("%v", err)
			}
		} else {
			fmt.Println(result)
		}

	}
}

func (schema *SchemaStruct) PrintSchema() {
	switch schema.DbType {
	case "sqlite":
		schema.typeSqlite()
	case "postgresql":
		schema.typePostgresql()
	}
}
