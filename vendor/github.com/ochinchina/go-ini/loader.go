package ini

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// remove inline comments
//
// inline comments must start with ';' or '#'
// and the char before the ';' or '#' must be a space
//
func removeComments(value string) string {
	pos := strings.LastIndexAny(value, ";#")

	//if no inline comments
	if pos == -1 || !unicode.IsSpace(rune(value[pos-1])) {
		return value
	}
	return strings.TrimSpace(value[0:pos])
}

// check if it is a oct char,e.g. must be char '0' to '7'
//
func isOctChar(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

// check if the char is a hex char, e.g. the char
// must be '0'..'9' or 'a'..'f' or 'A'..'F'
//
func isHexChar(ch byte) bool {
	return ch >= '0' && ch <= '9' ||
		ch >= 'a' && ch <= 'f' ||
		ch >= 'A' && ch <= 'F'
}

func fromEscape(value string) string {
	if strings.Index(value, "\\") == -1 {
		return value
	}

	r := ""
	n := len(value)
	for i := 0; i < n; i++ {
		if value[i] == '\\' {
			if i+1 < n {
				i++
				//if is it oct
				if i+2 < n && isOctChar(value[i]) && isOctChar(value[i+1]) && isOctChar(value[i+2]) {
					t, err := strconv.ParseInt(value[i:i+3], 8, 32)
					if err == nil {
						r = r + string(rune(t))
					}
					i += 2
					continue
				}
				switch value[i] {
				case '0':
					r = r + string(byte(0))
				case 'a':
					r = r + "\a"
				case 'b':
					r = r + "\b"
				case 'f':
					r = r + "\f"
				case 't':
					r = r + "\t"
				case 'r':
					r = r + "\r"
				case 'n':
					r = r + "\n"
				case 'v':
					r = r + "\v"
				case 'x':
					i++
					if i+3 < n && isHexChar(value[i]) &&
						isHexChar(value[i+1]) &&
						isHexChar(value[i+2]) &&
						isHexChar(value[i+3]) {

						t, err := strconv.ParseInt(value[i:i+4], 16, 32)
						if err == nil {
							r = r + string(rune(t))
						}
						i += 3
					}
				default:
					r = fmt.Sprintf("%s%c", r, value[i])
				}
			}
		} else {
			r = fmt.Sprintf("%s%c", r, value[i])
		}
	}
	return r
}

func toEscape(s string) string {
	result := bytes.NewBuffer(make([]byte, 0))

	n := len(s)

	for i := 0; i < n; i++ {
		switch s[i] {
		case 0:
			result.WriteString("\\0")
		case '\\':
			result.WriteString("\\\\")
		case '\a':
			result.WriteString("\\a")
		case '\b':
			result.WriteString("\\b")
		case '\t':
			result.WriteString("\\t")
		case '\r':
			result.WriteString("\\r")
		case '\n':
			result.WriteString("\\n")
		case ';':
			result.WriteString("\\;")
		case '#':
			result.WriteString("\\#")
		case '=':
			result.WriteString("\\=")
		case ':':
			result.WriteString("\\:")
		default:
			result.WriteByte(s[i])
		}
	}
	return result.String()
}
func removeContinuationSuffix(value string) (string, bool) {
	pos := strings.LastIndex(value, "\\")
	n := len(value)
	if pos == -1 || pos != n-1 {
		return "", false
	}
	for pos >= 0 {
		if value[pos] != '\\' {
			return "", false
		}
		pos--
		if pos < 0 || value[pos] != '\\' {
			return value[0 : n-1], true
		}
		pos--
	}
	return "", false
}

type lineReader struct {
	reader *bufio.Scanner
}

func newLineReader(reader io.Reader) *lineReader {
	return &lineReader{reader: bufio.NewScanner(reader)}
}

func (lr *lineReader) readLine() (string, error) {
	if lr.reader.Scan() {
		return lr.reader.Text(), nil
	}
	return "", errors.New("No data")

}

func readLinesUntilSuffix(lineReader *lineReader, suffix string) string {
	r := ""
	for {
		line, err := lineReader.readLine()
		if err != nil {
			break
		}
		t := strings.TrimRightFunc(line, unicode.IsSpace)
		if strings.HasSuffix(t, suffix) {
			r = r + t[0:len(t)-len(suffix)]
			break
		} else {
			r = r + line + "\n"
		}
	}
	return r
}

func readContinuationLines(lineReader *lineReader) string {
	r := ""
	for {
		line, err := lineReader.readLine()
		if err != nil {
			break
		}
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		if t, continuation := removeContinuationSuffix(line); continuation {
			r = r + t
		} else {
			r = r + line
			break
		}
	}
	return r
}

/*
Load from the sources, the source can be one of:
    - fileName
    - a string includes .ini
    - io.Reader the reader to load the .ini contents
    - byte array incldues .ini content
*/
func (ini *Ini) Load(sources ...interface{}) {
	for _, source := range sources {
		switch source.(type) {
		case string:
			s, _ := source.(string)
			if _, err := os.Stat(s); err == nil {
				ini.LoadFile(s)
			} else {
				ini.LoadString(s)
			}
		case io.Reader:
			reader, _ := source.(io.Reader)
			ini.LoadReader(reader)
		case []byte:
			b, _ := source.([]byte)
			ini.LoadBytes(b)
		}
	}

}

// Explicitly loads .ini from a reader
//
func (ini *Ini) LoadReader(reader io.Reader) {
	lineReader := newLineReader(reader)
	var curSection *Section = nil
	for {
		line, err := lineReader.readLine()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		//empty line or comments line
		if len(line) <= 0 || line[0] == ';' || line[0] == '#' {
			continue
		}
		//if it is a section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			if len(sectionName) > 0 {
				curSection = ini.NewSection(sectionName)
			}
			continue
		}
		pos := strings.IndexAny(line, "=;")
		if pos != -1 {
			key := strings.TrimSpace(line[0:pos])
			value := strings.TrimLeftFunc(line[pos+1:], unicode.IsSpace)
			//if it is a multiline indicator
			if strings.HasPrefix(value, "\"\"\"") {
				t := strings.TrimRightFunc(value, unicode.IsSpace)
				//if the end multiline indicator is found
				if strings.HasSuffix(t, "\"\"\"") {
					value = t[3 : len(t)-3]
				} else { //read lines until end multiline indicator is found
					value = value[3:] + "\n" + readLinesUntilSuffix(lineReader, "\"\"\"")
				}
			} else {
				value = strings.TrimRightFunc(value, unicode.IsSpace)
				//if is it a continuation line
				if t, continuation := removeContinuationSuffix(value); continuation {
					value = t + readContinuationLines(lineReader)
				}
			}

			if len(key) > 0 {
				if curSection == nil && len(ini.defaultSectionName) > 0 {
					curSection = ini.NewSection(ini.defaultSectionName)
				}
				if curSection != nil {
					//remove the comments and convert escape char to real
					curSection.Add(key, strings.TrimSpace(fromEscape(removeComments(value))))
				}
			}
		}
	}
}

// Load ini file from file named fileName
//
func (ini *Ini) LoadFile(fileName string) {
	f, err := os.Open(fileName)
	if err == nil {
		defer f.Close()
		ini.Load(f)
	}
}

var defaultSectionName string = "default"

func SetDefaultSectionName(defSectionName string) {
	defaultSectionName = defSectionName
}

// load ini from the content which contains the .ini formated string
//
func (ini *Ini) LoadString(content string) {
	ini.Load(bytes.NewBufferString(content))
}

// load .ini from a byte array which contains the .ini formated content
func (ini *Ini) LoadBytes(content []byte) {
	ini.Load(bytes.NewBuffer(content))
}

/*
Load the .ini from one of following resource:
    - file
    - string in .ini format
    - byte array in .ini format
    - io.Reader a reader to load .ini content

One or more source can be provided in this Load method, such as:
    var reader1 io.Reader = ...
    var reader2 io.Reader = ...
    ini.Load( "./my.ini", "[section]\nkey=1", "./my2.ini", reader1, reader2 )
*/
func Load(sources ...interface{}) *Ini {
	ini := NewIni()
	ini.SetDefaultSectionName(defaultSectionName)
	for _, source := range sources {
		ini.Load(source)
	}
	return ini
}
