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
	s2[9] = map[string]string{"name": "wallet_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "cb_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": ""}
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
	s["queue_candidateBlock"] = s1
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
	s2[0] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "ID тестируемого блока"}
	s2[1] = map[string]string{"name": "time", "mysql": "int(10) unsigned NOT NULL DEFAULT '0'", "sqlite": "int(10)  NOT NULL DEFAULT '0'", "postgresql": "int  NOT NULL DEFAULT '0'", "comment": "Время, когда блок попал сюда"}
	s2[2] = map[string]string{"name": "level", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Пишем сюда для использования при формировании заголовка"}
	s2[3] = map[string]string{"name": "user_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "По id вычисляем хэш шапки"}
	s2[4] = map[string]string{"name": "header_hash", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш шапки, им меряемся, у кого меньше - тот круче. Хэш генерим у себя, при получении данных блока"}
	s2[5] = map[string]string{"name": "signature", "mysql": "blob NOT NULL DEFAULT ''", "sqlite": "blob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": "Подпись блока юзером, чей минимальный хэш шапки мы приняли"}
	s2[6] = map[string]string{"name": "mrkl_root", "mysql": "binary(32) NOT NULL DEFAULT ''", "sqlite": "binary(32) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Хэш тр-ий. Чтобы каждый раз не проверять теже самые данные, просто сравниваем хэши"}
	s2[7] = map[string]string{"name": "status", "mysql": "enum('active','pending') NOT NULL DEFAULT 'active'", "sqlite": "varchar(100)  NOT NULL DEFAULT 'active'", "postgresql": "enum('active','pending') NOT NULL DEFAULT 'active'", "comment": "Указание демону candidateBlock_disseminator"}
	s2[8] = map[string]string{"name": "uniq", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "sent", "mysql": "tinyint(1) NOT NULL DEFAULT '0'", "sqlite": "tinyint(1) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"block_id"}
	s1["UNIQ"] = []string{"uniq"}
	s1["comment"] = "Нужно на этапе соревнования, у кого меньше хэш"
	s["candidateBlock"] = s1
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
	s["candidateBlock_lock"] = s1
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
	s2[0] = map[string]string{"name": "id", "mysql": "int(11) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "int NOT NULL  default nextval('transactions_candidateBlock_id_seq')", "comment": "Порядок следования очень важен"}
	s2[1] = map[string]string{"name": "hash", "mysql": "binary(16) NOT NULL DEFAULT ''", "sqlite": "binary(16) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "md5 для обмена только недостающими тр-ми"}
	s2[2] = map[string]string{"name": "data", "mysql": "longblob NOT NULL DEFAULT ''", "sqlite": "longblob NOT NULL DEFAULT ''", "postgresql": "bytea NOT NULL DEFAULT ''", "comment": ""}
	s2[3] = map[string]string{"name": "type", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Тип тр-ии. Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера"}
	s2[4] = map[string]string{"name": "user_id", "mysql": "tinyint(4) NOT NULL DEFAULT '0'", "sqlite": "tinyint(4) NOT NULL DEFAULT '0'", "postgresql": "smallint NOT NULL DEFAULT '0'", "comment": "Нужно для недопущения попадения в блок 2-х тр-ий одного типа от одного юзера"}
	s2[5] = map[string]string{"name": "third_var", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "Для исключения пересения в одном блоке удаления обещанной суммы и запроса на её обмен на DC. И для исключения голосования за один и тот же объект одним и тем же юзеров и одном блоке"}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"id"}
	s1["UNIQ"] = []string{"hash"}
	s1["AI"] = "id"
	s1["comment"] = "Тр-ии, которые используются в текущем candidateBlock"
	s["transactions_candidateBlock"] = s1
	schema.S = s
	schema.PrintSchema()

	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "wallet_id", "mysql": "bigint(20) unsigned NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint  NOT NULL  default nextval('wallets_wallet_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "hash", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "public_key_0", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Открытый ключ которым проверяются все транзакции от юзера"}
	s2[3] = map[string]string{"name": "public_key_1", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "2-й ключ, если есть"}
	s2[4] = map[string]string{"name": "public_key_2", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "3-й ключ, если есть"}
	s2[5] = map[string]string{"name": "node_public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "amount", "mysql": "decimal(30) NOT NULL DEFAULT '0'", "sqlite": "decimal(30) NOT NULL DEFAULT '0'", "postgresql": "decimal(30) NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "host", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "vote", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "log_id", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"wallet_id"}
	s1["AI"] = "wallet_id"
	s1["comment"] = ""
	s["wallets"] = s1
	schema.S = s
	schema.PrintSchema()


	s = make(Recmap)
	s1 = make(Recmap)
	s2 = make(Recmapi)
	s2[0] = map[string]string{"name": "rb_id", "mysql": "bigint(20) NOT NULL AUTO_INCREMENT DEFAULT '0'", "sqlite": "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL", "postgresql": "bigint NOT NULL  default nextval('rb_wallets_rb_id_seq')", "comment": ""}
	s2[1] = map[string]string{"name": "hash", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[2] = map[string]string{"name": "public_key_0", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "Открытый ключ которым проверяются все транзакции от юзера"}
	s2[3] = map[string]string{"name": "public_key_1", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "2-й ключ, если есть"}
	s2[4] = map[string]string{"name": "public_key_2", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": "3-й ключ, если есть"}
	s2[5] = map[string]string{"name": "node_public_key", "mysql": "varbinary(512) NOT NULL DEFAULT ''", "sqlite": "varbinary(512) NOT NULL DEFAULT ''", "postgresql": "bytea  NOT NULL DEFAULT ''", "comment": ""}
	s2[6] = map[string]string{"name": "amount", "mysql": "decimal(30) NOT NULL DEFAULT '0'", "sqlite": "decimal(30) NOT NULL DEFAULT '0'", "postgresql": "decimal(30) NOT NULL DEFAULT '0'", "comment": ""}
	s2[7] = map[string]string{"name": "host", "mysql": "varchar(50) NOT NULL DEFAULT ''", "sqlite": "varchar(50) NOT NULL DEFAULT ''", "postgresql": "varchar(50) NOT NULL DEFAULT ''", "comment": ""}
	s2[8] = map[string]string{"name": "vote", "mysql": "bigint(20) unsigned NOT NULL DEFAULT '0'", "sqlite": "bigint(20)  NOT NULL DEFAULT '0'", "postgresql": "bigint  NOT NULL DEFAULT '0'", "comment": ""}
	s2[9] = map[string]string{"name": "block_id", "mysql": "int(11) NOT NULL DEFAULT '0'", "sqlite": "int(11) NOT NULL DEFAULT '0'", "postgresql": "int NOT NULL DEFAULT '0'", "comment": "В каком блоке было занесено. Нужно для удаления старых данных"}
	s2[10] = map[string]string{"name": "prev_rb_id", "mysql": "bigint(20) NOT NULL DEFAULT '0'", "sqlite": "bigint(20) NOT NULL DEFAULT '0'", "postgresql": "bigint NOT NULL DEFAULT '0'", "comment": ""}
	s1["fields"] = s2
	s1["PRIMARY"] = []string{"rb_id"}
	s1["AI"] = "rb_id"
	s1["comment"] = ""
	s["rb_wallets"] = s1
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
	case "mysql":
		schema.typeMysql()
	case "sqlite":
		schema.typeSqlite()
	case "postgresql":
		schema.typePostgresql()
	}
}
