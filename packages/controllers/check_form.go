package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
	"fmt"
	"strings"
)

type listCheck struct {
	Action string
	Id     string
	Value  string
	Label  string
}

type warning struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

type outCheck struct {
	Warnings []*warning `json:"warnings"`
}

type inCheckForm struct {
	List []*listCheck `json:"list"`
}

func (c *Controller) CheckForm() (string, error) {

	var (
		input inCheckForm
		out   outCheck
	)
	resval := true
	result := func(msg string) (string, error) {
		var ret answerJson
		ret.Result = resval
		if data, err := json.Marshal(out); err == nil {
			ret.Data = string(data)
			ret.Error = msg
		} else {
			ret.Error = err.Error()
		}
		if len(ret.Error) == 0 {
			ret.Success = true
		}
		res, err := json.Marshal(ret)
		return string(res), err
	}
	input.List = make([]*listCheck, 0)
	out.Warnings = make([]*warning, 0)
	if err := json.Unmarshal([]byte(c.r.FormValue(`tocheck`)), &input); err != nil {
		return result(err.Error())
	}
	for _, item := range input.List {
		var (
			pars []string
			text string
		)
		if strings.IndexRune(item.Action, '|') > 0 {
			pars = strings.Split(item.Action, `|`)
			item.Action = pars[0]
		}
		switch item.Action {
		case `empty`:
			if len(item.Value) == 0 {
				text = fmt.Sprintf(c.Lang[`chform_empty`], item.Label)
			}
		case `zero`:
			if utils.StrToFloat64(item.Value) == 0 {
				text = fmt.Sprintf(c.Lang[`chform_zero`], item.Label)
			}
		case `interval`:
			if len(pars) == 3 {
				val := utils.StrToInt64(item.Value)
				if val < utils.StrToInt64(pars[1]) || val > utils.StrToInt64(pars[2]) {
					text = fmt.Sprintf(c.Lang[`chform_interval`], item.Label, pars[1], pars[2])
				}
			}
		case `userid`:
			if len(item.Value) == 0 {
				text = fmt.Sprintf(c.Lang[`chform_empty`], item.Label)
			} else {
				userid := utils.StrToInt64(item.Value)
				if userid <= 0 {
					text = c.Lang[`chform_userid`]
				} else if err := c.DCDB.CheckUser(userid); err != nil {
					text = c.Lang[`chform_userid`]
				}
			}
		}
		if len(text) > 0 {
			out.Warnings = append(out.Warnings, &warning{Id: item.Id, Text: text})
			resval = false
		}
	}

	return result(``)
}
