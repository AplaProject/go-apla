package consts


// текущая версия
const VERSION = "2.4.1b3"

// чтобы не выдавать одно и тоже голосование
const ASSIGN_TIME = 86400

const DAY = 3600 * 24
const DAY2 = 3600 * 24 * 2

// используется в confirmations
const COUNT_CONFIRMED_NODES = 5
const WAIT_CONFIRMED_NODES = 10

// на сколько % автоматически урезаем денежную массу
const AUTO_REDUCTION_PCT = 10

// У скольких нодов должен быть такой же блок как и у нас, чтобы считать, что блок у большей части DC-сети. для get_confirmed_block_id()
const MIN_CONFIRMED_NODES = 1


// примерный текущий крайний блок
const LAST_BLOCK = 330000

// примерный размер блокчейна
const BLOCKCHAIN_SIZE = 77000000

// где лежит блокчейн. для тех, кто не хочет собирать его с нодов
const BLOCKCHAIN_URL = "http://dcoin.club/blockchain"

// на сколько может бежать время в тр-ии
const MAX_TX_FORW = 0

// тр-ия может блуждать по сети сутки и потом попасть в блок
const MAX_TX_BACK = DAY

const MAX_BLOCK_SIZE = 16<<20

const USD_CURRENCY_ID = 71

const ARBITRATION_BLOCK_START = 189300

// через какое время админ имеет право изменить ключ юзера, если тот дал на это свое согласие. Это время дается юзеру на то, чтобы отменить запрос.
const CHANGE_KEY_PERIOD_170770 = 3600
const CHANGE_KEY_PERIOD = 3600 * 24 * 30

//  есть ли хотябы X юзеров, у которых на кошелках есть от 0.01 данной валюты
const AUTO_REDUCTION_PROMISED_AMOUNT_MIN = 10

// сколько должно быть процентов PROMISED_AMOUNT от кол-ва DC на кошельках, чтобы запустилось урезание
const AUTO_REDUCTION_PROMISED_AMOUNT_PCT = 1 // X*100%

const LIMIT_NEW_CF_PROJECT = 5
const LIMIT_NEW_CF_PROJECT_PERIOD = 3600 * 24 * 7
const LIMIT_CF_PROJECT_DATA = 10
const LIMIT_CF_PROJECT_DATA_PERIOD = 3600 * 24
const LIMIT_CF_SEND_DC = 10
const LIMIT_CF_SEND_DC_PERIOD = 3600 * 24
const LIMIT_CF_COMMENTS = 10
const LIMIT_CF_COMMENTS_PERIOD = 3600 * 24

// сколько можно делать комментов за сутки за 1 проект
const LIMIT_TIME_COMMENTS_CF_PROJECT = 3600 * 24

const LIMIT_USER_AVATAR = 5
const LIMIT_USER_AVATAR_PERIOD = 3600 * 24

const LIMIT_NEW_CREDIT = 10
const NEW_CREDIT_PERIOD = 3600 * 24
const LIMIT_CHANGE_CREDITOR = 10
const CHANGE_CREDITOR_PERIOD = 3600 * 24
const LIMIT_REPAYMENT_CREDIT = 5
const REPAYMENT_CREDIT_PERIOD = 3600 * 24
const LIMIT_CHANGE_CREDIT_PART = 10
const LIMIT_CHANGE_CREDIT_PART_PERIOD = 3600 * 24
const LIMIT_CHANGE_KEY_ACTIVE = 3
const LIMIT_CHANGE_KEY_ACTIVE_PERIOD = 3600 * 24 * 7
const LIMIT_CHANGE_KEY_REQUEST = 1
const LIMIT_CHANGE_KEY_REQUEST_PERIOD = 3600 * 24 * 7
const LIMIT_CHANGE_ARBITRATION_TRUST_LIST = 3
const LIMIT_CHANGE_ARBITRATION_TRUST_LIST_PERIOD = 3600 * 24
const LIMIT_CHANGE_ARBITRATOR_CONDITIONS = 3
const LIMIT_CHANGE_ARBITRATOR_CONDITIONS_PERIOD = 3600 * 24
const LIMIT_MONEY_BACK_REQUEST = 3
const LIMIT_MONEY_BACK_REQUEST_PERIOD = 3600 * 24
const LIMIT_CHANGE_SELLER_HOLD_BACK = 3
const LIMIT_CHANGE_SELLER_HOLD_BACK_PERIOD = 3600 * 24
const LIMIT_CHANGE_CA = 3
const LIMIT_CHANGE_CA_PERIOD = 3600 * 24
const LIMIT_AUTO_PAYMENTS = 5
const LIMIT_AUTO_PAYMENTS_PERIOD = 3600 * 24
const LIMIT_DEL_USER_FROM_POOL = 10
const LIMIT_DEL_USER_FROM_POOL_PERIOD = 3600 * 24
const LIMIT_SN_USER = 3
const LIMIT_SN_USER_PERIOD = 3600 * 24 * 2

