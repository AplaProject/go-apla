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

package queryBuilder

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/types"
)

func PrepareWhere(where string) string {
	whereSlice := regexp.MustCompile(`->([\w\d_]+)`).FindAllStringSubmatchIndex(where, -1)
	startWhere := 0
	out := ``
	for i := 0; i < len(whereSlice); i++ {
		slice := whereSlice[i]
		if len(slice) != 4 {
			continue
		}
		if i < len(whereSlice)-1 && slice[1] == whereSlice[i+1][0] {
			colsWhere := []string{where[slice[2]:slice[3]]}
			from := slice[0]
			for i < len(whereSlice)-1 && slice[1] == whereSlice[i+1][0] {
				i++
				slice = whereSlice[i]
				if len(slice) != 4 {
					break
				}
				colsWhere = append(colsWhere, where[slice[2]:slice[3]])
			}
			out += fmt.Sprintf(`%s::jsonb#>>'{%s}'`, where[startWhere:from], strings.Join(colsWhere, `,`))
			startWhere = slice[3]
		} else {
			out += fmt.Sprintf(`%s->>'%s'`, where[startWhere:slice[0]], where[slice[2]:slice[3]])
			startWhere = slice[3]
		}
	}
	if len(out) > 0 {
		return out + where[startWhere:]
	}
	return where
}

func GetWhere(inWhere *types.Map) (string, error) {
	var (
		where string
		cond  []string
	)
	if inWhere == nil {
		inWhere = types.NewMap()
	}
	escape := func(value interface{}) string {
		return strings.Replace(fmt.Sprint(value), `'`, `''`, -1)
	}
	oper := func(action string, v interface{}) (string, error) {
		switch value := v.(type) {
		default:
			return fmt.Sprintf(`%s '%s'`, action, escape(value)), nil
		}
	}
	like := func(pattern string, v interface{}) (string, error) {
		switch value := v.(type) {
		default:
			return fmt.Sprintf(pattern, escape(value)), nil
		}
	}
	in := func(action string, v interface{}) (ret string, err error) {
		switch value := v.(type) {
		case []interface{}:
			var list []string
			for _, ival := range value {
				list = append(list, escape(ival))
			}
			if len(list) > 0 {
				ret = fmt.Sprintf(`%s ('%s')`, action, strings.Join(list, `', '`))
			}
		}
		return
	}
	logic := func(action string, v interface{}) (ret string, err error) {
		switch value := v.(type) {
		case []interface{}:
			var list []string
			for _, ival := range value {
				switch avalue := ival.(type) {
				case *types.Map:
					where, err := GetWhere(avalue)
					if err != nil {
						return ``, err
					}
					if len(where) > 0 {
						list = append(list, where)
					}
				}
			}
			if len(list) > 0 {
				ret = fmt.Sprintf(`(%s)`, strings.Join(list, ` `+action+` `))
			}
		}
		return
	}
	for _, key := range inWhere.Keys() {
		v, _ := inWhere.Get(key)
		key = PrepareWhere(converter.Sanitize(strings.ToLower(key), `->$`))
		switch key {
		case `$like`:
			return like(`like '%%%s%%'`, v)
		case `$end`:
			return like(`like '%%%s'`, v)
		case `$begin`:
			return like(`like '%s%%'`, v)
		case `$ilike`:
			return like(`ilike '%%%s%%'`, v)
		case `$iend`:
			return like(`ilike '%%%s'`, v)
		case `$ibegin`:
			return like(`ilike '%s%%'`, v)
		case `$and`:
			icond, err := logic(`and`, v)
			if err != nil {
				return ``, err
			}
			cond = append(cond, icond)
		case `$or`:
			icond, err := logic(`or`, v)
			if err != nil {
				return ``, err
			}
			cond = append(cond, icond)
		case `$in`:
			return in(`in`, v)
		case `$nin`:
			return in(`not in`, v)
		case `$eq`:
			return oper(`=`, v)
		case `$neq`:
			return oper(`!=`, v)
		case `$gt`:
			return oper(`>`, v)
		case `$gte`:
			return oper(`>=`, v)
		case `$lt`:
			return oper(`<`, v)
		case `$lte`:
			return oper(`<=`, v)
		default:
			if !strings.Contains(key, `>`) && len(key) > 0 {
				key = `"` + key + `"`
			}
			switch value := v.(type) {
			case []interface{}:
				var acond []string
				for _, iarr := range value {
					switch avalue := iarr.(type) {
					case *types.Map:
						ret, err := GetWhere(avalue)
						if err != nil {
							return ``, err
						}
						acond = append(acond, fmt.Sprintf(`(%s %s)`, key, ret))
					default:
						acond = append(acond, fmt.Sprintf(`%s = '%s'`, key, escape(avalue)))
					}
				}
				if len(acond) > 0 {
					cond = append(cond, fmt.Sprintf(`(%s)`, strings.Join(acond, ` and `)))
				}
			case *types.Map:
				ret, err := GetWhere(value)
				if err != nil {
					return ``, err
				}
				cond = append(cond, fmt.Sprintf(`(%s %s)`, key, ret))
			default:
				ival := escape(value)
				if ival == `$isnull` {
					ival = fmt.Sprintf(`%s is null`, key)
				} else {
					ival = fmt.Sprintf(`%s = '%s'`, key, ival)
				}
				cond = append(cond, ival)
			}
		}
	}
	if len(cond) > 0 {
		where = strings.Join(cond, ` and `)
		if err := CheckNow(where); err != nil {
			return ``, err
		}
	}
	return where, nil
}
