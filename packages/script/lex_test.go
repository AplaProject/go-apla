package script

import (
	"fmt"
	"testing"
)

type TestLexem struct {
	Input  string
	Output string
}

func (lexems Lexems) String(source []rune) (ret string) {
	for _, item := range lexems {
		slex := string(source[item.Offset:item.Right])
		if item.Type == 0 {
			slex = `error`
		}
		ret += fmt.Sprintf("[%d %s]", item.Type, slex)
	}
	return
}

func TestLexParser(t *testing.T) {
	test := []TestLexem{
		{"(\r\n)\x03 -", "[1 (][1 \n][1 )][0 error]"},
		{` +( - )	/ `, `[2 +][1 (][2 -][1 )][2 /]`},
		{`23+13424 Тест`, `[3 23][2 +][3 13424][4 Тест]`},
		{` 0785/67+iname*(56-31)`, `[3 0785][2 /][3 67][2 +][4 iname][2 *][1 (][3 56][2 -][3 31][1 )]`},
		{`myvar_45 - a_qwe + t81you - 345rt`, `[4 myvar_45][2 -][4 a_qwe][2 +][4 t81you][2 -][0 error]`},
	}
	for _, item := range test {
		source := []rune(item.Input)
		out := LexParser(source)
		if out.String(source) != item.Output {
			t.Error(`error of lexical parser ` + item.Input)
		}
		//fmt.Println(out.String(source))
	}
}
