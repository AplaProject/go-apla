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

package consts

// текущая версия
// Current version
const VERSION = "0.1.6b9"
const BLOCK_VERSION = 1

const FIRST_QDLT = 1e+26
const EGS_DIGIT = 18 //money_digit for EGS 1000000000000000000

// используется в confirmations
// is used in confirmations
const COUNT_CONFIRMED_NODES = 5
const WAIT_CONFIRMED_NODES = 10
const MAX_TX_SIZE = 32 << 20
const GAPS_BETWEEN_BLOCKS = 3

// У скольких нодов должен быть такой же блок как и у нас, чтобы считать, что блок у большей части DC-сети. для get_confirmed_block_id()
// The number of nodes which should have the same block as we have for regarding this block belongs to the major part of DC-net. For get_confirmed_block_id()
const MIN_CONFIRMED_NODES = 0

// примерный текущий крайний блок
// Approximate current last block
const LAST_BLOCK = 330000

// примерный размер блокчейна
// Approximate blockchain size
const BLOCKCHAIN_SIZE = 1000
const DOWNLOAD_CHAIN_TRY_COUNT = 10

// на сколько может бежать время в тр-ии
// How fast could the time of transaction pass
const MAX_TX_FORW = 0

// тр-ия может блуждать по сети сутки и потом попасть в блок
// transaction may wander in the net for a day and then get into a block
const MAX_TX_BACK = 86400

const ERROR_TIME = 1

const ROUND_FIX = 0.00000000001

// таймауты для TCP
// timeouts for TCP
const READ_TIMEOUT = 20
const WRITE_TIMEOUT = 20

// дефолтное знаение, со скольким нодами устанавляиваем связь
// default value, with how many nodes we make the connection
const OUT_CONNECTIONS = 10

const COMMISSION = 1000
const TCP_PORT = "7078"
const RB_BLOCKS_1 = 30
const RB_BLOCKS_2 = 1440
const ALERT_ERROR_TIME = 1

const DATA_TYPE_MAX_BLOCK_ID = 10
const DATA_TYPE_BLOCK_BODY = 7

const CHANGE_KEY_PERIOD = 86400 * 7

const RATE_VOTING_PERIOD = 3600

const COUNT_BLOCK_BEFORE_SAVE = RB_BLOCKS_2

const ALERT_KEY = `30820122300d06092a864886f70d01010105000382010f003082010a0282010100d4a48242d0fb2c7c295bc9c87b1aa0c6d23b5f8cab2ec20c2dfde35513ef6066b92ee3935f9a38100493717b60bb7832411daee02012f44a9f58ac91056b2603661544116bfbc55181e5a693bace5ec9325ba0232b9c9c0a29096569d217243e5bf891cc7fc4bcd2e7d6518acc6f982aaa43a9ed737e3ea2845d6432a823ee5b40d1548f802d0c108bf6e5cb5a4daa7edb48764dcbfa6b7a961208833996cfee265ca2ce2655d444cf3c177b3841b1cc4f3102f89cb2bdb1e5a68eac270506147dd8391b7b3af40a50be13c3970077faffaf98ccc5b8c011146be9c2eb9dfd3454f67a68daaf385d334366d132308bffede27656a515ff69a260bbe2452bd2c30203010001`
const UPD_AND_VER_URL = "http://apla.io"
const GOOGLE_API_KEY = "AIzaSyBLZlUPgd9uhX05OrsFU68yJOZFrYhZe84"

const MainEco = 1

// LangMap contains supported languages
var LangMap = map[string]int{"en": 1, "ru": 42}

