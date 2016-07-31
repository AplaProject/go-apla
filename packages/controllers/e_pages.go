package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type ePages struct {
	Lang  map[string]string
	Title string
	Text  string
}

func (c *Controller) EPages() (string, error) {

	var err error

	var title, text string
	if len(c.Parameters["page"]) > 0 {
		if !utils.CheckInputData(c.Parameters["page"], "string") {
			return "", utils.ErrInfo(err)
		}
		data, err := c.OneRow(`SELECT title, text from e_pages WHERE name = ? AND lang = ?`, c.Parameters["page"], c.LangInt).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		title = data["title"]
		text = data["text"]
	}

	TemplateStr, err := makeTemplate("e_pages", "ePages", &ePages{
		Lang:  c.Lang,
		Title: title,
		Text:  text})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
