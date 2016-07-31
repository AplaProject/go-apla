package controllers

import (
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"net/http"
	"image"
	"image/draw"
	"image/png"
	"fmt"
	"github.com/golang/freetype"
	"regexp"
)


func IndexStat(w http.ResponseWriter, r *http.Request) {

	fmt.Println("IndexStat");
	if utils.DB != nil && utils.DB.DB != nil {

		c := new(Controller)
		c.r = r
		c.w = w
		c.DCDB = utils.DB

		r.ParseForm()

		var userId int64
		re := regexp.MustCompile(`([0-9]+)`)
		match := re.FindStringSubmatch(c.r.URL.RequestURI())
		if len(match) != 0 {
			userId = utils.StrToInt64(match[1])
		}

		w, h := 500, 300
		data, _ := static.Asset("static/img/stat.png")
		fSrc := bytes.NewReader(data)

		src, err := png.Decode(fSrc)
		if err != nil {
			log.Error("utils.DB == nil")
		}

		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		//white := image.NewUniform(color.White)
		//black := image.NewUniform(color.Black)
		draw.Draw(dst, dst.Bounds(), src, image.Point{0, 0}, draw.Src)

		fontBytes, err := static.Asset("static/fonts/luxisr.ttf")
		if err != nil {
			log.Error("utils.DB == nil")
		}
		font, err := freetype.ParseFont(fontBytes)
		if err != nil {
			log.Error("utils.DB == nil")
		}

		imText := freetype.NewContext()
		imText.SetDPI(72)
		imText.SetFont(font)
		imText.SetFontSize(15)
		imText.SetClip(dst.Bounds())
		imText.SetDst(dst)
		imText.SetSrc(image.Black)

		// Draw the text.
		pt := freetype.Pt(13, 300)
		_, err = imText.DrawString("User ID: "+utils.Int64ToStr(userId), pt)
		if err != nil {
			log.Error("utils.DB == nil")
		}

		buffer := new(bytes.Buffer)
		err = png.Encode(buffer, dst)
		if err != nil {
			log.Error("utils.DB == nil")
		}
		c.w.Header().Set("Content-Type", "image/png")
		c.w.Header().Set("Content-Length", utils.IntToStr(len(buffer.Bytes())))
		if _, err := c.w.Write(buffer.Bytes()); err != nil {
			log.Error("utils.DB == nil")
		}

	}
}