// Countries is the list of the countries
var Countries = []string{"Afghanistan", "Albania", "Algeria", "American Samoa", "Andorra", "Angola", "Anguilla", "Antarctica", "Antigua and Barbuda", "Argentina", "Armenia", "Aruba", "Australia", "Austria", "Azerbaijan", "Bahamas", "Bahrain", "Bangladesh", "Barbados", "Belarus", "Belgium", "Belize", "Benin", "Bermuda", "Bhutan", "Bolivia", "Bosnia and Herzegovina", "Botswana", "Bouvet Island", "Brazil", "British Indian Ocean Territory", "British Virgin Islands", "Brunei", "Bulgaria", "Burkina Faso", "Burundi", "Cambodia", "Cameroon", "Canada", "Cape Verde", "Cayman Islands", "Central African Republic", "Chad", "Chile", "China", "Christmas Island", "Cocos [Keeling] Islands", "Colombia", "Comoros", "Congo [DRC]", "Congo [Republic]", "Cook Islands", "Costa Rica", "Croatia", "Cuba", "Cyprus", "Czech Republic", "Côte d\"Ivoire", "Denmark", "Djibouti", "Dominica", "Dominican Republic", "Ecuador", "Egypt", "El Salvador", "Equatorial Guinea", "Eritrea", "Estonia", "Ethiopia", "Falkland Islands [Islas Malvinas]", "Faroe Islands", "Fiji", "Finland", "France", "French Guiana", "French Polynesia", "French Southern Territories", "Gabon", "Gambia", "Gaza Strip", "Georgia", "Germany", "Ghana", "Gibraltar", "Greece", "Greenland", "Grenada", "Guadeloupe", "Guam", "Guatemala", "Guernsey", "Guinea", "Guinea-Bissau", "Guyana", "Haiti", "Heard Island and McDonald Islands", "Honduras", "Hong Kong", "Hungary", "Iceland", "India", "Indonesia", "Iran", "Iraq", "Ireland", "Isle of Man", "Israel", "Italy", "Jamaica", "Japan", "Jersey", "Jordan", "Kazakhstan", "Kenya", "Kiribati", "Kosovo", "Kuwait", "Kyrgyzstan", "Laos", "Latvia", "Lebanon", "Lesotho", "Liberia", "Libya", "Liechtenstein", "Lithuania", "Luxembourg", "Macau", "Macedonia [FYROM]", "Madagascar", "Malawi", "Malaysia", "Maldives", "Mali", "Malta", "Marshall Islands", "Martinique", "Mauritania", "Mauritius", "Mayotte", "Mexico", "Micronesia", "Moldova", "Monaco", "Mongolia", "Montenegro", "Montserrat", "Morocco", "Mozambique", "Myanmar [Burma]", "Namibia", "Nauru", "Nepal", "Netherlands", "Netherlands Antilles", "New Caledonia", "New Zealand", "Nicaragua", "Niger", "Nigeria", "Niue", "Norfolk Island", "North Korea", "Northern Mariana Islands", "Norway", "Oman", "Pakistan", "Palau", "Palestinian Territories", "Panama", "Papua New Guinea", "Paraguay", "Peru", "Philippines", "Pitcairn Islands", "Poland", "Portugal", "Puerto Rico", "Qatar", "Romania", "Russia", "Rwanda", "Réunion", "Saint Helena", "Saint Kitts and Nevis", "Saint Lucia", "Saint Pierre and Miquelon", "Saint Vincent and the Grenadines", "Samoa", "San Marino", "Saudi Arabia", "Senegal", "Serbia", "Seychelles", "Sierra Leone", "Singapore", "Slovakia", "Slovenia", "Solomon Islands", "Somalia", "South Africa", "South Georgia and the South Sandwich Islands", "South Korea", "Spain", "Sri Lanka", "Sudan", "Suriname", "Svalbard and Jan Mayen", "Swaziland", "Sweden", "Switzerland", "Syria", "São Tomé and Príncipe", "Taiwan", "Tajikistan", "Tanzania", "Thailand", "Timor-Leste", "Togo", "Tokelau", "Tonga", "Trinidad and Tobago", "Tunisia", "Turkey", "Turkmenistan", "Turks and Caicos Islands", "Tuvalu", "U.S. Minor Outlying Islands", "U.S. Virgin Islands", "Uganda", "Ukraine", "United Arab Emirates", "United Kingdom", "United States", "Uruguay", "Uzbekistan", "Vanuatu", "Vatican City", "Venezuela", "Vietnam", "Wallis and Futuna", "Western Sahara", "Yemen", "Zambia", "Zimbabwe"}

// TxTypes is the list of the embedded transactions
var TxTypes = map[int]string{
	1:  "FirstBlock",
	2:  "Reserved1",
	3:  "Reserved2",
	4:  "Reserved3",
	5:  "DLTTransfer",
	6:  "DLTChangeHostVote",
	7:  "UpdFullNodes",
	8:  "ChangeNodeKey",
	9:  "NewState",
	10: "NewColumn",
	11: "NewTable",
	12: "EditPage",
	13: "EditMenu",
	14: "EditContract",
	15: "NewContract",
	16: "EditColumn",
	17: "EditTable",
	18: "EditStateParameters",
	19: "NewStateParameters",
	20: "NewPage",
	21: "NewMenu",
	22: "ChangeNodeKeyDLT",
	23: "AppendPage",
	24: "RestoreAccessActive",
	25: "RestoreAccessClose",
	26: "RestoreAccessRequest",
	27: "RestoreAccess",
	28: "NewLang",
	29: "EditLang",
	30: "AppendMenu",
	31: "NewSign",
	32: "EditSign",
	33: "EditWallet",
	34: "ActivateContract",
	35: "NewAccount",
}

var FillSize = 32

func init() {
}