const LIMIT_SN_VOTES_0 = 5
const LIMIT_SN_VOTES_1 = 5
const LIMIT_SN_VOTES_PERIOD = 86400

const SN_USER_ATTEMPTS = 3

const CRON_CHECKED_TIME_SEC = 86400 * 3

const ROUND_FIX = 0.00000000001

// таймауты для TCP
const READ_TIMEOUT = 20
const WRITE_TIMEOUT = 20

// дефолтное знаение, со скольким нодами устанавляиваем связь
const OUT_CONNECTIONS = 10

// на какое время баним нода, давшего нам плохие данные
const NODE_BAN_TIME = 3600

// через сколько можно делать следующее урезание.
// важно учитывать то, что не должно быть роллбеков дальше чем на 1 урезание
// т.к. при урезании используется backup в этой же табле вместо отдельной таблы log_
const AUTO_REDUCTION_PERIOD = 3600 * 24 * 2

const LIMIT_ACTUALIZATION = 1
const LIMIT_ACTUALIZATION_PERIOD = 3600 * 24 * 14

// на сколько арбитр может продлить время рассмотрения манибека
const MAX_MONEY_BACK_TIME = 180

const CHAT_PORT = "8150"

const COUNT_CHAT_NODES = 10

const CHAT_COUNT_MESSAGES = 20
const CHAT_MAX_MESSAGES = 1000

const ALERT_KEY = `30820122300d06092a864886f70d01010105000382010f003082010a0282010100d4a48242d0fb2c7c295bc9c87b1aa0c6d23b5f8cab2ec20c2dfde35513ef6066b92ee3935f9a38100493717b60bb7832411daee02012f44a9f58ac91056b2603661544116bfbc55181e5a693bace5ec9325ba0232b9c9c0a29096569d217243e5bf891cc7fc4bcd2e7d6518acc6f982aaa43a9ed737e3ea2845d6432a823ee5b40d1548f802d0c108bf6e5cb5a4daa7edb48764dcbfa6b7a961208833996cfee265ca2ce2655d444cf3c177b3841b1cc4f3102f89cb2bdb1e5a68eac270506147dd8391b7b3af40a50be13c3970077faffaf98ccc5b8c011146be9c2eb9dfd3454f67a68daaf385d334366d132308bffede27656a515ff69a260bbe2452bd2c30203010001`
const UPD_AND_VER_URL = "http://dcoin.club"
const GOOGLE_API_KEY = "AIzaSyBLZlUPgd9uhX05OrsFU68yJOZFrYhZe84"

var LangMap = map[string]int{"en": 1, "ru": 42}

var MyTables = []string{"my_admin_messages", "my_cash_requests", "my_comments", "my_commission", "my_complex_votes", "my_dc_transactions", "my_holidays", "my_keys", "my_new_users", "my_node_keys", "my_notifications", "my_promised_amount", "my_table", "my_tasks", "my_cf_funding", "my_tokens"}

