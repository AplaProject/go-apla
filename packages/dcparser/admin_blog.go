package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) AdminBlogInit() error {

	fields := []map[string]string{{"lng": "string"}, {"title": "string"}, {"message": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminBlogFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.TxMaps.String["title"]) > 255 {
		return p.ErrInfo("len title>255")
	}
	if len(p.TxMaps.String["message"]) > 1048576 {
		return p.ErrInfo("len message>1048576")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["lng"], p.TxMap["title"], p.TxMap["message"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) AdminBlog() error {
	err := p.ExecSql("INSERT INTO admin_blog ( time, lng, title, message ) VALUES ( ?, ?, ?, ? )", p.BlockData.Time, p.TxMaps.String["lng"], p.TxMaps.String["title"], p.TxMaps.String["message"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminBlogRollback() error {
	err := p.ExecSql("DELETE FROM admin_blog WHERE time = ? AND lng = ? AND title = ? AND message = ?", p.BlockData.Time, p.TxMaps.String["lng"], p.TxMaps.String["title"], p.TxMaps.String["message"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("admin_blog", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminBlogRollbackFront() error {
	return nil
}
