// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package template

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/types"

	log "github.com/sirupsen/logrus"
)

const (
	columnTypeText     = "text"
	columnTypeLongText = "long_text"
	columnTypeBlob     = "blob"

	substringLength = 32
)

func dbfindExpressionBlob(column string) string {
	return fmt.Sprintf(`md5(%s) "%[1]s"`, column)
}

func dbfindExpressionLongText(column string) string {
	return fmt.Sprintf(`json_build_array(
		substr(%s, 1, %d),
		CASE WHEN length(%[1]s)>%[2]d THEN md5(%[1]s) END) "%[1]s"`, column, substringLength)
}

type valueLink struct {
	title string

	id     string
	table  string
	column string
	hash   string
}

func (vl *valueLink) link() string {
	if len(vl.hash) > 0 {
		return fmt.Sprintf("/data/%s/%s/%s/%s", vl.table, vl.id, vl.column, vl.hash)
	}
	return ""
}

func (vl *valueLink) marshal() (string, error) {
	b, err := json.Marshal(map[string]string{
		"title": vl.title,
		"link":  vl.link(),
	})
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling valueLink to JSON")
		return "", err
	}
	return string(b), nil
}

func trimString(in []rune) string {
	out := strings.TrimSpace(string(in))
	if len(out) > 0 && out[0] == '"' && out[len(out)-1] == '"' {
		out = out[1 : len(out)-1]
	}
	return out
}

func parseObject(in []rune) (interface{}, int) {
	var (
		ret            interface{}
		key            string
		mapMode, quote bool
	)

	length := len(in)
	if in[0] == '[' {
		ret = make([]interface{}, 0)
	} else if in[0] == '{' {
		ret = types.NewMap()
		mapMode = true
	} else {
		return nil, 0
	}
	start := 1
	i := 1
main:
	for ; i < length; i++ {
		ch := in[i]
		if quote && ch != '"' {
			continue
		}
		switch ch {
		case ']':
			if !mapMode {
				break main
			}
		case '}':
			if mapMode {
				break main
			}
		case '{', '[':
			par, off := parseObject(in[i:])
			if mapMode {
				if len(key) == 0 {
					switch v := par.(type) {
					case map[string]interface{}:
						for ikey, ival := range v {
							ret.(*types.Map).Set(ikey, ival)
						}
					}
				} else {
					ret.(*types.Map).Set(key, par)
					key = ``
				}
			} else {
				if len(key) > 0 {
					par = map[string]interface{}{key: par}
					key = ``
				}
				ret = append(ret.([]interface{}), par)
			}
			i += off
			start = i + 1
		case '"':
			quote = !quote
		case ':':
			if len(key) == 0 {
				key = trimString(in[start:i])
				start = i + 1
			}
		case ',':
			val := trimString(in[start:i])
			if len(val) == 0 && (len(key) > 0 || mapMode) {
				key = ``
			}
			if len(val) > 0 {
				if mapMode {
					ret.(*types.Map).Set(key, val)
					key = ``
				} else {
					if len(key) > 0 {
						ret = append(ret.([]interface{}), map[string]interface{}{key: val})
						key = ``
					} else {
						ret = append(ret.([]interface{}), val)
					}
				}
			}
			start = i + 1
		}
	}
	if start < i {
		if last := trimString(in[start:i]); len(last) > 0 {
			if mapMode {
				ret.(*types.Map).Set(key, last)
			} else {
				if len(key) > 0 {
					ret = append(ret.([]interface{}), map[string]interface{}{key: last})
					key = ``
				} else {
					ret = append(ret.([]interface{}), last)
				}
			}
		} else if len(key) > 0 || mapMode {
			ret.(*types.Map).Set(key, "")
		}
	}
	switch v := ret.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			ret = ``
		}
	case []interface{}:
		if len(v) == 0 {
			ret = ``
		}
	}
	return ret, i
}
