package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
	"bytes"
	"html/template"
	"strings"
	"unicode"
)

type CommonPage struct {
	//Lang   map[string]string
	Test string
}

type pageTemplate struct {
	Template  string
	Name      string
}

var (
	globPages = make(map[string]*pageTemplate)
)

func newPage(name string) {
	page := pageTemplate{Template: name}
	parts := strings.Split(name, `_`)
	for i := range parts {
		if i>0 {
			a := []rune(parts[i])
			a[0] = unicode.ToUpper(a[0])
			parts[i] = string(a)
		}
	}
	page.Name = strings.Join(parts,``)
	globPages[name] = &page
}

func isPage(name string) bool {
	_,ok := globPages[name]
	return ok
}

func CallPage(c *Controller, name string) string {
	a := []rune(name)
	a[0] = unicode.ToUpper(a[0])
	name = string(a)
	html, err := CallMethod(c, name)
	if err != nil {
		html = fmt.Sprintf(`{"error":%q}`, err)
		log.Error("err: %s / Controller: %s", html, name)
	}
	return html
}

func proceedTemplate(c *Controller, html string, data interface{}) (string, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("proceedTemplate Recovered", r)
			fmt.Println(r)
		}
	}()
	pattern, err := static.Asset("static/" + html + ".html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
/*	funcMap := template.FuncMap{
		"makeCurrencyName": func(currencyId int64) string {
			if currencyId >= 1000 {
				return ""
			} else {
				return "d"
			}
		},
		"div": func(a, b interface{}) float64 {
			return utils.InterfaceToFloat64(a) / utils.InterfaceToFloat64(b)
		},
		"mult": func(a, b interface{}) float64 {
			return utils.InterfaceToFloat64(a) * utils.InterfaceToFloat64(b)
		},
		"round": func(a interface{}, num int) float64 {
			return utils.Round(utils.InterfaceToFloat64(a), num)
		},
		"len": func(s []map[string]string) int {
			return len(s)
		},
		"lenMap": func(s map[string]string) int {
			return len(s)
		},
		"sum": func(a, b interface{}) float64 {
			return utils.InterfaceToFloat64(a) + utils.InterfaceToFloat64(b)
		},
		"minus": func(a, b interface{}) float64 {
			return utils.InterfaceToFloat64(a) - utils.InterfaceToFloat64(b)
		},
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
		"js": func(s string) template.JS {
			return template.JS(s)
		},
		"join": func(s []string, sep string) string {
			return strings.Join(s, sep)
		},
		"strToInt64": func(text string) int64 {
			return utils.StrToInt64(text)
		},
		"strToInt": func(text string) int {
			return utils.StrToInt(text)
		},
		"bin2hex": func(text string) string {
			return string(utils.BinToHex([]byte(text)))
		},
		"int64ToStr": func(text int64) string {
			return utils.Int64ToStr(text)
		},
		"intToStr": func(text int) string {
			return utils.IntToStr(text)
		},
		"intToInt64": func(text int) int64 {
			return int64(text)
		},
		"rand": func() int {
			return utils.RandInt(0, 99999999)
		},
		"append": func(args ...interface{}) string {
			var result string
			for _, value := range args {
				switch value.(type) {
				case int64:
					result += utils.Int64ToStr(value.(int64))
				case float64:
					result += utils.Float64ToStr(value.(float64))
				case string:
					result += value.(string)
				}
			}
			return result
		},
		"replaceCurrency": func(text, name string) string {
			return strings.Replace(text, "[currency]", name, -1)
		},
		"replaceCurrencyName": func(text, name string) string {
			return strings.Replace(text, "[currency]", "D"+name, -1)
		},
		"cfCategoryLang": func(lang map[string]string, name string) string {
			return lang["cf_category_"+name]
		},
		"progressBarLang": func(lang map[string]string, name string) string {
			return lang["progress_bar_pct_"+name]
		},
		"checkProjectPs": func(ProjectPs map[string]string, id string) bool {
			if len(ProjectPs["ps"+id]) > 0 {
				return true
			} else {
				return false
			}
		},
		"cfPageTypeLang": func(lang map[string]string, name string) string {
			return lang["cf_"+name]
		},
		"notificationsLang": func(lang map[string]string, name string) string {
			return lang["notifications_"+name]
		},
		"issuffix": func(text, name string) bool {
			return strings.HasSuffix(text,name)
		},

	}
	t := template.Must(template.New("template").Funcs(funcMap).Parse(string(data)))*/
//	t = template.Must(t.Parse(string(alert_success)))
//	t = template.Must(t.Parse(string(signatures)))
/*	t := template.New("template").Funcs(funcMap)
	t, err = t.Parse(string(data))
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		}
		
		b := new(bytes.Buffer)
		err = t.Execute(b, c)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		}
		w.Write(b.Bytes())
		*/
	t := template.Must(template.New("template").Parse(string(pattern)))
	b := new(bytes.Buffer)
	err = t.Execute(b, data)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return b.String(), nil
}