var ReductionDC = []int64{0, 10, 25, 50, 90}
var Countries = []string{"Afghanistan", "Albania", "Algeria", "American Samoa", "Andorra", "Angola", "Anguilla", "Antarctica", "Antigua and Barbuda", "Argentina", "Armenia", "Aruba", "Australia", "Austria", "Azerbaijan", "Bahamas", "Bahrain", "Bangladesh", "Barbados", "Belarus", "Belgium", "Belize", "Benin", "Bermuda", "Bhutan", "Bolivia", "Bosnia and Herzegovina", "Botswana", "Bouvet Island", "Brazil", "British Indian Ocean Territory", "British Virgin Islands", "Brunei", "Bulgaria", "Burkina Faso", "Burundi", "Cambodia", "Cameroon", "Canada", "Cape Verde", "Cayman Islands", "Central African Republic", "Chad", "Chile", "China", "Christmas Island", "Cocos [Keeling] Islands", "Colombia", "Comoros", "Congo [DRC]", "Congo [Republic]", "Cook Islands", "Costa Rica", "Croatia", "Cuba", "Cyprus", "Czech Republic", "Côte d\"Ivoire", "Denmark", "Djibouti", "Dominica", "Dominican Republic", "Ecuador", "Egypt", "El Salvador", "Equatorial Guinea", "Eritrea", "Estonia", "Ethiopia", "Falkland Islands [Islas Malvinas]", "Faroe Islands", "Fiji", "Finland", "France", "French Guiana", "French Polynesia", "French Southern Territories", "Gabon", "Gambia", "Gaza Strip", "Georgia", "Germany", "Ghana", "Gibraltar", "Greece", "Greenland", "Grenada", "Guadeloupe", "Guam", "Guatemala", "Guernsey", "Guinea", "Guinea-Bissau", "Guyana", "Haiti", "Heard Island and McDonald Islands", "Honduras", "Hong Kong", "Hungary", "Iceland", "India", "Indonesia", "Iran", "Iraq", "Ireland", "Isle of Man", "Israel", "Italy", "Jamaica", "Japan", "Jersey", "Jordan", "Kazakhstan", "Kenya", "Kiribati", "Kosovo", "Kuwait", "Kyrgyzstan", "Laos", "Latvia", "Lebanon", "Lesotho", "Liberia", "Libya", "Liechtenstein", "Lithuania", "Luxembourg", "Macau", "Macedonia [FYROM]", "Madagascar", "Malawi", "Malaysia", "Maldives", "Mali", "Malta", "Marshall Islands", "Martinique", "Mauritania", "Mauritius", "Mayotte", "Mexico", "Micronesia", "Moldova", "Monaco", "Mongolia", "Montenegro", "Montserrat", "Morocco", "Mozambique", "Myanmar [Burma]", "Namibia", "Nauru", "Nepal", "Netherlands", "Netherlands Antilles", "New Caledonia", "New Zealand", "Nicaragua", "Niger", "Nigeria", "Niue", "Norfolk Island", "North Korea", "Northern Mariana Islands", "Norway", "Oman", "Pakistan", "Palau", "Palestinian Territories", "Panama", "Papua New Guinea", "Paraguay", "Peru", "Philippines", "Pitcairn Islands", "Poland", "Portugal", "Puerto Rico", "Qatar", "Romania", "Russia", "Rwanda", "Réunion", "Saint Helena", "Saint Kitts and Nevis", "Saint Lucia", "Saint Pierre and Miquelon", "Saint Vincent and the Grenadines", "Samoa", "San Marino", "Saudi Arabia", "Senegal", "Serbia", "Seychelles", "Sierra Leone", "Singapore", "Slovakia", "Slovenia", "Solomon Islands", "Somalia", "South Africa", "South Georgia and the South Sandwich Islands", "South Korea", "Spain", "Sri Lanka", "Sudan", "Suriname", "Svalbard and Jan Mayen", "Swaziland", "Sweden", "Switzerland", "Syria", "São Tomé and Príncipe", "Taiwan", "Tajikistan", "Tanzania", "Thailand", "Timor-Leste", "Togo", "Tokelau", "Tonga", "Trinidad and Tobago", "Tunisia", "Turkey", "Turkmenistan", "Turks and Caicos Islands", "Tuvalu", "U.S. Minor Outlying Islands", "U.S. Virgin Islands", "Uganda", "Ukraine", "United Arab Emirates", "United Kingdom", "United States", "Uruguay", "Uzbekistan", "Vanuatu", "Vatican City", "Venezuela", "Vietnam", "Wallis and Futuna", "Western Sahara", "Yemen", "Zambia", "Zimbabwe"}

