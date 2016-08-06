package controllers

import (
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"math"
	"regexp"
	"strings"
	"fmt"
)

func (c *Controller) AlertMessage() (string, error) {

	if c.SessRestricted != 0 {
		return "", nil
	}

	c.r.ParseForm()
	if ok, _ := regexp.MatchString(`install`, c.r.FormValue("tpl_name")); ok {
		return "", nil
	}

	show := false
	// проверим, есть ли сообщения от админа
	data, err := c.OneRow("SELECT * FROM alert_messages WHERE close  =  0 ORDER BY id DESC").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	var message map[string]string
	var adminMessage string
	if len(data["message"]) > 0 {
		err = json.Unmarshal([]byte(data["message"]), &message)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(message[utils.Int64ToStr(c.LangInt)]) > 0 {
			adminMessage = message[utils.Int64ToStr(c.LangInt)]
		} else {
			adminMessage = message["gen"]
		}
		if data["currency_list"] != "ALL" {
			// проверим, есть ли у нас обещанные суммы с такой валютой
			promisedAmount, err := c.Single("SELECT id FROM promised_amount WHERE currency_id IN (" + data["currency_list"] + ")").Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if promisedAmount > 0 {
				show = true
			}
		} else {
			show = true
		}
	}
	result := ""
	if show {
		result += `<script>
			$('#close_alert').bind('click', function () {
				$.post( 'ajax?controllerName=closeAlert', {
					'id' : '` + data["id"] + `'
				} );
			});
			</script>
			 <div class="alert alert-danger alert-dismissable" style='margin-top: 30px'><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
			 <h4>Warning!</h4>
			    ` + adminMessage + `
			  </div>`
	}

	// сообщение о новой версии движка
	myVer, err := c.Single("SELECT current_version FROM info_block").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// возможны 2 сценария:
	// 1. информация о новой версии есть в блоках
	newVer, err := c.GetList("SELECT version FROM new_version WHERE alert  =  1").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	newMaxVer := "0"
	for i := 0; i < len(newVer); i++ {
		if utils.VersionOrdinal(newVer[i]) > utils.VersionOrdinal(myVer) && newMaxVer == "0" {
			newMaxVer = newVer[i]
		}
		if utils.VersionOrdinal(newVer[i]) > utils.VersionOrdinal(newMaxVer) && newMaxVer != "0" {
			newMaxVer = newVer[i]
		}
	}
	var newVersion string
	if newMaxVer != "0" {
		newVersion = strings.Replace(c.Lang["new_version"], "[ver]", newMaxVer, -1)
	}

	// для пулов и ограниченных юзеров выводим сообщение без кнопок
	if (c.Community || c.SessRestricted != 0) && newMaxVer != "0" {
		newVersion = strings.Replace(c.Lang["new_version_pulls"], "[ver]", newMaxVer, -1)
	}

	if newMaxVer != "0" && len(myVer) > 0 {
		result += `<script>
				$('#btn_install').bind('click', function () {
					$('#new_version_text').text('Please wait');
					$.post( 'ajax?controllerName=installNewVersion', {}, function(data) {
						$('#new_version_text').text(data);
					});
				});
				$('#btn_upgrade').bind('click', function () {
					$('#new_version_text').text('Please wait');
					$.post( 'ajax?controllerName=upgradeToNewVersion', {}, function(data) {
						$('#new_version_text').text(data);
					});
				});
			</script>

			<div class="alert alert-danger alert-dismissable" style='margin-top: 30px'><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
			    <h4>Warning!</h4>
			   <div id='new_version_text'>` + newVersion + `</div>
			 </div>`
	}

	if c.SessRestricted == 0 && (!c.Community || c.PoolAdmin) {
		myMinerId, err := c.GetMinerId(c.SessUserId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// если юзер уже майнер, то у него должно быть настроено точное время
		if myMinerId > 0 {
			networkTime, err := utils.GetNetworkTime()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			diff := int64(math.Abs(float64(utils.Time() - networkTime.Unix())))
			if diff > c.Variables.Int64["alert_error_time"] {
				alertTime := strings.Replace(c.Lang["alert_time"], "[sec]", utils.Int64ToStr(diff), -1)
				result += `<div class="alert alert-danger alert-dismissable" style='margin-top: 30px'><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
				     <h4>Warning!</h4>
				     <div>` + alertTime + `</div>
				     </div>`
			}
		}
	}

	if c.SessRestricted == 0 {
		// после обнуления таблиц my_node_key будет пуст
		// получим время из последнего блока
		myNodePrivateKey, err := c.GetNodePrivateKey()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		minerId, err := c.GetMinerId(c.SessUserId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		// возможно юзер новенький на пуле и у него разные нод-ключи
		nodePublicKey, err := c.GetNodePublicKey(c.SessUserId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		myNodePublicKey, err := c.GetMyNodePublicKey(c.MyPrefix)
		fmt.Println("Node public key", (string(nodePublicKey)))
		fmt.Println("My Node public key", myNodePublicKey)

		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if (len(myNodePrivateKey) == 0 && minerId > 0) || (string(nodePublicKey) != myNodePublicKey && len(nodePublicKey) > 0) {
			// Смотрим отправлен ли запрос на смену ключа
			var last_tx []map[string]string
			if last_tx, err = c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangeNodeKey"}), 
		                    1, c.TimeFormat); err != nil {
				return "", utils.ErrInfo(err)
			}
			// Транзакции завершились успешно или с ошибкой
			if len(last_tx) == 0 || last_tx[0][`block_id`] != `0` || len( last_tx[0][`error`] ) > 0 {
				result += `<div class="alert alert-danger alert-dismissable" style='margin-top: 30px'><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
				     <h4>Warning!</h4>
				     <div>` + c.Lang["alert_change_node_key"] + `</div>
				     </div>`
			}
		}
	}

	// просто информируем, что в данном разделе у юзера нет прав
	skipCommunity := []string{"nodeConfig", "nulling", "startStop"}
	skipRestrictedUsers := []string{"nodeConfig", "changeNodeKey", "nulling", "startStop", "cashRequestIn", "cashRequestOut", "upgrade", "notifications", "interface"}
	if (!c.NodeAdmin && utils.InSliceString(c.TplName, skipCommunity)) || (c.SessRestricted != 0 && utils.InSliceString(c.TplName, skipRestrictedUsers)) {
		result += `<div class="alert alert-danger alert-dismissable" style='margin-top: 30px'><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
			  <h4>Warning!</h4>
			  <div>` + c.Lang["permission_denied"] + `</div>
			  </div>`
	}

	// информируем, что у юзера нет прав и нужно стать майнером
	minersOnly := []string{"myCfProjects", "newCfProject", "cashRequestIn", "cashRequestOut", "changeNodeKey", "voting", "geolocation", "promisedAmountList", "promisedAmountAdd", "holidaysList", "newHolidays", "points", "tasks", "changeHost", "newUser", "changeCommission"}
	if utils.InSliceString(c.TplName, minersOnly) {
		minerId, err := c.Single("SELECT miner_id FROM users LEFT JOIN miners_data ON users.user_id  =  miners_data.user_id WHERE users.user_id  =  ?", c.SessUserId).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if minerId == 0 {
			result += `<div class="alert alert-danger alert-dismissable" style='margin-top: 30px'><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
				 <h4>Warning!</h4>
				 <div>` + c.Lang["only_for_miners"] + `</div>
				 </div>`
		}
	}

	// информируем, что необходимо вначале сменить праймари-ключ
	logId, err := c.Single("SELECT log_id FROM users WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if logId == 0 {
		text := ""
		// проверим, есть ли запросы на смену в тр-ях
		lastTx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangePrimaryKey"}), 1, c.TimeFormat)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(lastTx) == 0 { // юзер еще не начинал смену ключа
			text = c.Lang["alert_change_primary_key"]
		} else if len(lastTx[0]["error"]) > 0 || utils.Time()-utils.StrToInt64(lastTx[0]["time_int"]) > 3600 {
			text = c.Lang["please_try_again_change_key"]
		} else {
			text = c.Lang["please_wait_changing_key"]
		}
		result += `<div class="alert alert-danger alert-dismissable" style='margin-top: 30px'><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
				  <h4>` + c.Lang["warning"] + `</h4>
				  <div>` + text + `</div>
				  </div>`
	}

	return result, nil
}
