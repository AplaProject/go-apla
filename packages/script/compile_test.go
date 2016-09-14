package script

import (
	"fmt"
	"testing"
)

type TestComp struct {
	Input  string
	Output string
}

func (bytecode Bytecodes) String(source []rune) (ret string) {
	for _, item := range bytecode {
		if item.Cmd == CMD_ERROR {
			item.Value = item.Value.(string) + fmt.Sprintf(` [Ln:%d Col:%d]`, item.Lex.Line, item.Lex.Column)
		}
		ret += fmt.Sprintf("[%d %v]", item.Cmd, item.Value)
	}
	return
}

func TestCompile(t *testing.T) {
	test := []TestComp{
		{"12346 7890", "[1 12346][1 7890]"},
		{"460+ 1540", "[1 460][1 1540][2 1]"},
		{"10 - 2 *3", "[1 10][1 2][1 3][4 2][3 1]"},
		{"20/5 + 78 * 23*1", "[1 20][1 5][5 2][1 78][1 23][4 2][1 1][4 2][2 1]"},
		{"5*(2 + 3)", "[1 5][1 2][1 3][2 1][4 2]"},
		{"(67-23)*45 + (2*7-56)/100", "[1 67][1 23][3 1][1 45][4 2][1 2][1 7][4 2][1 56][3 1][1 100][5 2][2 1]"},
		{"5*(25 / (3+2) - 1)", "[1 5][1 25][1 3][1 2][2 1][5 2][1 1][3 1][4 2]"},
		{"(8 +(3+2*((33-11))))", "[1 8][1 3][1 2][1 33][1 11][3 1][4 2][2 1][2 1]"},
		{"(8 +11))+56", "[1 8][1 11][2 1][0 there is not pair ) [Ln:1 Col:8]]"},
		{"(99+ 76)(1+67", "[1 99][1 76][2 1][1 1][1 67][2 1][0 there is not pair [Ln:1 Col:9]]"},
	}
	for _, item := range test {
		source := []rune(item.Input)
		out := Compile(source).String(source)

		if out != item.Output {
			t.Error(`error of compile ` + item.Input)
		}
		//		fmt.Println(out)
	}
}

func TestEval(t *testing.T) {
	test := []TestComp{
		{"789 63", "63"},
		{"+421", "Stack empty [1:1]"},
		{"1256778+223445", "1480223"},
		{"(67-34789)*3", "-104166"},
		{"(5+78)*(1563-527)", "85988"},
		{"124 * (143-527", "there is not pair [1:7]"},
		{"341 * 234/0", "divided by zero [1:10]"},
		{"((15+82)*2 + 5)/2", "99"},
	}
	for _, item := range test {
		out := Eval(item.Input)
		if fmt.Sprint(out) != item.Output {
			t.Error(`error of eval ` + item.Input)
		}
		fmt.Println(out)
	}
}