var TxTypes = map[int]string{
	// Новый юзер
	1: "NewUser",
	// Новый майнер (тоже самое, что "шлюз" в Ripple)
	2: "NewMiner",
	// Добавление новой обещанной суммы (тоже самое, что "лимиты" в Ripple)
	3: "NewPromisedAmount",
	// Изменение лимита обещанной суммы (Лимита)
	4: "ChangePromisedAmount",
	// Голос за претендента на майнера (Шлюз)
	5: "VotesMiner",
	// Новый ордер на продажу или покупку валюты на внутренней децентрализованной бирже
	6: "NewForexOrder",
	// Удаление своего ордера на бирже
	7: "DelForexOrder",
	// Новый набор max_other_currencies от нода-генератора блока
	8: "NewMaxOtherCurrencies",
	// Майнер изменил свои координаты
	9: "ChangeGeolocation",
	// Голосование за возможность добавить новую обещанную сумму (Лимит)
	10: "VotesPromisedAmount",
	// Удаление обещанной суммы (Лимита)
	11: "DelPromisedAmount",
	// Отправка монет другому пользователю
	12: "SendDc",
	// Отправка запроса на обмен фиата на указанную ранее майнером (Шлюзом) обещанную сумму (Лимит)
	13: "CashRequestOut",
	// Удовлетворение запроса CashRequestOut
	14: "CashRequestIn",
	// Голосование за параметры по валютам, имеющимся у майнера (Шлюза)
	15: "VotesComplex",
	// Изменение праймари ключа
	16: "ChangePrimaryKey",
	// Изменение ключа ноды
	17: "ChangeNodeKey",
	// Исправление проблемы с зависнувшим статусом for_repaid юзера.
	18: "ForRepaidFix",
	// занесение в БД данных из первого блока
	19: "Admin1Block",
	// Админ разжаловал майнера в юзеры (отключил недобросовестный Шлюз)
	20: "AdminBanMiners",
	// Админ изменил variables
	21: "AdminVariables",
	// Админ обновил набор точек для проверки лиц
	22: "AdminSpots",
	// Юзер создал кредит
	23: "NewCredit",
	// Админ вернул майнерам звание "майнер" (вернул отключенный Шлюз в работу)
	24: "AdminUnbanMiners",
	// Админ отправил alert message
	25: "AdminMessage",
	// Майнер хочет, чтобы указанные им майнеры были разжалованы в юзеры (жалобы на недобросовестные шлюзы)
	26: "Abuses",
	// Майнер (Шлюз) хочет, чтобы в указанные дни ему не приходили запросы на обмен DC
	27: "NewHolidays",
	// Актуализация обещанной суммы (шлюза)
	28: "ActualizationPromisedAmounts",
	// Перевод средств с обещанных сумм (Лимитов) на свой расчетный счет внутри Dcoin
	29: "Mining",
	// Голосование нода за фото нового майнера
	30: "VotesNodeNewMiner",
	// Юзер исправил проблему с отдачей фото и шлет повторный запрос на получение статуса "майнер" (Шлюз)
	31: "NewMinerUpdate",
	//  Новый набор max_promised_amount от нода-генератора блока
	32: "NewMaxPromisedAmounts",
	//  Новый набор % от нода-генератора блока
	33: "NewPct",
	// Добавление новой валюты
	34: "AdminAddCurrency",
	// Новый краудфандинг проект
	35: "NewCfProject",
	// Новая версия ПО, которая кладется каждому в диру public
	36: "AdminNewVersion",
	// После того, как новая версия протестируется, выдаем сообщение, что необходимо обновиться
	37: "AdminNewVersionAlert",
	// Баг репорты
	38: "MessageToAdmin",
	// Админ может ответить юзеру
	39: "AdminAnswer",
	// Заполненение описания для краудфандинг-проекта
	40: "CfProjectData",
	// Блог админа
	41: "AdminBlog",
	// Майнер (Шлюз) меняет свой хост
	42: "ChangeHost",
	// Майнер (Шлюз) меняет комиссию, которую он хочет получать с тр-ий
	43: "ChangeCommission",
	// Отмена своего взноса в краудфандинг-проект
	44: "DelCfFunding",
	// Запуск урезания на основе голосования или если созданных моент стало больше, чем обещанных сумм (Лимитов). Генерит нод-генератор блока
	45: "NewReduction",
	// Удаление своего краудфандинг проекта
	46: "DelCfProject",
	// Добавление коммента в краудфандинг проект
	47: "CfComment",
	// Отправление средства в краудфандинг проект
	48: "CfSendDc",
	// Смена аватарки у юзера
	49: "UserAvatar",
	// Изменение категории краудфандинг проекта
	50: "CfProjectChangeCategory",
	// Передача чужих обстоятельств другому юзеру (продажа кредита)
	51: "ChangeCreditor",
	// Удаление кредита (может сделать только кредитор)
	52: "DelCredit",
	// Погашение кредита
	53: "RepaymentCredit",
	// Изменение % от поступающих средств, которые будут тратиться на погашение кредита (можно только увеличить)
	54: "ChangeCreditPart",
	// Выборы нового Админа
	55: "NewAdmin",
	// По истечении 30 дней после поступления запроса о восстановлении утерянного ключа, админ может изменить ключ юзера
	56: "AdminChangePrimaryKey",
	// Юзер разрешает или отменяет разрешение на смену своего ключа админом
	57: "ChangeKeyActive",
	// Юзер отменяет запрос на смену ключа
	58: "ChangeKeyClose",
	// Юзер отправляет с другого акка запрос на получение доступа к акку, ключ к которому он потерял
	59: "ChangeKeyRequest",
	// Юзер решил стать арбитром или же действующий арбитр меняет комиссии
	60: "ChangeArbitratorConditions",
	// Продавец меняет % и кол-во дней для новых сделок.
	61: "ChangeSellerHoldBack",
	// Покупатель или продавец указал список арбитров, кому доверяет
	62: "ChangeArbitratorList",
	// Покупатель хочет манибэк
	63: "MoneyBackRequest",
	// Магазин добровольно делает манибэк или же арбитр делает манибек
	64: "MoneyBack",
	// Арбитр увеличивает время манибэка, чтобы успеть разобраться в ситуации
	65: "ChangeMoneyBackTime",
	// Юзер меняет url центров сертификации, где хранятся его приватные ключи
	66: "ChangeCa",
	// Бан юзера в чате
	67: "AdminBanUnbanChat",
	// Голосование за биржи
	68: "VotesExchange",
	// Создание авто-платежа
	69: "NewAutoPayment",
	// Выполнение авто-платежа с ключом ноды
	70: "AutoPayment",
	// Удаление авто-платежа
	71: "DelAutoPayment",
	// Админ пула включает или выключает режим пула
	72: "SwitchPool",
	// Админ пула удаляет юзера со своего пула
	73: "DelUserFromPool",
	// Юзер меняет пул
	74: "ChangePool",
	// Новая ограниченная обещанная сумма (бонус-лимит)
	75: "NewRestrictedPromisedAmount",
	// Добавление акка соц. сети
	76: "UpgradeUser",
	// Голоса майнеров за нового с указанной соц. сетью
	77: "VotesSnUser",
	// Снятие набежавших dUSD с ограниченной обещанной (бонус-лимита)
	78: "MiningSn",
}

