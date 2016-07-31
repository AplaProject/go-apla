package schema

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
	"regexp"
	"strings"
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
	OnlyPrint bool
	AddColumn bool
	ChangeType bool
}

/*
В самом начале разработки dcoin-а таблицы log_ использовались для логирования, потом я их стал использовать для откатов, но название log_ так и осталось
*/

func (schema *SchemaStruct) GetSchema() {

	s := make(Recmap)
	s1 := make(Recmap)
	s2 := make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('[my_prefix]my_cf_funding_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "from_user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "project_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "amount", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "del_block_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Блок, в котором данная транзакция была отменена"}
	s2[5] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время, когда транзакцию создал юзер"}
	s2[6] = map[string]string{"name": "block_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Блок, в котором данная транзакция была запечатана. При откате блока все транзакции с таким block_id будут удалены"}
	s2[7] = map[string]string{"name": "comment", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "comment_status", "mysql": "enum('encrypted','decrypted') NOT NULL DEFAULT 'decrypted'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'decrypted'", "postgresql": "enum('encrypted','decrypted') NOT NULL DEFAULT 'decrypted'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Нужно чтобы автор проекта мог узнать, кому какие товары отправлять"
	s["[my_prefix]my_cf_funding"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "type", "mysql": "enum('promised_amount','miner','sn', 'null') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100) ", "postgresql": "enum('promised_amount','miner','sn', 'null') NOT NULL DEFAULT 'null'", "comment": ""}
	s2[1] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["[my_prefix]my_tasks"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(10) NOT NULL DEFAULT '0'", "sqlite": "int(10) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "add_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "для удаления старых my_pending"}
	s2[2] = map[string]string{"name": "public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Нужен просто чтобы опознать в блоке зареганного юзера и отметить approved"}
	s2[3] = map[string]string{"name": "private_key", "mysql": "varchar(3096) NOT NULL DEFAULT ''", "sqlite": "varchar(3096) NOT NULL DEFAULT ''", "postgresql": "varchar(3096) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "status", "mysql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = "Чтобы после генерации нового юзера не потерять его приватный ключ можно сохранить его тут"
	s["[my_prefix]my_new_users"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('[my_prefix]my_admin_messages_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "add_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "для удаления старых my_pending"}
	s2[2] = map[string]string{"name": "parent_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "subject", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Появляется после расшифровки"}
	s2[4] = map[string]string{"name": "message", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "message_type", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "0-баг-репорты"}
	s2[6] = map[string]string{"name": "message_subtype", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "encrypted", "mysql": "blob NOT NULL DEFAULT ''", "sqlite": "blob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "decrypted", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "status", "mysql": "enum('approved','my_pending') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('approved','my_pending') NOT NULL DEFAULT 'my_pending'", "comment": ""}
	s2[10] = map[string]string{"name": "type", "mysql": "enum('from_admin','to_admin') NOT NULL DEFAULT 'to_admin'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'to_admin'", "postgresql": "enum('from_admin','to_admin') NOT NULL DEFAULT 'to_admin'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Общение с админом, баг-репорты и пр."
	s["[my_prefix]my_admin_messages"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "tinyint(3) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "smallint  NOT NULL  default nextval('[my_prefix]my_promised_amount_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "add_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "для удаления старых my_pending"}
	s2[2] = map[string]string{"name": "amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Просто показываем, какие данные еще не попали в блоки. Те, что уже попали тут удалены"
	s["[my_prefix]my_promised_amount"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('[my_prefix]my_cash_requests_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "add_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "для удаления старых my_pending"}
	s2[2] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время попадания в блок"}
	s2[3] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "to_user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "comment", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "comment_status", "mysql": "enum('encrypted','decrypted') NOT NULL DEFAULT 'decrypted'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'decrypted'", "postgresql": "enum('encrypted','decrypted') NOT NULL DEFAULT 'decrypted'", "comment": ""}
	s2[9] = map[string]string{"name": "code", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": "Секретный код, который нужно передать тому, кто отдает фиат"}
	s2[10] = map[string]string{"name": "hash_code", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": ""}
	s2[11] = map[string]string{"name": "status", "mysql": "enum('my_pending','pending','approved','rejected') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('my_pending','pending','approved','rejected') NOT NULL DEFAULT 'my_pending'", "comment": ""}
	s2[12] = map[string]string{"name": "cash_request_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["[my_prefix]my_cash_requests"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('[my_prefix]my_dc_transactions_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "status", "mysql": "enum('pending','approved') NOT NULL DEFAULT 'approved'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'approved'", "postgresql": "enum('pending','approved') NOT NULL DEFAULT 'approved'", "comment": "pending - только при отправки DC с нашего кошелька, т.к. нужно показать юзеру, что запрос принят"}
	s2[2] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Уведомления по sms и email"}
	s2[3] = map[string]string{"name": "type", "mysql": "enum('null','cash_request','from_mining_id','from_repaid','from_user','node_commission','system_commission','referral','cf_project','cf_project_refund','loan_payment','arbitrator_commission', 'money_back') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100) ", "postgresql": "enum('null','cash_request','from_mining_id','from_repaid','from_user','node_commission','system_commission','referral','cf_project','cf_project_refund','loan_payment','arbitrator_commission', 'money_back') NOT NULL DEFAULT 'null'", "comment": ""}
	s2[4] = map[string]string{"name": "type_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "to_user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Тут не всегда user_id, может быть ID проекта или cash_request"}
	s2[6] = map[string]string{"name": "amount", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "commission", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "del_block_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Блок, в котором данная транзакция была отменена"}
	s2[9] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время, когда транзакцию создал юзер"}
	s2[10] = map[string]string{"name": "block_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Блок, в котором данная транзакция была запечатана. При откате блока все транзакции с таким block_id будут удалены"}
	s2[11] = map[string]string{"name": "currency_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[12] = map[string]string{"name": "comment", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Если это перевод средств между юзерами или это комиссия, то тут будет расшифрованный комментарий"}
	s2[13] = map[string]string{"name": "comment_status", "mysql": "enum('encrypted','decrypted') NOT NULL DEFAULT 'decrypted'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'decrypted'", "postgresql": "enum('encrypted','decrypted') NOT NULL DEFAULT 'decrypted'", "comment": ""}
	s2[14] = map[string]string{"name": "merchant_checked", "mysql": "tinyint(1) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(1)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[15] = map[string]string{"name": "exchange_checked", "mysql": "tinyint(1) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(1)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Нужно только для отчетов, которые показываются юзеру"
	s["[my_prefix]my_dc_transactions"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "smallint(6) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "smallint NOT NULL  default nextval('[my_prefix]my_holidays_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "add_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "для удаления старых my_pending"}
	s2[2] = map[string]string{"name": "start_time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "end_time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["[my_prefix]my_holidays"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('[my_prefix]my_keys_id_seq')", "comment": ""}
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
	s["[my_prefix]my_keys"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('[my_prefix]my_node_keys_id_seq')", "comment": ""}
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
	s["[my_prefix]my_node_keys"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "name", "mysql": "varchar(200) NOT NULL DEFAULT ''", "sqlite": "varchar(200) NOT NULL DEFAULT ''", "postgresql": "varchar(200) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "email", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "sms", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "mobile", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "sort", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "important", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"name"}
	s1["comment"] = ""
	s["[my_prefix]my_notifications"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "last_voting", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время последнего голосования"}
	s2[1] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Уведомление о том, что со времени последнего голоса прошло более 2 недель"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"last_voting"}
	s1["comment"] = "Нужно только для отсылки уведомлений, что пора голосовать"
	s["[my_prefix]my_complex_votes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "miner_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "status", "mysql": "enum('waiting_accept_new_key','waiting_set_new_key','bad_key','my_pending','miner','user','passive_miner','suspended_miner') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('waiting_accept_new_key','waiting_set_new_key','bad_key','my_pending','miner','user','passive_miner','suspended_miner') NOT NULL DEFAULT 'my_pending'", "comment": "bad_key - это когда юзер зарегался по чужому ключу, который нашел в паблике, либо если указал старый ключ вместо нового"}
	s2[3] = map[string]string{"name": "race", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Раса. От 1 до 3"}
	s2[4] = map[string]string{"name": "country", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": "Используется только локально для проверки майнеров из нужной страны"}
	s2[5] = map[string]string{"name": "notification_status", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Уведомления. При смене статуса обнуляется"}
	s2[6] = map[string]string{"name": "mail_code", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "login_code", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для подписания при авторизации"}
	s2[8] = map[string]string{"name": "email", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[9] = map[string]string{"name": "notification_email", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "http_host", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Хост юзера, по которому он доступен из вне"}
	s2[11] = map[string]string{"name": "tcp_host", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Хост юзера, по которому он доступен из вне"}
	s2[12] = map[string]string{"name": "host_status", "mysql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "comment": ""}
	s2[13] = map[string]string{"name": "geolocation", "mysql": "varchar(200) NOT NULL DEFAULT ''", "sqlite": "varchar(200) NOT NULL DEFAULT ''", "postgresql": "varchar(200) NOT NULL DEFAULT ''", "comment": "Текущее местонахождение майнера"}
	s2[14] = map[string]string{"name": "geolocation_status", "mysql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "comment": ""}
	s2[15] = map[string]string{"name": "location_country", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[16] = map[string]string{"name": "invite", "mysql": "char(128) NOT NULL DEFAULT ''", "sqlite": "char(128) NOT NULL DEFAULT ''", "postgresql": "char(128) NOT NULL DEFAULT ''", "comment": ""}
	s2[17] = map[string]string{"name": "face_coords", "mysql": "varchar(1024) NOT NULL DEFAULT ''", "sqlite": "varchar(1024) NOT NULL DEFAULT ''", "postgresql": "varchar(1024) NOT NULL DEFAULT ''", "comment": "Точки, которе юзер нанес на свое фото"}
	s2[18] = map[string]string{"name": "node_voting_send_request", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Когда мы отправили запрос в DC-сеть на присвоение нам статуса \"майнер\""}
	s2[19] = map[string]string{"name": "profile_coords", "mysql": "varchar(1024) NOT NULL DEFAULT ''", "sqlite": "varchar(1024) NOT NULL DEFAULT ''", "postgresql": "varchar(1024) NOT NULL DEFAULT ''", "comment": "Точки, которе юзер нанес на свое фото"}
	s2[20] = map[string]string{"name": "video_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Видео, где показывается лицо юзера"}
	s2[21] = map[string]string{"name": "video_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[22] = map[string]string{"name": "lang", "mysql": "char(2) NOT NULL DEFAULT ''", "sqlite": "char(2) NOT NULL DEFAULT ''", "postgresql": "char(2) NOT NULL DEFAULT ''", "comment": "Запоминаем язык для юзера"}
	s2[23] = map[string]string{"name": "use_smtp", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[24] = map[string]string{"name": "smtp_server", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[25] = map[string]string{"name": "smtp_port", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[26] = map[string]string{"name": "smtp_ssl", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[27] = map[string]string{"name": "smtp_auth", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[28] = map[string]string{"name": "smtp_username", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[29] = map[string]string{"name": "smtp_password", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[30] = map[string]string{"name": "miner_pct_id", "mysql": "smallint(5) NOT NULL DEFAULT '0'", "sqlite": "smallint(5) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[31] = map[string]string{"name": "user_pct_id", "mysql": "smallint(5) NOT NULL DEFAULT '0'", "sqlite": "smallint(5) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[32] = map[string]string{"name": "repaid_pct_id", "mysql": "smallint(5) NOT NULL DEFAULT '0'", "sqlite": "smallint(5) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[33] = map[string]string{"name": "api_token_hash", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": ""}
	s2[34] = map[string]string{"name": "sms_http_get_request", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[35] = map[string]string{"name": "notification_sms_http_get_request", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[36] = map[string]string{"name": "show_sign_data", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Если 0, тогда не показываем данные для подписи, если у юзера только один праймари ключ"}
	s2[37] = map[string]string{"name": "show_map", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[38] = map[string]string{"name": "show_progress_bar", "mysql": "tinyint(1) NOT NULL DEFAULT '1'", "sqlite": "tinyint(1) NOT NULL DEFAULT '1'", "postgresql": "smallint NOT NULL DEFAULT '1'", "comment": ""}
	s2[39] = map[string]string{"name": "hide_first_promised_amount", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[40] = map[string]string{"name": "hide_first_commission", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[41] = map[string]string{"name": "shop_secret_key", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Секреный ключ, который используется в демоне shop в хэше для проверки данных на callback-е "}
	s2[42] = map[string]string{"name": "shop_callback_url", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Куда демон shop будет отстукивать"}
	s2[43] = map[string]string{"name": "key_password", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Пароль для ключа. Храним тут временно"}
	s2[44] = map[string]string{"name": "first_select", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Выбор юзера при первом заходе в кошель. Аноним или майнер"}
	s2[45] = map[string]string{"name": "tcp_listening", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Нужно чтобы листенинг не запускался у тех, кто удаленно зарегался на пуле"}
	s2[46] = map[string]string{"name": "pool_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[47] = map[string]string{"name": "uniq", "mysql": "tinyint(1) NOT NULL DEFAULT '1'", "sqlite": "tinyint(1) NOT NULL DEFAULT '1'", "postgresql": "smallint NOT NULL DEFAULT '1'", "comment": ""}
	s1["fields"] = s2
	s1["UNIQ"] = []string{"uniq"}
	s1["comment"] = ""
	s["[my_prefix]my_table"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "pct", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "min", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "max", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = "Каждый майнер определяет, какая комиссия с тр-ий будет доставаться ему, если он будет генерить блок"
	s["[my_prefix]my_commission"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "type", "mysql": "enum('null','miner','promised_amount','arbitrator','seller','sn_user') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'null'", "postgresql": "enum('null','miner','promised_amount','arbitrator','seller', 'sn_user') NOT NULL DEFAULT 'null'", "comment": ""}
	s2[1] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "comment", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "comment_status", "mysql": "enum('encrypted','decrypted') NOT NULL DEFAULT 'decrypted'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'decrypted'", "postgresql": "enum('encrypted','decrypted') NOT NULL DEFAULT 'decrypted'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = "Чтобы было проще понять причину отказов при апгрейде акка или добавлении обещанной суммы. Также сюда пишутся комменты арбитрам и продавцам, когда покупатели запрашивают манибек"
	s["[my_prefix]my_comments"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_arbitrator_conditions_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "conditions", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "prev_log_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_arbitrator_conditions"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "conditions", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = ""
	s["arbitrator_conditions"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_ca_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_ca"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_seller_hold_back_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_seller_hold_back"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_arbitrator_conditions_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_arbitrator_conditions"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_arbitration_trust_list_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_arbitration_trust_list"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_money_back_request_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_money_back_request"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "arbitrator_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = "Список арбитров, кому доверяют юзеры"
	s["arbitration_trust_list"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_arbitration_trust_list_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "arbitration_trust_list", "mysql": "varchar(512) NOT NULL DEFAULT ''", "sqlite": "varchar(512) NOT NULL DEFAULT ''", "postgresql": "varchar(512) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "prev_log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_arbitration_trust_list"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('orders_id_seq')", "comment": "Этот ID указывается в тр-ии при запросе манибека"}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Время блока, в котором запечатана данная сделка"}
	s2[2] = map[string]string{"name": "buyer", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "user_id покупателя"}
	s2[3] = map[string]string{"name": "seller", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "user_id продавца"}
	s2[4] = map[string]string{"name": "arbitrator0", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "user_id арбитра 0"}
	s2[5] = map[string]string{"name": "arbitrator1", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "user_id арбитра 1"}
	s2[6] = map[string]string{"name": "arbitrator2", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "user_id арбитра 2"}
	s2[7] = map[string]string{"name": "arbitrator3", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "user_id арбитра 3"}
	s2[8] = map[string]string{"name": "arbitrator4", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "user_id арбитра 4"}
	s2[9] = map[string]string{"name": "amount", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": "Сумма сделки"}
	s2[10] = map[string]string{"name": "hold_back_amount", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": "Сумма, которая замораживается на счету продавца. % для новых сделок задается в users.hold_back_pct"}
	s2[11] = map[string]string{"name": "currency_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[12] = map[string]string{"name": "end_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Время окончения возможности сделать манибек через арбитра. Может быть однократно увеличино арбитром. Используется для подсчета, сколько на данный момент времени есть активных сделок, чтобы посчитать от них 10% и не дать их списать со счета продавца"}
	s2[13] = map[string]string{"name": "end_time_changed", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Если арбитр изменил время окончания, то тут будет 1, чтобы нельзя было изменить повторно"}
	s2[14] = map[string]string{"name": "status", "mysql": "enum('normal','refund') NOT NULL DEFAULT 'normal'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'normal'", "postgresql": "enum('normal','refund') NOT NULL DEFAULT 'normal'", "comment": "Чтобы арбитр мог понять, что покупатель сделал запрос манибека. Когда юзер шлет тр-ию с запросом манибека, тут меняется статус на refund"}
	s2[15] = map[string]string{"name": "refund", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": "Сумма к возврату, которую определил арбитр. Она не может быть больше, чем сумма сделки.  Повторно отправить транзакцию с  манибеком  не даем, дабы не захламлять тр-ми сеть"}
	s2[16] = map[string]string{"name": "refund_arbitrator_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для статы. ID арбитра, который сделал манибек юзеру"}
	s2[17] = map[string]string{"name": "arbitrator_refund_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для статы. Время, когда арбитр сделал манибек"}
	s2[18] = map[string]string{"name": "voluntary_refund", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": "Сумма, которую продавец добровольно вернул покупателю.  Повторно отправить транзакцию с добровольным манибеком  не даем, дабы не захламлять тр-ми сеть. Если сумма не вся, то арбитр может довести процесс до конца, если посчитает нужным"}
	s2[19] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[20] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["orders"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_orders_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "end_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "status", "mysql": "enum('normal','refund') NOT NULL DEFAULT 'normal'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'normal'", "postgresql": "enum('normal','refund') NOT NULL DEFAULT 'normal'", "comment": ""}
	s2[3] = map[string]string{"name": "end_time_changed", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "refund", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "arbitrator_refund_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "refund_arbitrator_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "voluntary_refund", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "remaining_refund", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "prev_log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_orders"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "referral", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "amount", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "currency_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = "Для вывода статы по рефам"
	s["referral_stats"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "type", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "error", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
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
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_key_request_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_key_request"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_key_active_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_key_active"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["admin"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_admin_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "prev_log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_admin"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "admin_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = ""
	s["votes_admin"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_votes_admin_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "admin_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "prev_log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_votes_admin"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_new_credit_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_new_credit"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_creditor_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_creditor"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_repayment_credit_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_repayment_credit"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_credit_part_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_credit_part"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('credits_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "del_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "amount", "mysql": "decimal(10,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(10,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(10,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "from_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "to_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "currency_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "pct", "mysql": "decimal(5,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(5,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(5,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "monthly_payment", "mysql": "decimal(10,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(10,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(10,2) NOT NULL DEFAULT '0'", "comment": "Ежемесячный платеж по кредиту. Пока не используется"}
	s2[9] = map[string]string{"name": "last_payment", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Время последнего платежа по кредиту. Пока не используется"}
	s2[10] = map[string]string{"name": "surety_1", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Поручитель 1. Пока не используется"}
	s2[11] = map[string]string{"name": "surety_2", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Поручитель 2. Пока не используется"}
	s2[12] = map[string]string{"name": "surety_3", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Поручитель 3. Пока не используется"}
	s2[13] = map[string]string{"name": "surety_4", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Поручитель 4. Пока не используется"}
	s2[14] = map[string]string{"name": "surety_5", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Поручитель 5. Пока не используется"}
	s2[15] = map[string]string{"name": "tx_hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[16] = map[string]string{"name": "tx_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[17] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["credits"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_credits_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "amount", "mysql": "decimal(10,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(10,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(10,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "to_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "last_payment", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "tx_hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "tx_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_credits"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "project_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "ps1", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "ps2", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "ps3", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "ps4", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "ps5", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "ps6", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "ps7", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "ps8", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"project_id"}
	s1["comment"] = "Каждому CF-проекту вручную указывается платежные системы"
	s["cf_projects_ps"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "project_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"project_id"}
	s1["comment"] = "Какие проекты не выводим в CF-каталоге"
	s["cf_blacklist"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "email", "mysql": "varchar(200) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(200) NOT NULL DEFAULT ''", "postgresql": "varchar(200) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"email"}
	s1["comment"] = ""
	s["pool_waiting_list"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('cf_lang_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "varchar(200) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(200) NOT NULL DEFAULT ''", "postgresql": "varchar(200) NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["cf_lang"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_user_avatar_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_user_avatar"] = s1

	schema.S = s
	schema.PrintSchema()
	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_user_upgrade_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_user_upgrade"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_cf_comments_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_cf_comments"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_new_cf_project_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_new_cf_project"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_cf_project_data_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_cf_project_data"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_cf_send_dc_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_cf_send_dc"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('cf_comments_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "project_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "lang_id", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "comment", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для того, чтобы можно было отсчитать время до размещения следующего коммента"}
	s2[6] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для откатов"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["cf_comments"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('cf_funding_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "project_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "amount", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "amount_backup", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "currency_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "DC растут с юзерским %"}
	s2[7] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для откатов"}
	s2[8] = map[string]string{"name": "del_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Фундер передумал и до завершения проекта вернул деньги"}
	s2[9] = map[string]string{"name": "checked", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Для определения по крону cf_project.funding и cf_project.funding"}
	s2[10] = map[string]string{"name": "del_checked", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Для определения по крону cf_project.funding и cf_project.funding"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["cf_funding"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('cf_currency_id_seq')", "comment": "ID идет от 1000, чтобы id CF-валют не пересекались с DC-валютами"}
	s2[1] = map[string]string{"name": "name", "mysql": "char(7) NOT NULL DEFAULT ''", "sqlite": "char(7) NOT NULL DEFAULT ''", "postgresql": "char(7) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "project_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["AI_START"] = "1000"
	s1["comment"] = ""
	s["cf_currency"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('cf_projects_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "amount", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "funding", "mysql": "decimal(15,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2) NOT NULL DEFAULT '0'", "comment": "Получаем в кроне. Сколько собрано средств. Нужно для вывода проектов в каталоге, чтобы не дергать cf_funding"}
	s2[5] = map[string]string{"name": "funders", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Получаем в кроне. Кол-во инвесторов. Нужно для вывода проектов в каталоге, чтобы не дергать cf_funding"}
	s2[6] = map[string]string{"name": "project_currency_name", "mysql": "char(7) NOT NULL DEFAULT ''", "sqlite": "char(7) NOT NULL DEFAULT ''", "postgresql": "char(7) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "end_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "latitude", "mysql": "decimal(8,5) NOT NULL DEFAULT '0'", "sqlite": "decimal(8,5) NOT NULL DEFAULT '0'", "postgresql": "decimal(8,5) NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "longitude", "mysql": "decimal(8,5) NOT NULL DEFAULT '0'", "sqlite": "decimal(8,5) NOT NULL DEFAULT '0'", "postgresql": "decimal(8,5) NOT NULL DEFAULT '0'", "comment": ""}
	s2[11] = map[string]string{"name": "country", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[12] = map[string]string{"name": "city", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[13] = map[string]string{"name": "category_id", "mysql": "smallint(6) NOT NULL DEFAULT '0'", "sqlite": "smallint(6) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[14] = map[string]string{"name": "close_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Чтобы знать, когда проект завершился и можно было бы удалить старые данные из cf_funding. Также используется для определения статус проекта - открыт/закрыт"}
	s2[15] = map[string]string{"name": "del_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Проект был закрыт автором, а средства возвращены инвесторам"}
	s2[16] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для откатов"}
	s2[17] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[18] = map[string]string{"name": "geo_checked", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "По крону превращаем координаты в названия страны и города и отмечаем тут"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["cf_projects"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('cf_projects_data_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "hide", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "project_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "lang_id", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "blurb_img", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "head_img", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "description_img", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "picture", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": "Если нет видео, то выводится эта картинка"}
	s2[8] = map[string]string{"name": "video_type", "mysql": "varchar(10) NOT NULL DEFAULT ''", "sqlite": "varchar(10) NOT NULL DEFAULT ''", "postgresql": "varchar(10) NOT NULL DEFAULT ''", "comment": ""}
	s2[9] = map[string]string{"name": "video_url_id", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[10] = map[string]string{"name": "news_img", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[11] = map[string]string{"name": "links", "mysql": "varchar(512) NOT NULL DEFAULT ''", "sqlite": "varchar(512) NOT NULL DEFAULT ''", "postgresql": "varchar(512) NOT NULL DEFAULT ''", "comment": ""}
	s2[12] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["UNIQ"] = []string{"project_id", "lang_id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["cf_projects_data"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_cf_projects_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "category_id", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_cf_projects"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_cf_projects_data_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "hide", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "lang_id", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "blurb_img", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "head_img", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "description_img", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "picture", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "video_type", "mysql": "varchar(10) NOT NULL DEFAULT ''", "sqlite": "varchar(10) NOT NULL DEFAULT ''", "postgresql": "varchar(10) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "video_url_id", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[9] = map[string]string{"name": "news_img", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[10] = map[string]string{"name": "links", "mysql": "varchar(512) NOT NULL DEFAULT ''", "sqlite": "varchar(512) NOT NULL DEFAULT ''", "postgresql": "varchar(512) NOT NULL DEFAULT ''", "comment": ""}
	s2[11] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[12] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_cf_projects_data"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "from_user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "comment", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = "Абузы на майнеров от майнеров"
	s["abuses"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('admin_blog_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "lng", "mysql": "varchar(5) NOT NULL DEFAULT ''", "sqlite": "varchar(5) NOT NULL DEFAULT ''", "postgresql": "varchar(5) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "title", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "message", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Блог админа"
	s["admin_blog"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('alert_messages_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "close", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Юзер может закрыть сообщение и оно больше не появится"}
	s2[3] = map[string]string{"name": "message", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "json. Каждому языку свое сообщение и gen - для тех, на кого языков не хватило"}
	s2[4] = map[string]string{"name": "currency_list", "mysql": "varchar(1024) NOT NULL DEFAULT ''", "sqlite": "varchar(1024) NOT NULL DEFAULT ''", "postgresql": "varchar(1024) NOT NULL DEFAULT ''", "comment": "Для каких валют выводим сообщение. ALL - всем"}
	s2[5] = map[string]string{"name": "block_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Для откатов"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Сообщения от админа, которые выводятся в интерфейсе софта"
	s["alert_messages"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш от полного заголовка блока (new_block_id,prev_block_hash,merkle_root,time,user_id,level). Используется как PREV_BLOCK_HASH"}
	s2[2] = map[string]string{"name": "head_hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш от заголовка блока (user_id,block_id,prev_head_hash). Используется для обновления head_hash в info_block при восстановлении после вилки в upd_block_info()"}
	s2[3] = map[string]string{"name": "data", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "tx", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "cur_0l_miner_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Майнер, который должен был сгенерить блок на 0-м уровне. Для отладки"}
	s2[7] = map[string]string{"name": "max_miner_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Макс. miner_id на момент, когда был записан этот блок. Для отладки"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["comment"] = "Главная таблица. Хранит цепочку блоков"
	s["block_chain"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('cash_requests_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время создания запроса. От него отсчитываем 48 часов"}
	s2[2] = map[string]string{"name": "from_user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "to_user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "notification", "mysql": "tinyint(1) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(1)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "amount", "mysql": "decimal(13,2)  NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2)  NOT NULL DEFAULT '0'", "comment": "На эту сумму должны быть выданы наличные"}
	s2[7] = map[string]string{"name": "hash_code", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш от кода, а сам код пердается при личной встрече. "}
	s2[8] = map[string]string{"name": "status", "mysql": "enum('approved','pending') NOT NULL DEFAULT 'pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'pending'", "postgresql": "enum('approved','pending') NOT NULL DEFAULT 'pending'", "comment": "Если в блоке указан верный код для хэша, то тут будет approved. Rejected нет, т.к. можно и без него понять, что запрос невыполнен, просто посмотрев время"}
	s2[9] = map[string]string{"name": "for_repaid_del_block_id", "mysql": "int(11)  unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)   NOT NULL DEFAULT '0'", "postgresql": "int   NOT NULL DEFAULT '0'", "comment": "если больше нет for_repaid ни по одной валюте у данного юзера, то нужно проверить, нет ли у него просроченных cash_requests, которым нужно отметить for_repaid_del_block_id, чтобы cash_request_out не переводил более обещанные суммы данного юзера в for_repaid из-за просроченных cash_requests"}
	s2[10] = map[string]string{"name": "del_block_id", "mysql": "int(11)  unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)   NOT NULL DEFAULT '0'", "postgresql": "int   NOT NULL DEFAULT '0'", "comment": "Во время reduction все текущие cash_requests, т.е. по которым не прошло 2 суток удаляются"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Запросы на обмен DC на наличные"
	s["cash_requests"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "tinyint(3) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "smallint  NOT NULL  default nextval('currency_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "char(3) NOT NULL DEFAULT ''", "sqlite": "char(3) NOT NULL DEFAULT ''", "postgresql": "char(3) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "full_name", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "max_other_currencies", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": "Со сколькими валютами данная валюта может майниться"}
	s2[4] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
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
	s2[0] = map[string]string{"name": "log_id", "mysql": "int(11) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int  NOT NULL  default nextval('log_currency_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "max_other_currencies", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[3] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_currency"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "name", "mysql": "char(15) NOT NULL DEFAULT ''", "sqlite": "char(15) NOT NULL DEFAULT ''", "postgresql": "char(15) NOT NULL DEFAULT ''", "comment": "Кодовое обозначение демона"}
	s2[1] = map[string]string{"name": "script", "mysql": "char(40) NOT NULL DEFAULT ''", "sqlite": "char(40) NOT NULL DEFAULT ''", "postgresql": "char(40) NOT NULL DEFAULT ''", "comment": "Название скрипта"}
	s2[2] = map[string]string{"name": "param", "mysql": "char(5) NOT NULL DEFAULT ''", "sqlite": "char(5) NOT NULL DEFAULT ''", "postgresql": "char(5) NOT NULL DEFAULT ''", "comment": "Параметры для запуска"}
	s2[3] = map[string]string{"name": "pid", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Pid демона для детекта дублей"}
	s2[4] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Время последней активности демона"}
	s2[5] = map[string]string{"name": "first", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "memory", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "restart", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Команда демону, что нужно выйти"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"script"}
	s1["comment"] = "Демоны"
	s["daemons"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "race", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Раса. От 1 до 3"}
	s2[2] = map[string]string{"name": "country", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "version", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Версия набора точек"}
	s2[4] = map[string]string{"name": "status", "mysql": "enum('pending','used') NOT NULL DEFAULT 'pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'pending'", "postgresql": "enum('pending','used') NOT NULL DEFAULT 'pending'", "comment": "При new_miner ставим pending, при отрицательном завершении юзерского голосования - pending. used ставится только если юзерское голосование завершилось положительно"}
	s2[5] = map[string]string{"name": "f1", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": "Отрезок 1"}
	s2[6] = map[string]string{"name": "f2", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "f3", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "f4", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "f5", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "f6", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[11] = map[string]string{"name": "f7", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[12] = map[string]string{"name": "f8", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[13] = map[string]string{"name": "f9", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[14] = map[string]string{"name": "f10", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[15] = map[string]string{"name": "f11", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[16] = map[string]string{"name": "f12", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[17] = map[string]string{"name": "f13", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[18] = map[string]string{"name": "f14", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[19] = map[string]string{"name": "f15", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[20] = map[string]string{"name": "f16", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[21] = map[string]string{"name": "f17", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[22] = map[string]string{"name": "f18", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[23] = map[string]string{"name": "f19", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[24] = map[string]string{"name": "f20", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[25] = map[string]string{"name": "p1", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[26] = map[string]string{"name": "p2", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[27] = map[string]string{"name": "p3", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[28] = map[string]string{"name": "p4", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[29] = map[string]string{"name": "p5", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[30] = map[string]string{"name": "p6", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[31] = map[string]string{"name": "p7", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[32] = map[string]string{"name": "p8", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[33] = map[string]string{"name": "p9", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[34] = map[string]string{"name": "p10", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[35] = map[string]string{"name": "p11", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[36] = map[string]string{"name": "p12", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[37] = map[string]string{"name": "p13", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[38] = map[string]string{"name": "p14", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[39] = map[string]string{"name": "p15", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[40] = map[string]string{"name": "p16", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[41] = map[string]string{"name": "p17", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[42] = map[string]string{"name": "p18", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[43] = map[string]string{"name": "p19", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[44] = map[string]string{"name": "p20", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[45] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = "Точки по каждому юзеру"
	s["faces"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('holidays_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "del", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "1-удалено. нужно для отката"}
	s2[3] = map[string]string{"name": "start_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "end_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Время, в которое майнер не получает %, т.к. отдыхает"
	s["holidays"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш от полного заголовка блока (new_block_id,prev_block_hash,merkle_root,time,user_id,level). Используется как prev_hash"}
	s2[1] = map[string]string{"name": "head_hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш от заголовка блока (user_id,block_id,prev_head_hash)"}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время создания блока"}
	s2[4] = map[string]string{"name": "level", "mysql": "tinyint(4) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(4)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": "На каком уровне был сгенерирован блок"}
	s2[5] = map[string]string{"name": "current_version", "mysql": "varchar(50) NOT NULL DEFAULT '0.0.1'", "sqlite": "varchar(50) NOT NULL DEFAULT '0.0.1'", "postgresql": "varchar(50) NOT NULL DEFAULT '0.0.1'", "comment": ""}
	s2[6] = map[string]string{"name": "sent", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Был ли блок отправлен нодам, указанным в nodes_connections"}
	s1["fields"] = s2
	s1["comment"] = "Текущий блок, данные из которого мы уже занесли к себе"
	s["info_block"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('promised_amount_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "del_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "user_id", "mysql": "bigint(16) NOT NULL DEFAULT '0'", "sqlite": "bigint(16) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": "Обещанная сумма. На неё влияет reduction и она будет урезаться при обновлении max_promised_amount (очень важно на случай деноминации фиата). Если же статус = repaid, то тут храниться кол-во денег, которые майнер отдал. Нужно хранить только чтобы знать общую сумму и не превысить max_promised_amount. Для WOC  amount не нужен, т.к. WOC полностью зависит от max_promised_amount"}
	s2[4] = map[string]string{"name": "amount_backup", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": "Нужно для откатов при reduction"}
	s2[5] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "ps1", "mysql": "smallint(5) unsigned NOT NULL DEFAULT '0'", "sqlite": "smallint(5)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": "ID платежной системы, в валюте которой он готов сделать перевод в случае входящего запроса"}
	s2[7] = map[string]string{"name": "ps2", "mysql": "smallint(5) unsigned NOT NULL DEFAULT '0'", "sqlite": "smallint(5)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "ps3", "mysql": "smallint(5) unsigned NOT NULL DEFAULT '0'", "sqlite": "smallint(5)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "ps4", "mysql": "smallint(5) unsigned NOT NULL DEFAULT '0'", "sqlite": "smallint(5)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "ps5", "mysql": "smallint(5) unsigned NOT NULL DEFAULT '0'", "sqlite": "smallint(5)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[11] = map[string]string{"name": "start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Используется, когда нужно узнать, кто имеет право голосовать за данную валюту, т.е. прошло ли 60 дней с момента получения статуса miner или repaid(учитывая время со статусом miner). Изменяется при каждой смене статуса. Сущетвует только со статусом mining и repaid. Это защита от атаки клонов, когда каким-то образом 100500 майнеров прошли проверку, добавили какую-то валюту и проголосовали за reduction 90%. 90 дней - это время админу, чтобы заметить и среагировать на такую атаку"}
	s2[12] = map[string]string{"name": "status", "mysql": "enum('pending','mining','rejected','repaid','change_geo','suspended') NOT NULL DEFAULT 'pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'pending'", "postgresql": "enum('pending','mining','rejected','repaid','change_geo','suspended') NOT NULL DEFAULT 'pending'", "comment": "pending - при первом добавлении или при повтороном запросе.  change_geo ставится когда идет смена местоположения, suspended - когда админ разжаловал майнера в юзеры. TDC набегают только когда статус mining, repaid с майнерским или же юзерским % (если статус майнера = passive_miner)"}
	s2[13] = map[string]string{"name": "status_backup", "mysql": "enum('pending','mining','rejected','repaid','change_geo','suspended','null') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'null'", "postgresql": "enum('pending','mining','rejected','repaid','change_geo','suspended','null') NOT NULL DEFAULT 'null'", "comment": "Когда админ банит майнера, то в status пишется suspended, а сюда - статус из  status"}
	s2[14] = map[string]string{"name": "tdc_amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": "Набежавшая сумма за счет % роста. Пересчитывается при переводе TDC на кошелек"}
	s2[15] = map[string]string{"name": "tdc_amount_backup", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": "Нужно для откатов при reduction"}
	s2[16] = map[string]string{"name": "tdc_amount_update", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время обновления tdc_amount"}
	s2[17] = map[string]string{"name": "video_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[18] = map[string]string{"name": "video_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Если пусто, то видео берем по ID юзера.flv. На видео майнер говорит, что хочет майнить выбранную валюту"}
	s2[19] = map[string]string{"name": "votes_start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "При каждой смене местоположения начинается новое голосование. Менять местоположение можно не чаще раза в сутки"}
	s2[20] = map[string]string{"name": "votes_0", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[21] = map[string]string{"name": "votes_1", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[22] = map[string]string{"name": "woc_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Нужно для отката добавления woc"}
	s2[23] = map[string]string{"name": "cash_request_out_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Любой cash_request_out приводит к появлению данной записи у получателя запроса. Убирается она только после того, как у юзера не остается непогашенных cash_request-ов. Нужно для reduction_generator, чтобы учитывать только те обещанные суммы, которые еще не заморожены невыполенными cash_request-ами"}
	s2[24] = map[string]string{"name": "cash_request_out_time_backup", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Используется в new_reduction()"}
	s2[25] = map[string]string{"name": "cash_request_in_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Нужно для отката cash_request_in"}
	s2[26] = map[string]string{"name": "del_mining_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Нужно для отката del_promised_amount"}
	s2[27] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["promised_amount"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_promised_amount_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "del_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "amount_backup", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "status", "mysql": "enum('null','pending','mining','rejected','repaid','change_geo','suspended')  NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)   NOT NULL DEFAULT 'null'", "postgresql": "enum('null','pending','mining','rejected','repaid','change_geo','suspended')  NOT NULL DEFAULT 'null'", "comment": ""}
	s2[6] = map[string]string{"name": "status_backup", "mysql": "enum('pending','mining','rejected','repaid','change_geo','suspended','null')  NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)   NOT NULL DEFAULT 'null'", "postgresql": "enum('pending','mining','rejected','repaid','change_geo','suspended','null')  NOT NULL DEFAULT 'null'", "comment": ""}
	s2[7] = map[string]string{"name": "tdc_and_profit", "mysql": "decimal(13,2)  NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2)  NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "tdc_amount", "mysql": "decimal(13,2)  NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2)  NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "tdc_amount_update", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "video_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[11] = map[string]string{"name": "video_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Если пусто, то видео берем по ID юзера.flv. На видео майнер говорит, что хочет майнить выбранную валюту"}
	s2[12] = map[string]string{"name": "votes_start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "При каждой смене местоположения начинается новое голосование"}
	s2[13] = map[string]string{"name": "votes_0", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[14] = map[string]string{"name": "votes_1", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[15] = map[string]string{"name": "cash_request_out_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[16] = map[string]string{"name": "cash_request_in_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Нужно для отката cash_request_in"}
	s2[17] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[18] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_promised_amount"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_faces_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "race", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "country", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "version", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Версия набора точек"}
	s2[5] = map[string]string{"name": "status", "mysql": "enum('null','approved','rejected','pending') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'null'", "postgresql": "enum('null','approved','rejected','pending') NOT NULL DEFAULT 'null'", "comment": ""}
	s2[6] = map[string]string{"name": "f1", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": "Отрезок 1"}
	s2[7] = map[string]string{"name": "f2", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "f3", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "f4", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "f5", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[11] = map[string]string{"name": "f6", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[12] = map[string]string{"name": "f7", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[13] = map[string]string{"name": "f8", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[14] = map[string]string{"name": "f9", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[15] = map[string]string{"name": "f10", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[16] = map[string]string{"name": "f11", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[17] = map[string]string{"name": "f12", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[18] = map[string]string{"name": "f13", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[19] = map[string]string{"name": "f14", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[20] = map[string]string{"name": "f15", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[21] = map[string]string{"name": "f16", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[22] = map[string]string{"name": "f17", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[23] = map[string]string{"name": "f18", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[24] = map[string]string{"name": "f19", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[25] = map[string]string{"name": "f20", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[26] = map[string]string{"name": "p1", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[27] = map[string]string{"name": "p2", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[28] = map[string]string{"name": "p3", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[29] = map[string]string{"name": "p4", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[30] = map[string]string{"name": "p5", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[31] = map[string]string{"name": "p6", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[32] = map[string]string{"name": "p7", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[33] = map[string]string{"name": "p8", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[34] = map[string]string{"name": "p9", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[35] = map[string]string{"name": "p10", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[36] = map[string]string{"name": "p11", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[37] = map[string]string{"name": "p12", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[38] = map[string]string{"name": "p13", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[39] = map[string]string{"name": "p14", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[40] = map[string]string{"name": "p15", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[41] = map[string]string{"name": "p16", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[42] = map[string]string{"name": "p17", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[43] = map[string]string{"name": "p18", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[44] = map[string]string{"name": "p19", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[45] = map[string]string{"name": "p20", "mysql": "float NOT NULL DEFAULT '0'", "sqlite": "float NOT NULL DEFAULT '0'", "postgresql": "float NOT NULL DEFAULT '0'", "comment": ""}
	s2[46] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[47] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = "Точки по каждому юзеру"
	s["log_faces"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_miners_data_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "miner_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "reg_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "status", "mysql": "enum('null','miner','user','passive_miner','suspended_miner') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'null'", "postgresql": "enum('null','miner','user','passive_miner','suspended_miner') NOT NULL DEFAULT 'null'", "comment": ""}
	s2[5] = map[string]string{"name": "node_public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "face_hash", "mysql": "varchar(128) NOT NULL DEFAULT ''", "sqlite": "varchar(128) NOT NULL DEFAULT ''", "postgresql": "varchar(128) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "profile_hash", "mysql": "varchar(128) NOT NULL DEFAULT ''", "sqlite": "varchar(128) NOT NULL DEFAULT ''", "postgresql": "varchar(128) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "photo_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "photo_max_miner_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "miners_keepers", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[11] = map[string]string{"name": "face_coords", "mysql": "varchar(1024) NOT NULL DEFAULT ''", "sqlite": "varchar(1024) NOT NULL DEFAULT ''", "postgresql": "varchar(1024) NOT NULL DEFAULT ''", "comment": ""}
	s2[12] = map[string]string{"name": "profile_coords", "mysql": "varchar(1024) NOT NULL DEFAULT ''", "sqlite": "varchar(1024) NOT NULL DEFAULT ''", "postgresql": "varchar(1024) NOT NULL DEFAULT ''", "comment": ""}
	s2[13] = map[string]string{"name": "video_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[14] = map[string]string{"name": "video_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[15] = map[string]string{"name": "http_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[16] = map[string]string{"name": "tcp_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[17] = map[string]string{"name": "e_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[18] = map[string]string{"name": "latitude", "mysql": "decimal(8,5) NOT NULL DEFAULT '0'", "sqlite": "decimal(8,5) NOT NULL DEFAULT '0'", "postgresql": "decimal(8,5) NOT NULL DEFAULT '0'", "comment": ""}
	s2[19] = map[string]string{"name": "longitude", "mysql": "decimal(8,5) NOT NULL DEFAULT '0'", "sqlite": "decimal(8,5) NOT NULL DEFAULT '0'", "postgresql": "decimal(8,5) NOT NULL DEFAULT '0'", "comment": ""}
	s2[20] = map[string]string{"name": "country", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[21] = map[string]string{"name": "pool_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[22] = map[string]string{"name": "backup_pool_users", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Когда майнер передумал быть пулом, то нужно залогировать всех, кто его указали как админа пула, чтобы можно было сделать rollback"}
	s2[23] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[24] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_miners_data"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "count", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Сколько новых транзакций сделал юзер за минуту"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = ""
	s["log_minute"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_recycle_bin_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "profile_file_name", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "face_file_name", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[5] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_recycle_bin"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_spots_compatibility_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "version", "mysql": "double NOT NULL DEFAULT '0'", "sqlite": "double NOT NULL DEFAULT '0'", "postgresql": "money NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "example_spots", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "compatibility", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "segments", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "tolerances", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[7] = map[string]string{"name": "prev_log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_spots_compatibility"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_actualization_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_actualization"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_abuses_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "можно создавать только 1 тр-ю с абузами за 24h"
	s["log_time_abuses"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_for_repaid_fix_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_for_repaid_fix"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_commission_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_commission"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_promised_amount_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Для учета кол-ва запр. на доб. / удал. / изменение promised_amount. Чистим кроном"
	s["log_time_promised_amount"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_cash_requests_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_cash_requests"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_geolocation_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_geolocation"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_holidays_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_holidays"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_message_to_admin_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_message_to_admin"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_mining_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_mining"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_change_host_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_change_host"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_new_miner_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_new_miner"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_new_user_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_new_user"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_node_key_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_node_key"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_primary_key_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_primary_key"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_votes_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Храним данные за 1 сутки"
	s["log_time_votes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_votes_miners_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Лимиты для повторых запросов, за которые голосуют ноды"
	s["log_time_votes_miners"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_votes_nodes_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Голоса от нодов"
	s["log_time_votes_nodes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_votes_complex_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_votes_complex"] = s1
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
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_users_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "varchar(30) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "avatar", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "public_key_0", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "public_key_1", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "public_key_2", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "ca1", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "ca2", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "ca3", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[9] = map[string]string{"name": "referral", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "credit_part", "mysql": "decimal(5,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(5,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(5,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[11] = map[string]string{"name": "change_key", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[12] = map[string]string{"name": "change_key_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[13] = map[string]string{"name": "change_key_close", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[14] = map[string]string{"name": "seller_hold_back_pct", "mysql": "decimal(5,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(5,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(5,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[15] = map[string]string{"name": "arbitration_days_refund", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[16] = map[string]string{"name": "url", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[17] = map[string]string{"name": "chat_ban", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[18] = map[string]string{"name": "sn_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[19] = map[string]string{"name": "sn_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[20] = map[string]string{"name": "votes_start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[21] = map[string]string{"name": "votes_0", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[22] = map[string]string{"name": "votes_1", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[23] = map[string]string{"name": "status", "mysql": "enum('user','sn_user', 'rejected_sn_user') NOT NULL DEFAULT 'user'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'user'", "postgresql": "enum('user','sn_user', 'rejected_sn_user') NOT NULL DEFAULT 'user'", "comment": ""}
	s2[24] = map[string]string{"name": "sn_attempts", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[25] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[26] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_users"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_variables_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "data", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_variables"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Кто голосует"}
	s2[1] = map[string]string{"name": "voting_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "За что голосует. тут может быть id geolocation и пр"}
	s2[2] = map[string]string{"name": "type", "mysql": "enum('null','votes_miners','promised_amount','sn_user') NOT NULL", "sqlite": "varchar(100)  NOT NULL", "postgresql": "enum('null','votes_miners','promised_amount','sn_user') NOT NULL", "comment": "Нужно для voting_id' DEFAULT 'null"}
	s2[3] = map[string]string{"name": "del_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было удаление. Нужно для чистки по крону старых данных и для откатов."}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id", "voting_id", "type"}
	s1["comment"] = "Чтобы 1 юзер не смог проголосовать 2 раза за одно и тоже"
	s["log_votes"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_wallets_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "amount", "mysql": "decimal(15,2) UNSIGNED NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "amount_backup", "mysql": "decimal(15,2)  NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "last_update", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[5] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Id предыдщуего log_id, который запишем в wallet"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = "Таблица, где будет браться инфа при откате блока"
	s["log_wallets"] = s1
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
	s2[0] = map[string]string{"name": "miner_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('miners_miner_id_seq')", "comment": "Если есть забаненные, то на их место становятся новички, т.о. все miner_id будут заняты без пробелов"}
	s2[1] = map[string]string{"name": "active", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "1 - активен, 0 - забанен"}
	s2[2] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Без log_id нельзя определить, был ли апдейт в табле miners или же инсерт, т.к. по AUTO_INCREMENT не понять, т.к. обновление может быть в самой последней строке"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"miner_id"}
	s1["AI"] = "miner_id"
	s1["comment"] = ""
	s["miners"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_miners_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[2] = map[string]string{"name": "prev_log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_miners"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "miner_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Из таблицы miners"}
	s2[2] = map[string]string{"name": "reg_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Время, когда майнер получил miner_id по итогам голосования. Определеяется один раз и не меняется. Нужно, чтобы не давать новым майнерам генерить тр-ии регистрации новых юзеров и исходящих запросов"}
	s2[3] = map[string]string{"name": "ban_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке майнер был разжалован в suspended_miner. Нужно для исключения пересечения тр-ий разжалованного майнера и самой тр-ии разжалования"}
	s2[4] = map[string]string{"name": "status", "mysql": "enum('miner','user','passive_miner','suspended_miner') NOT NULL DEFAULT 'user'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'user'", "postgresql": "enum('miner','user','passive_miner','suspended_miner') NOT NULL DEFAULT 'user'", "comment": "Измнеения вызывают персчет TDC в promised_amount"}
	s2[5] = map[string]string{"name": "node_public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "face_hash", "mysql": "varchar(128) NOT NULL DEFAULT ''", "sqlite": "varchar(128) NOT NULL DEFAULT ''", "postgresql": "varchar(128) NOT NULL DEFAULT ''", "comment": "Хэш фото юзера"}
	s2[7] = map[string]string{"name": "profile_hash", "mysql": "varchar(128) NOT NULL DEFAULT ''", "sqlite": "varchar(128) NOT NULL DEFAULT ''", "postgresql": "varchar(128) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "photo_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Блок, в котором было добавлено фото"}
	s2[9] = map[string]string{"name": "photo_max_miner_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Макс. майнер id в момент добавления фото. Это и photo_block_id нужны для определения 10-и нодов, где лежат фото"}
	s2[10] = map[string]string{"name": "miners_keepers", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": "Скольким майнерам копируем фото юзера. По дефолту = 10"}
	s2[11] = map[string]string{"name": "face_coords", "mysql": "varchar(1024) NOT NULL DEFAULT ''", "sqlite": "varchar(1024) NOT NULL DEFAULT ''", "postgresql": "varchar(1024) NOT NULL DEFAULT ''", "comment": ""}
	s2[12] = map[string]string{"name": "profile_coords", "mysql": "varchar(1024) NOT NULL DEFAULT ''", "sqlite": "varchar(1024) NOT NULL DEFAULT ''", "postgresql": "varchar(1024) NOT NULL DEFAULT ''", "comment": ""}
	s2[13] = map[string]string{"name": "video_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[14] = map[string]string{"name": "video_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Если пусто, то видео берем по ID юзера.flv"}
	s2[15] = map[string]string{"name": "http_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "адрес домена:порт или IP:порт, где брать фото и видео данного майнера"}
	s2[16] = map[string]string{"name": "tcp_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "адрес домена:порт или IP:порт, куда ноды должны слать свои TCP пакеты с блоками/хэшами/тр-ми"}
	s2[17] = map[string]string{"name": "e_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "просто для апи запросов на биржу"}
	s2[18] = map[string]string{"name": "latitude", "mysql": "decimal(8,5) NOT NULL DEFAULT '0'", "sqlite": "decimal(8,5) NOT NULL DEFAULT '0'", "postgresql": "decimal(8,5) NOT NULL DEFAULT '0'", "comment": "Местоположение можно сменить без проблем, но это одновременно ведет запуск голосования у promised_amount по всем валютам, где статус mining или hold"}
	s2[19] = map[string]string{"name": "longitude", "mysql": "decimal(8,5) NOT NULL DEFAULT '0'", "sqlite": "decimal(8,5) NOT NULL DEFAULT '0'", "postgresql": "decimal(8,5) NOT NULL DEFAULT '0'", "comment": ""}
	s2[20] = map[string]string{"name": "country", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[21] = map[string]string{"name": "i_am_pool", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[22] = map[string]string{"name": "pool_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[23] = map[string]string{"name": "pool_count_users", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[24] = map[string]string{"name": "backup_pool_users", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Используется только в log_miners_data"}
	s2[25] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = ""
	s["miners_data"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "version", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "alert", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"version"}
	s1["comment"] = "Сюда пишется новая версия, которая загружена в public"
	s["new_version"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "ban_start", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "info", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = "Баним на 1 час тех, кто дает нам данные с ошибками"
	s["nodes_ban"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "host", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Чтобы получать открытый ключ, которым шифруем блоки и тр-ии"}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "ID блока, который есть у данного нода. Чтобы слать ему только >="}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"host"}
	s1["comment"] = "Ноды, которым шлем данные и от которых принимаем данные"
	s["nodes_connection"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('pct_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время блока, в котором были новые %"}
	s2[2] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "miner", "mysql": "decimal(13,13) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,13) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,13) NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "user", "mysql": "decimal(13,13) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,13) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,13) NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "block_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Нужно для откатов"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "% майнера, юзера. На основе  pct_votes"
	s["pct"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('max_promised_amounts_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время блока, в котором были новые max_promised_amount"}
	s2[2] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "amount", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "block_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Нужно для откатов"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "На основе votes_max_promised_amount"
	s["max_promised_amounts"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"time"}
	s1["comment"] = "Время последнего обновления max_other_currencies_time в currency "
	s["max_other_currencies_time"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('reduction_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время блока, в котором было произведено уполовинивание"}
	s2[2] = map[string]string{"name": "notification", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "type", "mysql": "enum('manual','auto') NOT NULL DEFAULT 'auto'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'auto'", "postgresql": "enum('manual','auto') NOT NULL DEFAULT 'auto'", "comment": ""}
	s2[5] = map[string]string{"name": "pct", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "block_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Нужно для откатов"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Когда была последняя процедура урезания для конкретной валюты. Чтобы отсчитывать 2 недели до следующей"
	s["reduction"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Нужно только для того, чтобы определять, голосовал ли юзер или нет. От этого зависит, будет он получать майнерский или юзерский %"}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "pct", "mysql": "decimal(13,13) NOT NULL DEFAULT '0'", "sqlite": "TEXT NOT NULL DEFAULT '0'", "postgresql": "decimal(13,13) NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id", "currency_id"}
	s1["comment"] = "Голосвание за %. Каждые 14 дней пересчет"
	s["votes_miner_pct"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_votes_miner_pct_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "pct", "mysql": "decimal(13,13) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,13) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,13) NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[5] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_votes_miner_pct"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "pct", "mysql": "decimal(13,13) NOT NULL DEFAULT '0'", "sqlite": "TEXT NOT NULL DEFAULT '0'", "postgresql": "decimal(13,13) NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id", "currency_id"}
	s1["comment"] = "Голосвание за %. Каждые 14 дней пересчет"
	s["votes_user_pct"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "e_owner_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": "user_id владельца биржи, за которую идет голосование"}
	s2[2] = map[string]string{"name": "result", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id","e_owner_id"}
	s1["comment"] = "Голосование за биржи"
	s["votes_exchange"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_votes_user_pct_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "pct", "mysql": "decimal(13,13) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,13) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,13) NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[4] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_votes_user_pct"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_votes_exchange_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "result", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[3] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_votes_exchange"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Учитываются только свежие голоса, т.е. один голос только за одно урезание"}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "pct", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id", "currency_id"}
	s1["comment"] = "Голосвание за уполовинивание денежной массы. Каждые 14 дней пересчет"
	s["votes_reduction"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_votes_reduction_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "pct", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[5] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_votes_reduction"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "amount", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Возможные варианты задаются в скрипте, иначе будут проблемы с поиском варианта-победителя"}
	s2[3] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id", "currency_id"}
	s1["comment"] = ""
	s["votes_max_promised_amount"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_votes_max_promised_amount_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "amount", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[4] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_votes_max_promised_amount"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "count", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Возможные варианты задаются в скрипте, иначе будут проблемы с поиском варианта-победителя"}
	s2[3] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id", "currency_id"}
	s1["comment"] = ""
	s["votes_max_other_currencies"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_votes_max_other_currencies_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "count", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[4] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_votes_max_other_currencies"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "time_start", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "От какого времени отсчитывается 1 месяц"}
	s2[2] = map[string]string{"name": "points", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Баллы, полученные майнером за голосования"}
	s2[3] = map[string]string{"name": "log_id", "mysql": "bigint(20)  NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = "Баллы майнеров, по которым решается - получат они майнерские % или юзерские"
	s["points"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_points_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time_start", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "points", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[4] = map[string]string{"name": "prev_log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_points"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "time_start", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Время начала действия статуса. До какого времени действует данный статус определяем простым добавлением в массив времени, которое будет через 30 дней"}
	s2[2] = map[string]string{"name": "status", "mysql": "enum('user','miner') NOT NULL DEFAULT 'user'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'user'", "postgresql": "enum('user','miner') NOT NULL DEFAULT 'user'", "comment": ""}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Нужно для удобного отката"}
	s1["fields"] = s2
	s1["comment"] = "Статусы юзеров на основе подсчета points"
	s["points_status"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "head_hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "user_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"head_hash", "hash"}
	s1["comment"] = "Блоки, которые мы должны забрать у указанных нодов"
	s["queue_blocks"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "head_hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш от заголовка блока (user_id,block_id,prev_head_hash)"}
	s2[1] = map[string]string{"name": "data", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"head_hash"}
	s1["comment"] = "Очередь на фронтальную проверку соревнующихся блоков"
	s["queue_testblock"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "md5 от тр-ии"}
	s2[1] = map[string]string{"name": "high_rate", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Если 1, значит это админская тр-ия"}
	s2[2] = map[string]string{"name": "data", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "_tmp_node_user_id", "mysql": "VARCHAR(255) DEFAULT ''", "sqlite": "VARCHAR(255) DEFAULT ''", "postgresql": "VARCHAR(255) DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = "Тр-ии, которые мы должны проверить"
	s["queue_tx"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "profile_file_name", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "face_file_name", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = ""
	s["recycle_bin"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "version", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "example_spots", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Точки, которые наносим на 2 фото-примера (анфас и профиль)"}
	s2[2] = map[string]string{"name": "compatibility", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "С какими версиями совместимо"}
	s2[3] = map[string]string{"name": "segments", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Нужно для составления отрезков в new_miner()"}
	s2[4] = map[string]string{"name": "tolerances", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Допустимые расхождения между точками при поиске фото-дублей"}
	s2[5] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"version"}
	s1["comment"] = "Совместимость текущей версии точек с предыдущими"
	s["spots_compatibility"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "ID тестируемого блока"}
	s2[1] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время, когда блок попал сюда"}
	s2[2] = map[string]string{"name": "level", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Пишем сюда для использования при формировании заголовка"}
	s2[3] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "По id вычисляем хэш шапки"}
	s2[4] = map[string]string{"name": "header_hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш шапки, им меряемся, у кого меньше - тот круче. Хэш генерим у себя, при получении данных блока"}
	s2[5] = map[string]string{"name": "signature", "mysql": "blob NOT NULL DEFAULT ''", "sqlite": "blob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": "Подпись блока юзером, чей минимальный хэш шапки мы приняли"}
	s2[6] = map[string]string{"name": "mrkl_root", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш тр-ий. Чтобы каждый раз не проверять теже самые данные, просто сравниваем хэши"}
	s2[7] = map[string]string{"name": "status", "mysql": "enum('active','pending') NOT NULL DEFAULT 'active'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'active'", "postgresql": "enum('active','pending') NOT NULL DEFAULT 'active'", "comment": "Указание демону testblock_disseminator"}
	s2[8] = map[string]string{"name": "uniq", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "sent", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"block_id"}
	s1["UNIQ"] = []string{"uniq"}
	s1["comment"] = "Нужно на этапе соревнования, у кого меньше хэш"
	s["testblock"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "lock_time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "script_name", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "uniq", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["UNIQ"] = []string{"uniq"}
	s1["comment"] = ""
	s["testblock_lock"] = s1
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
	s2[7] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера"}
	s2[8] = map[string]string{"name": "third_var", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для исключения пересения в одном блоке удаления обещанной суммы и запроса на её обмен на DC. И для исключения голосования за один и тот же объект одним и тем же юзеров и одном блоке"}
	s2[9] = map[string]string{"name": "counter", "mysql": "tinyint(3) NOT NULL DEFAULT '0'", "sqlite": "tinyint(3) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Чтобы избежать зацикливания при проверке тр-ии: verified=1, новый блок, verified=0. При достижении 10-и - удаляем тр-ию "}
	s2[10] = map[string]string{"name": "sent", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Была отправлена нодам, указанным в nodes_connections"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = "Все незанесенные в блок тр-ии, которые у нас есть"
	s["transactions"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('transactions_testblock_id_seq')", "comment": "Порядок следования очень важен"}
	s2[1] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "md5 для обмена только недостающими тр-ми"}
	s2[2] = map[string]string{"name": "data", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "type", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Тип тр-ии. Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера"}
	s2[4] = map[string]string{"name": "user_id", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера"}
	s2[5] = map[string]string{"name": "third_var", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для исключения пересения в одном блоке удаления обещанной суммы и запроса на её обмен на DC. И для исключения голосования за один и тот же объект одним и тем же юзеров и одном блоке"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["UNIQ"] = []string{"hash"}
	s1["AI"] = "id"
	s1["comment"] = "Тр-ии, которые используются в текущем testblock"
	s["transactions_testblock"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_commission_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "commission", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[3] = map[string]string{"name": "prev_log_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = "Каждый майнер определяет, какая комиссия с тр-ий будет доставаться ему, если он будет генерить блок"
	s["log_commission"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "commission", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Комиссии по всем валютам в json. Если какой-то валюты нет в списке, то комиссия будет равна нулю. currency_id, %, мин., макс."}
	s2[2] = map[string]string{"name": "log_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = "Каждый майнер определяет, какая комиссия с тр-ий будет доставаться ему, если он будет генерить блок"
	s["commission"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('users_user_id_seq')", "comment": "На него будут слаться деньги"}
	s2[1] = map[string]string{"name": "name", "mysql": "varchar(30) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "avatar", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "public_key_0", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Открытый ключ которым проверяются все транзакции от юзера"}
	s2[4] = map[string]string{"name": "public_key_1", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "2-й ключ, если есть"}
	s2[5] = map[string]string{"name": "public_key_2", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "3-й ключ, если есть"}
	s2[6] = map[string]string{"name": "ca1", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "ca2", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "ca3", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[9] = map[string]string{"name": "referral", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Тот, кто зарегал данного юзера и теперь получает с него рефские"}
	s2[10] = map[string]string{"name": "credit_part", "mysql": "decimal(5,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(5,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(5,2) NOT NULL DEFAULT '0'", "comment": "% от поступлений, которые юзер осталяет себе. Если есть активные кредиты, то можно только уменьшать"}
	s2[11] = map[string]string{"name": "change_key", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[12] = map[string]string{"name": "change_key_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[13] = map[string]string{"name": "change_key_close", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[14] = map[string]string{"name": "seller_hold_back_pct", "mysql": "decimal(5,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(5,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(5,2) NOT NULL DEFAULT '0'", "comment": "% холдбека для новых сделок"}
	s2[15] = map[string]string{"name": "arbitration_days_refund", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Продавец тут указывает кол-во дней для новых сделок, в течение которых он готов сделать манибек. Если стоит 0, значит продавец больше не работает с манибеком"}
	s2[16] = map[string]string{"name": "url", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[17] = map[string]string{"name": "chat_ban", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[18] = map[string]string{"name": "sn_type", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[19] = map[string]string{"name": "sn_url_id", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[20] = map[string]string{"name": "votes_start_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[21] = map[string]string{"name": "votes_0", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[22] = map[string]string{"name": "votes_1", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[23] = map[string]string{"name": "status", "mysql": "enum('user','sn_user', 'rejected_sn_user') NOT NULL DEFAULT 'user'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'user'", "postgresql": "enum('user','sn_user', 'rejected_sn_user') NOT NULL DEFAULT 'user'", "comment": ""}
	s2[24] = map[string]string{"name": "sn_attempts", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[25] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["AI"] = "user_id"
	s1["comment"] = ""
	s["users"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "name", "mysql": "varchar(35) NOT NULL DEFAULT ''", "sqlite": "varchar(35) NOT NULL DEFAULT ''", "postgresql": "varchar(35) NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "value", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "comment", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"name"}
	s1["comment"] = ""
	s["variables"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('votes_miners_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "type", "mysql": "enum('null','node_voting','user_voting') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'null'", "postgresql": "enum('null','node_voting','user_voting') NOT NULL DEFAULT 'null'", "comment": ""}
	s2[2] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": "За кого голосуем"}
	s2[3] = map[string]string{"name": "votes_start_time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "votes_0", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "votes_1", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "votes_end", "mysql": "tinyint(1) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(1)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "end_block_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": "В каком блоке мы выставили принудительное end для node"}
	s2[8] = map[string]string{"name": "cron_checked_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "По крону проверили, не нужно ли нам скачать фотки юзера к себе на сервер"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Отдел. от miners_data, чтобы гол. шли точно за свежие данные"
	s["votes_miners"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "first", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "second", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "third", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = "Голосвание за рефские %. Каждые 14 дней пересчет"
	s["votes_referral"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_votes_referral_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "first", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "second", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "third", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "prev_log_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_votes_referral"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "first", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "second", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "third", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "log_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["referral"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_referral_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "first", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "second", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "third", "mysql": "tinyint(2) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(2)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "prev_log_id", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_referral"] = s1
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
	s2[0] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "INTEGER NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "amount", "mysql": "decimal(15,2) unsigned NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "amount_backup", "mysql": "decimal(15,2) unsigned NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": "Может неравномерно обнуляться из-за обработки, а затем - отката new_reduction()"}
	s2[4] = map[string]string{"name": "last_update", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Время последнего пересчета суммы с учетом % из miner_pct"}
	s2[5] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "ID log_wallets, откуда будет брать данные при откате на 1 блок. 0 - значит при откате нужно удалить строку"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id", "currency_id"}
	s1["comment"] = "У кого сколько какой валюты"
	s["wallets"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш транзакции. Нужно для удаления данных из буфера, после того, как транзакция была обработана в блоке, либо анулирована из-за ошибок при повторной проверке"}
	s2[1] = map[string]string{"name": "del_block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Т.к. удалять нельзя из-за возможного отката блока, приходится делать delete=1, а через сутки - чистить"}
	s2[2] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "currency_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "amount", "mysql": "decimal(15,2) unsigned NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Может быть = 0. Номер блока, в котором была занесена запись. Если блок в процессе фронт. проверки окажется невалдиным, то просто удалим все данные по block_id"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = "Суммируем все списания, которые еще не в блоке"
	s["wallets_buffer"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('forex_orders_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Чей ордер"}
	s2[2] = map[string]string{"name": "sell_currency_id", "mysql": "int(10) NOT NULL DEFAULT '0'", "sqlite": "int(10) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Что продается"}
	s2[3] = map[string]string{"name": "sell_rate", "mysql": "decimal(20,10) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,10) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,10) NOT NULL DEFAULT '0'", "comment": "По какому курсу к buy_currency_id"}
	s2[4] = map[string]string{"name": "amount", "mysql": "decimal(15,2)  NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": "Сколько осталось на данном ордере"}
	s2[5] = map[string]string{"name": "amount_backup", "mysql": "decimal(15,2)  NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "buy_currency_id", "mysql": "int(10) NOT NULL DEFAULT '0'", "sqlite": "int(10) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Какая валюта нужна"}
	s2[7] = map[string]string{"name": "commission", "mysql": "decimal(15,2)  NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": "Какую отдали комиссию ноду-генератору"}
	s2[8] = map[string]string{"name": "empty_block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Если ордер опустошили, то тут будет номер блока. Чтобы потом удалить старые записи"}
	s2[9] = map[string]string{"name": "del_block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Если юзер решил удалить ордер, то тут будет номер блока"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["forex_orders"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_forex_orders_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "main_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "ID из log_forex_orders_main. Для откатов"}
	s2[2] = map[string]string{"name": "order_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": "Какой ордер был задействован. Для откатов"}
	s2[3] = map[string]string{"name": "amount", "mysql": "decimal(15,2) unsigned NOT NULL DEFAULT '0'", "sqlite": "decimal(15,2)  NOT NULL DEFAULT '0'", "postgresql": "decimal(15,2)  NOT NULL DEFAULT '0'", "comment": "Какая сумма была вычтена из ордера"}
	s2[4] = map[string]string{"name": "to_user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": "Какому юзеру была начислено amount "}
	s2[5] = map[string]string{"name": "new", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": "Если 1, то был создан новый  ордер. при 1 amount не указывается, т.к. при откате будет просто удалена запись из forex_orders"}
	s2[6] = map[string]string{"name": "block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Для откатов"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Все ордеры, который были затронуты в результате тр-ии"
	s["log_forex_orders"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('log_forex_orders_main_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "Чтобы можно было понять, какие данные можно смело удалять из-за их давности"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Каждый ордер пишется сюда. При откате любого ордера просто берем последнюю строку отсюда"
	s["log_forex_orders_main"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "tx_hash", "mysql": "binary(16) DEFAULT ''", "sqlite": "binary(16) DEFAULT ''", "postgresql": "bytea  DEFAULT ''", "comment": "По этому хэшу отмечается, что данная тр-ия попала в блок и ставится del_block_id"}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "del_block_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": "block_id сюда пишется в тот момент, когда тр-ия попала в блок и уже не используется для фронтальной проверки. Нужно чтобы можно было понять, какие данные можно смело удалять из-за их давности"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"tx_hash"}
	s1["comment"] = "В один блок не должно попасть более чем 10 тр-ий перевода средств или создания forex-ордеров на суммы менее эквивалента 0.05-0.1$ по текущему курсу"
	s["log_time_money_orders"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('x_my_admin_messages_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "add_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "для удаления старых my_pending"}
	s2[2] = map[string]string{"name": "user_int_message_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "ID сообщения, который присылает юзер"}
	s2[3] = map[string]string{"name": "parent_user_int_message_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Parent_id, который присылает юзер"}
	s2[4] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "type", "mysql": "enum('null','from_user','to_user') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'null'", "postgresql": "enum('null','from_user','to_user') NOT NULL DEFAULT 'null'", "comment": ""}
	s2[6] = map[string]string{"name": "subject", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "encrypted", "mysql": "blob NOT NULL DEFAULT ''", "sqlite": "blob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "decrypted", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "message", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[10] = map[string]string{"name": "message_type", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[11] = map[string]string{"name": "message_subtype", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[12] = map[string]string{"name": "status", "mysql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'my_pending'", "postgresql": "enum('my_pending','approved') NOT NULL DEFAULT 'my_pending'", "comment": ""}
	s2[13] = map[string]string{"name": "close", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Воспрос закрыли, чтобы больше не маячил"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Эта табла видна только админу"
	s["x_my_admin_messages"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "data", "mysql": "varchar(20) NOT NULL DEFAULT ''", "sqlite": "varchar(20) NOT NULL DEFAULT ''", "postgresql": "varchar(20) NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = ""
	s["authorization"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"user_id"}
	s1["comment"] = "Если не пусто, то работаем в режиме пула"
	s["community"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "uniq", "mysql": "enum('1') NOT NULL DEFAULT '1'", "sqlite": "varchar(100)  NOT NULL DEFAULT '1'", "postgresql": "enum('1') NOT NULL DEFAULT '1'", "comment": ""}
	s2[1] = map[string]string{"name": "data", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"uniq"}
	s1["comment"] = ""
	s["backup_community"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('payment_systems_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = "Для тех, кто не хочет встречаться для обмена кода на наличные"
	s["payment_systems"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "my_block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Параллельно с info_block пишем и сюда. Нужно при обнулении рабочих таблиц, чтобы знать до какого блока не трогаем таблы my_"}
	s2[1] = map[string]string{"name": "local_gate_ip", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "Если тут не пусто, то connector будет не активным, а ip для disseminator будет браться тут. Нужно для защищенного режима"}
	s2[2] = map[string]string{"name": "static_node_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Все исходящие тр-ии будут подписаны публичным ключом этой ноды. Нужно для защищенного режима"}
	s2[3] = map[string]string{"name": "in_connections_ip_limit", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Кол-во запросов от 1 ip за минуту"}
	s2[4] = map[string]string{"name": "in_connections", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Кол-во нодов и просто юзеров, от кого принимаем данные. Считаем кол-во ip за 1 минуту"}
	s2[5] = map[string]string{"name": "out_connections", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Кол-во нодов, кому шлем данные"}
	s2[6] = map[string]string{"name": "bad_blocks", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Номера и sign плохих блоков. Нужно, чтобы не подцепить более длинную, но глючную цепочку блоков"}
	s2[7] = map[string]string{"name": "pool_max_users", "mysql": "int(11) NOT NULL DEFAULT '100'", "sqlite": "int(11) NOT NULL DEFAULT '100'", "postgresql": "int NOT NULL DEFAULT '100'", "comment": ""}
	s2[8] = map[string]string{"name": "pool_admin_user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "pool_tech_works", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "exchange_api_url", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "На home далается ajax-запрос к api биржи и выдается инфа о курсе и пр."}
	s2[11] = map[string]string{"name": "cf_url", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "URL, который отображается в соц. кнопках и с которого подгружаются css/js/img/fonts при прямом заходе в CF-каталог"}
	s2[12] = map[string]string{"name": "pool_url", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "URL, на который ссылается кнопка Contribute now из внешнего CF-каталога "}
	s2[13] = map[string]string{"name": "pool_email", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "В режиме пула используется как адрес отправителя при рассылке уведомлений"}
	s2[14] = map[string]string{"name": "cf_available_coins_url", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "URL биржи, где можно узнать, сколько там осталось монет в продаже по курсу 1"}
	s2[15] = map[string]string{"name": "cf_exchange_url", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "URL биржи. Просто, чтобы дать на неё ссылку в сообщении, где говорится, что монеты на бирже кончились"}
	s2[16] = map[string]string{"name": "cf_top_html", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "html-код с платежными системами для страницы cf_page_preview"}
	s2[17] = map[string]string{"name": "cf_bottom_html", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "html-код с платежными системами для страницы cf_page_preview"}
	s2[18] = map[string]string{"name": "cf_ps", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Массива с платежными системами, которые будут выводиться на cf_page_preview"}
	s2[19] = map[string]string{"name": "auto_reload", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Если произойдет сбой и в main_lock будет висеть запись более auto_reload секунд, тогда будет запущен сбор блоков с чистого листа"}
	s2[20] = map[string]string{"name": "commission", "mysql": "text NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": "Максимальная комиссия, которую могут поставить ноды на данном пуле"}
	s2[21] = map[string]string{"name": "setup_password", "mysql": "varchar(255)  NOT NULL DEFAULT ''", "sqlite": "varchar(255)  NOT NULL DEFAULT ''", "postgresql": "varchar(255)  NOT NULL DEFAULT ''", "comment": "После установки и после сбора блоков, появляется окно, когда кто-угодно может ввести главный ключ"}
	s2[22] = map[string]string{"name": "sqlite_db_url", "mysql": "varchar(255)  NOT NULL DEFAULT ''", "sqlite": "varchar(255)  NOT NULL DEFAULT ''", "postgresql": "varchar(255)  NOT NULL DEFAULT ''", "comment": "Если не пусто, значит качаем с сервера sqlite базу данных"}
	s2[23] = map[string]string{"name": "first_load_blockchain_url", "mysql": "varchar(255)  NOT NULL DEFAULT ''", "sqlite": "varchar(255)  NOT NULL DEFAULT ''", "postgresql": "varchar(255)  NOT NULL DEFAULT ''", "comment": ""}
	s2[24] = map[string]string{"name": "first_load_blockchain", "mysql": "enum('nodes','file','null') DEFAULT 'null'", "sqlite": "varchar(100)  DEFAULT 'null'", "postgresql": "enum('nodes','file','null') DEFAULT 'null'", "comment": ""}
	s2[25] = map[string]string{"name": "current_load_blockchain", "mysql": "enum('nodes','file','null') DEFAULT 'null'", "sqlite": "varchar(100)  DEFAULT 'null'", "postgresql": "enum('nodes','file','null') DEFAULT 'null'", "comment": "Откуда сейчас собирается база данных"}
	s2[26] = map[string]string{"name": "http_host", "mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": "адрес, по которому будет висеть панель юзера.  Если это майнер, то адрес должен совпадать с my_table.http_host"}
	s2[27] = map[string]string{"name": "auto_update", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[28] = map[string]string{"name": "auto_update_url","mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[29] = map[string]string{"name": "chat_enabled", "mysql": "tinyint(1) NOT NULL DEFAULT '1'", "sqlite": "tinyint(1) NOT NULL DEFAULT '1'", "postgresql": "smallint NOT NULL DEFAULT '1'", "comment": ""}
	s2[30] = map[string]string{"name": "analytics_disabled", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[31] = map[string]string{"name": "stat_host","mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[32] = map[string]string{"name": "getpool_host","mysql": "varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}


	s1["fields"] = s2
	s1["comment"] = ""
	s["config"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "private_key", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "used_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["comment"] = "Ключи для tools/available_keys"
	s["_my_refs"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "token", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255)  NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "e_owner_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"token"}
	s1["comment"] = "Токены для получения инфы с какой-то центральной биржы"
	s["[my_prefix]my_tokens"] = s1
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
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('chat_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "md5 от тр-ии"}
	s2[2] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "lang", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "room", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "receiver", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "sender", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "status", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "message", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[9] = map[string]string{"name": "enc_message", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[10] = map[string]string{"name": "sign_time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для разбавления хэша, иначе одинаковые сообщения содержат одинаковую подпись и хэш получается тоже одинаковый"}
	s2[11] = map[string]string{"name": "signature", "mysql": "blob NOT NULL DEFAULT ''", "sqlite": "blob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[12] = map[string]string{"name": "sent", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["UNIQ"] = []string{"hash"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["chat"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_adding_funds_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "tx_hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "md5 от тр-ии из блока, чтобы исключить повторное зачисление средств"}
	s2[2] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "amount", "mysql": "decimal(10,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(10,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(10,2) NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_adding_funds"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_adding_funds_cp_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "amount", "mysql": "decimal(10,4) NOT NULL DEFAULT '0'", "sqlite": "decimal(10,4) NOT NULL DEFAULT '0'", "postgresql": "decimal(10,4) NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_adding_funds_cp"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_adding_funds_pm_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "amount", "mysql": "decimal(10,4) NOT NULL DEFAULT '0'", "sqlite": "decimal(10,4) NOT NULL DEFAULT '0'", "postgresql": "decimal(10,4) NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_adding_funds_pm"] = s1
	schema.S = s
	schema.PrintSchema()
	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_adding_funds_payeer_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "amount", "mysql": "decimal(10,4) NOT NULL DEFAULT '0'", "sqlite": "decimal(10,4) NOT NULL DEFAULT '0'", "postgresql": "decimal(10,4) NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_adding_funds_payeer"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[1] = map[string]string{"name": "data", "mysql": "varchar(20) NOT NULL DEFAULT ''", "sqlite": "varchar(20) NOT NULL DEFAULT ''", "postgresql": "varchar(20) NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"hash"}
	s1["comment"] = ""
	s["e_authorization"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_charts_data_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "wallets", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "promised_amounts", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_charts_data"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "rate", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["e_charts_data_dwoc"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_config_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "value", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_config"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "sort_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_currency_id_seq')", "comment": ""}
	s2[2] = map[string]string{"name": "name", "mysql": "varchar(4) NOT NULL DEFAULT ''", "sqlite": "varchar(4) NOT NULL DEFAULT ''", "postgresql": "varchar(4) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "min_withdraw", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_currency"] = s1
	schema.S = s
	schema.PrintSchema()



	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_currency_pair_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "currency", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "dc_currency", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_currency_pair"] = s1
	schema.S = s
	schema.PrintSchema()



	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_pages_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "name", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "lang", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "title", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "text", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_pages"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_orders_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "sell_currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "sell_rate", "mysql": "decimal(20,10) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,10) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,10) NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "begin_amount", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "amount", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "buy_currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "empty_time", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "del_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_orders"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "block_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "pct", "mysql": "tinyint(2) NOT NULL DEFAULT '0'", "sqlite": "tinyint(2) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["e_reduction"] = s1
	schema.S = s
	schema.PrintSchema()



	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "uniq", "mysql": "tinyint(1) NOT NULL DEFAULT '1'", "sqlite": "tinyint(1) NOT NULL DEFAULT '1'", "postgresql": "smallint NOT NULL DEFAULT '1'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["e_reduction_lock"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_tokens_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "token", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "status", "mysql": "enum('null','wait','paid') NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'null'", "postgresql": "enum('null','wait','paid') NOT NULL DEFAULT 'null'", "comment": ""}
	s2[3] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "ps", "mysql": "varchar(11) NOT NULL DEFAULT ''", "sqlite": "varchar(11) NOT NULL DEFAULT ''", "postgresql": "varchar(11) NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "buy_currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "amount_fiat", "mysql": "decimal(10,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(10,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(10,2) NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_tokens"] = s1
	schema.S = s
	schema.PrintSchema()



	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_trade_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "sell_currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "sell_rate", "mysql": "decimal(20,10) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,10) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,10) NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "amount", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "buy_currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "main", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_trade"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "block_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "pct", "mysql": "decimal(13,13) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,13) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,13) NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["e_user_pct"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_users_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "project_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "email", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "password", "mysql": "varchar(64) NOT NULL DEFAULT ''", "sqlite": "varchar(64) NOT NULL DEFAULT ''", "postgresql": "varchar(64) NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "ip", "mysql": "varchar(30) NOT NULL DEFAULT ''", "sqlite": "varchar(30) NOT NULL DEFAULT ''", "postgresql": "varchar(30) NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "salt", "mysql": "varchar(32) NOT NULL DEFAULT ''", "sqlite": "varchar(32) NOT NULL DEFAULT ''", "postgresql": "varchar(32) NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "phone", "mysql": "varchar(20) NOT NULL DEFAULT ''", "sqlite": "varchar(20) NOT NULL DEFAULT ''", "postgresql": "varchar(20) NOT NULL DEFAULT ''", "comment": ""}
	s2[7] = map[string]string{"name": "sms_count", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "sms_count_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "lock", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[10] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_users"] = s1
	schema.S = s
	schema.PrintSchema()




	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "amount", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "last_update", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["comment"] = ""
	s["e_wallets"] = s1
	schema.S = s
	schema.PrintSchema()



	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('e_withdraw_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "open_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "close_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "user_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "account", "mysql": "varchar(100) NOT NULL DEFAULT ''", "sqlite": "varchar(100) NOT NULL DEFAULT ''", "postgresql": "varchar(100) NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "amount", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "wd_amount", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "method", "mysql": "varchar(100) NOT NULL DEFAULT 'null'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'null'", "postgresql": "varchar(100) NOT NULL DEFAULT 'null'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_withdraw"] = s1
	schema.S = s
	schema.PrintSchema()





	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "block_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[1] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "cf_funding", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "forex_orders", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "promised_amount_cash_request_out_time", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "promised_amount_tdc_amount", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "wallets", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"block_id", "currency_id"}
	s1["comment"] = ""
	s["reduction_backup"] = s1
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
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_arbitrator_conditions_log_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "last_payment_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "prev_log_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_auto_payments"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('auto_payments_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "amount", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "commission", "mysql": "decimal(20,8) NOT NULL DEFAULT '0'", "sqlite": "decimal(20,8) NOT NULL DEFAULT '0'", "postgresql": "decimal(20,8) NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "currency_id", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "last_payment_time", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Когда был последний платеж. При создании авто-платежа пишется текущее время"}
	s2[5] = map[string]string{"name": "period", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "sender", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "recipient", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s2[8] = map[string]string{"name": "comment", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[9] = map[string]string{"name": "block_id", "mysql": "int(11)  unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)   NOT NULL DEFAULT '0'", "postgresql": "int   NOT NULL DEFAULT '0'", "comment": "Для отката новой записи об авто-платеже"}
	s2[10] = map[string]string{"name": "del_block_id", "mysql": "int(11)  unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)   NOT NULL DEFAULT '0'", "postgresql": "int   NOT NULL DEFAULT '0'", "comment": "Чистим по крону старые данные раз в сутки. Удалять нельзя, т.к. нужно откатывать"}
	s2[11] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["auto_payments"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_auto_payments_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_auto_payments"] = s1
	schema.S = s
	schema.PrintSchema()

	

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('log_time_del_user_from_pool_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["log_time_del_user_from_pool"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('stats_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "day", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "month", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "year", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "currency_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "dc", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "promised_amount", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["UNIQ"] = []string{"day", "month", "year", "currency_id"}
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["stats"] = s1
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
	s2[0] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('log_promised_amount_restricted_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "dc_amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": "Списанная сумма намайненного"}
	s2[2] = map[string]string{"name": "last_update", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время последнего перевода намайненного на счет"}
	s2[3] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[4] = map[string]string{"name": "prev_log_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"log_id"}
	s1["AI"] = "log_id"
	s1["comment"] = ""
	s["log_promised_amount_restricted"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('promised_amount_restricted_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "currency_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "dc_amount", "mysql": "decimal(13,2) NOT NULL DEFAULT '0'", "sqlite": "decimal(13,2) NOT NULL DEFAULT '0'", "postgresql": "decimal(13,2) NOT NULL DEFAULT '0'", "comment": "Списанная сумма намайненного"}
	s2[5] = map[string]string{"name": "last_update", "mysql": "int(11) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(11)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время последнего перевода намайненного на счет"}
	s2[6] = map[string]string{"name": "log_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["promised_amount_restricted"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('notifications_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[3] = map[string]string{"name": "cmd_id", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[4] = map[string]string{"name": "params", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[5] = map[string]string{"name": "isread", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}

	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["notifications"] = s1
	schema.S = s
	schema.PrintSchema()
	schema.DB.Exec(`CREATE INDEX notifications_ur ON notifications (user_id,isread)`)

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('notifications_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "user_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[2] = map[string]string{"name": "subject", "mysql": "varchar(255) NOT NULL DEFAULT ''", "sqlite": "varchar(255) NOT NULL DEFAULT ''", "postgresql": "varchar(255) NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "topic", "mysql": "text CHARACTER SET utf8 NOT NULL DEFAULT ''", "sqlite": "text NOT NULL DEFAULT ''", "postgresql": "text NOT NULL DEFAULT ''", "comment": ""}
	s2[4] = map[string]string{"name": "idroot", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[5] = map[string]string{"name": "time", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
	s2[6] = map[string]string{"name": "status", "mysql": "tinyint(3) unsigned NOT NULL DEFAULT '0'", "sqlite": "tinyint(3)  NOT NULL DEFAULT '0'", "postgresql": "smallint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "uptime", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}

	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["AI"] = "id"
	s1["comment"] = ""
	s["e_tickets"] = s1
	schema.S = s
	schema.PrintSchema()
	schema.DB.Exec(`CREATE INDEX e_ticket_idroot ON e_tickets (idroot)`)
	schema.DB.Exec(`CREATE INDEX e_ticket_uptime ON e_tickets (uptime)`)

	prefix := ""
	if schema.PrefixUserId > 0 {
		prefix = utils.IntToStr(schema.PrefixUserId) + `_`
	}
	schema.DB.Exec(`INSERT INTO ` + prefix + `my_notifications (name, email, sms, mobile) VALUES ('admin_messages',1,1,1),('change_in_status',1,0,0),('dc_came_from',1,0,1),('dc_sent',1,0,0),('incoming_cash_requests',1,1,1),('node_time',0,0,0),('system_error',1,1,0),('update_email',1,0,0),('update_primary_key',1,0,0),('update_sms_request',1,0,0),('voting_results',0,0,0),('voting_time',1,0,0)`)
	schema.DB.Exec(`INSERT INTO e_currency VALUES (1,1001,'USD',5),(2,72,'dUSD',5),(3,23,'dEUR',10),(4,1,'dWOC',100),(5,1002,'BTC',0.01)`)
	schema.DB.Exec(`INSERT INTO e_currency_pair VALUES (1,1001,72),(2,1001,1),(3,1001,23),(4,1002,72)`)
}

func (schema *SchemaStruct) typeMysql() {
	var err error
	var result string
	for table_name, v := range schema.S {
		if ok, _ := regexp.MatchString(`\[my_prefix\]`, table_name); !ok {
			if schema.PrefixUserId > 0 {
				continue
			}
		}
		AI := ""
		AI_START := "1"
		schema.replMy(&table_name)

		result = ""
		/*if schema.ChangeType {
			if !schema.OnlyPrint {
				err = schema.DCDB.ExecSql(fmt.Sprintf("ALTER TABLE \"%[1]s\" RENAME TO tmp;\n", table_name))
			} else {
				fmt.Println(fmt.Sprintf("ALTER TABLE \"%[1]s\" RENAME TO tmp;\n", table_name))
			}
		}

		if !schema.AddColumn {
			if !schema.OnlyPrint {
				err = schema.DCDB.ExecSql("DROP TABLE IF EXISTS " + table_name)
			} else {
				fmt.Println("DROP TABLE IF EXISTS " + table_name+";")
			}
		}*/
		if schema.ChangeType {
			if !schema.OnlyPrint {
				err = schema.DCDB.ExecSql(fmt.Sprintf("ALTER TABLE %[1]s RENAME TO tmp;\n", table_name))
			} else {
				fmt.Println(fmt.Sprintf("ALTER TABLE %[1]s RENAME TO tmp;\n", table_name))
			}
			//result += fmt.Sprintf("ALTER TABLE %[1]s RENAME TO tmp;\n", table_name)
		}
		if !schema.AddColumn {
			if !schema.OnlyPrint {
				err = schema.DCDB.ExecSql(fmt.Sprintf("DROP TABLE IF EXISTS %[1]s;\n", table_name))
			} else {
				fmt.Println(fmt.Sprintf("DROP TABLE IF EXISTS %[1]s;\n", table_name))
			}
			//result += fmt.Sprintf("DROP TABLE IF EXISTS %[1]s;\n", table_name)
		}
		if err != nil {
			log.Error("%v %v", err, table_name)
		}

		if !schema.AddColumn {
			result += fmt.Sprintf("CREATE TABLE IF NOT EXISTS %[1]s (\n", table_name)
		} else {
			result += fmt.Sprintf("ALTER TABLE %[1]s\n", table_name)
		}

		var tableComment string
		primaryKey := ""
		uniqKey := ""
		var tableSlice []string
		for k, v1 := range v.(Recmap) {
			if k == "comment" {
				tableComment = v1.(string)
				//fmt.Println(k, v1.(string), v1)
			} else if k == "fields" {
				//fmt.Println(k, v1)
				//i:=0
				//end:=""
				for i := 0; i < len(v1.(Recmapi)); i++ {
					/*if i == len(v1.(Recmap)) - 1 {
						end = ""
					} else {
						end = ","
					}*/
					dType := v1.(Recmapi)[i].(map[string]string)["mysql"]
					if ok, _ := regexp.MatchString(`AUTO_INCREMENT`, dType); ok {
						dType = strings.Replace(dType, "DEFAULT '0'", "", -1)
					}
					tableSlice = append(tableSlice, fmt.Sprintf("`%s` %s COMMENT '%s'", v1.(Recmapi)[i].(map[string]string)["name"], dType, v1.(Recmapi)[i].(map[string]string)["comment"]))
					//fmt.Println(i)
					//i++
				}
			} else if k == "PRIMARY" {
				primaryKey = fmt.Sprintf("PRIMARY KEY (`%s`)", strings.Join(v1.([]string), "`,`"))
			} else if k == "UNIQ" {
				uniqKey = fmt.Sprintf("UNIQUE KEY (`%v`)", strings.Join(v1.([]string), "`,`"))
			} else if k == "AI" {
				AI = v1.(string)
			} else if k == "AI_START" {
				AI_START = v1.(string)
			}
		}
		if len(uniqKey) > 0 {
			tableSlice = append(tableSlice, uniqKey)
			//fmt.Printf("%s,\n", uniqKey)
		}
		if len(primaryKey) > 0 {
			tableSlice = append(tableSlice, primaryKey)
			//fmt.Printf("%s\n", primaryKey)
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
			if len(AI) > 0 {
				result += fmt.Sprintf(") ENGINE=MyISAM  DEFAULT CHARSET=latin1 AUTO_INCREMENT=%s COMMENT='%s';\n\n", AI_START, tableComment)
			} else {
				result += fmt.Sprintf(") ENGINE=MyISAM  DEFAULT CHARSET=latin1 COMMENT='%s';\n\n", tableComment)
			}
		} else {
			result += ";"
		}
		if schema.ChangeType {
			result += fmt.Sprintf("INSERT INTO %[1]s SELECT * FROM tmp;\nDROP TABLE tmp;\n", table_name)
		}
		if !schema.OnlyPrint {
			err = schema.DCDB.ExecSql(result)
			log.Debug("sql", result)
		} else {
			fmt.Println(result)
		}
		if err != nil {
			log.Error("%s", err)
		}
	}
}

func (schema *SchemaStruct) typePostgresql() {
	var result string
	var err error
	for table_name, v := range schema.S {
		if ok, _ := regexp.MatchString(`\[my_prefix\]`, table_name); !ok {
			if schema.PrefixUserId > 0 {
				continue
			}
		}
		result = ""
		schema.replMy(&table_name)
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
								//fmt.Println(match)
								//fmt.Println(enumSlice)
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

func (schema *SchemaStruct) replMy(table_name *string) {
	if ok, _ := regexp.MatchString(`\[my_prefix\]`, *table_name); ok {
		if schema.PrefixUserId == 0 {
			*table_name = strings.Replace(*table_name, "[my_prefix]", "", -1)
		} else {
			*table_name = strings.Replace(*table_name, "[my_prefix]", utils.IntToStr(schema.PrefixUserId)+"_", -1)
		}
	}
}

func (schema *SchemaStruct) typeSqlite() {
	var result string
	for table_name, v := range schema.S {
		log.Debug("table_name", table_name)
		if ok, _ := regexp.MatchString(`\[my_prefix\]`, table_name); !ok {
			if schema.PrefixUserId > 0 {
				continue
			}
		}
		result = ""
		schema.replMy(&table_name)

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
		AI_START := "1"
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
			} else if k == "AI_START" {
				AI_START = v1.(string)
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
		//log.Println("AI_START=", AI_START)
		if AI_START != "1" {
			q := `BEGIN TRANSACTION; UPDATE sqlite_sequence SET seq = 999 WHERE name = 'cf_currency';INSERT INTO sqlite_sequence (name,seq) SELECT 'cf_currency', 999 WHERE NOT EXISTS (SELECT changes() AS change FROM sqlite_sequence WHERE change <> 0);COMMIT;`
			if !schema.OnlyPrint {
				err := schema.DCDB.ExecSql(q)
				//log.Println(q)
				if err != nil {
					log.Error("%v", err)
				}
			} else {
				fmt.Println(result)
			}
		}
	}
}

func (schema *SchemaStruct) PrintSchema() {
	switch schema.DbType {
	case "mysql":
		schema.typeMysql()
	case "sqlite":
		schema.typeSqlite()
	case "postgresql":
		schema.typePostgresql()
	}
}
