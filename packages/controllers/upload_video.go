package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

func (c *Controller) UploadVideo() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	var binaryVideo []byte

	c.r.ParseMultipartForm(32 << 20)
	file, _, err := c.r.FormFile("file")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	videoBuffer := new(bytes.Buffer)
	_, err = io.Copy(videoBuffer, file)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer file.Close()
	binaryVideo = videoBuffer.Bytes()
	fmt.Println(c.r.MultipartForm.File["file"][0].Filename)
	fmt.Println(c.r.MultipartForm.File["file"][0].Header.Get("Content-Type"))
	fmt.Println(c.r.MultipartForm.Value["type"][0])

	var contentType, videoType string
	if _, ok := c.r.MultipartForm.File["file"]; ok {
		contentType = c.r.MultipartForm.File["file"][0].Header.Get("Content-Type")
	}
	if _, ok := c.r.MultipartForm.Value["type"]; ok {
		videoType = c.r.MultipartForm.Value["type"][0]
	}
	end := "mp4"
	switch contentType {
	case "video/mp4", "video/quicktime":
		end = "mp4"
	case "video/ogg":
		end = "ogv"
	case "video/webm":
		end = "webm"
	case "video/3gpp":

		fmt.Println("3gpp")
		conn, err := net.DialTimeout("tcp", "3gp.dcoin.club:8099", 5*time.Second)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		defer conn.Close()

		conn.SetReadDeadline(time.Now().Add(240 * time.Second))
		conn.SetWriteDeadline(time.Now().Add(240 * time.Second))

		// в 4-х байтах пишем размер данных, которые пошлем далее
		size := utils.DecToBin(len(videoBuffer.Bytes()), 4)
		_, err = conn.Write(size)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		// далее шлем сами данные
		_, err = conn.Write(videoBuffer.Bytes())
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		// в ответ получаем размер данных, которые нам хочет передать сервер
		buf := make([]byte, 4)
		n, err := conn.Read(buf)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		log.Debug("dataSize buf: %x / get: %v", buf, n)

		// и если данных менее 10мб, то получаем их
		dataSize := utils.BinToDec(buf)
		var binaryBlock []byte
		log.Debug("dataSize: %v", dataSize)
		if dataSize < 10485760 && dataSize > 0 {
			binaryBlock = make([]byte, dataSize)
			//binaryBlock, err = ioutil.ReadAll(conn)
			_, err = io.ReadFull(conn, binaryBlock)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			log.Debug("len(binaryBlock):", len(binaryBlock))
			binaryVideo = binaryBlock
		}
	}

	log.Debug(videoType, end)

	var name string
	if videoType == "user_video" {
		name = "public/" + utils.Int64ToStr(c.SessUserId) + "_user_video." + end
	} else {
		x := strings.Split(videoType, "-")
		if len(x) < 2 {
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
		name = "public/" + utils.Int64ToStr(c.SessUserId) + "_promised_amount_" + x[1] + "." + end
	}
	log.Debug(*utils.Dir + "/" + name)
	err = ioutil.WriteFile(*utils.Dir+"/"+name, binaryVideo, 0644)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return utils.JsonAnswer(string(utils.DSha256(binaryVideo)), "success").String(), nil
}