func init() {
}

var MaxGreen = map[int64]int64{
	1:100,
	2:500,
	3:10000,
	4:500,
	5:100,
	6:100,
	7:10000,
	8:100,
	9:1000,
	10:200,
	11:1000000,
	12:100,
	13:100,
	14:50000,
	15:500,
	16:200000,
	17:50000,
	18:2000,
	19:500,
	20:5000,
	21:10000,
	22:1000,
	23:100,
	24:100,
	25:200,
	26:200,
	27:1000,
	28:1000,
	29:500,
	30:20000,
	31:1000000,
	32:500,
	33:5000,
	34:100000,
	35:20000000,
	36:100,
	37:10000,
	38:10000,
	39:100000,
	40:100,
	41:20000,
	42:200000,
	43:10000,
	44:1000,
	45:1000,
	46:500,
	47:20000,
	48:500,
	49:10000,
	50:100,
	51:500,
	52:5000,
	53:10000,
	54:500,
	55:500,
	56:200,
	57:10000,
	58:5000,
	59:500,
	60:500,
	61:500,
	62:100,
	63:1000,
	64:10000,
	65:5000,
	66:200,
	67:200,
	68:2000,
	69:100000,
	70:1000,
	71:200000,
	72:100,
	73:200000,
	74:500,
	75:2000000,
	76:20000,
	77:1000,
}


var DCTarget = map[int64]int64{
	72:3000000000000,
	58:16000000000000,
	23:3000000000000,
}