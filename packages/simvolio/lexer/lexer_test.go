package lexer

import (
	"fmt"
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/simvolio/token"

	"github.com/stretchr/testify/assert"
)

func TestTokenScan(t *testing.T) {
	type out struct {
		typ     token.TokenType
		literal string
	}
	tests := []struct {
		in  string
		out []out
	}{
		{"\n\r\t", []out{{
			token.NewLine, "\n"}}},
		{"var one = 1", []out{
			{token.Var, "var"},
			{token.Ident, "one"},
			{token.Assign, "="},
			{token.Int, "1"}}},
		{`3.14 314 "314"`, []out{
			{token.Float, "3.14"},
			{token.Int, "314"},
			{token.String, `"314"`}}},
		{"1+2-3*4/2%2", []out{
			{token.Int, "1"},
			{token.Plus, "+"},
			{token.Int, "2"},
			{token.Minus, "-"},
			{token.Int, "3"},
			{token.Asterisk, "*"},
			{token.Int, "4"},
			{token.Slash, "/"},
			{token.Int, "2"},
			{token.Percent, "%"},
			{token.Int, "2"}}},
		{"1<2>3<=4>=5", []out{
			{token.Int, "1"},
			{token.Lt, "<"},
			{token.Int, "2"},
			{token.Gt, ">"},
			{token.Int, "3"},
			{token.LtEq, "<="},
			{token.Int, "4"},
			{token.GtEq, ">="},
			{token.Int, "5"}}},
		{"1==2 || 1 != 2 || 1 && 2", []out{
			{token.Int, "1"},
			{token.Eq, "=="},
			{token.Int, "2"},
			{token.Or, "||"},
			{token.Int, "1"},
			{token.NotEq, "!="},
			{token.Int, "2"},
			{token.Or, "||"},
			{token.Int, "1"},
			{token.And, "&&"},
			{token.Int, "2"}}},
		{"fn(a, ...arr){return}", []out{
			{token.Ident, "fn"},
			{token.LParen, "("},
			{token.Ident, "a"},
			{token.Comma, ","},
			{token.Tail, "..."},
			{token.Ident, "arr"},
			{token.RParen, ")"},
			{token.LBrace, "{"},
			{token.Return, "return"},
			{token.RBrace, "}"}}},
		{"a[:1]", []out{
			{token.Ident, "a"},
			{token.LBraket, "["},
			{token.Colon, ":"},
			{token.Int, "1"},
			{token.RBraket, "]"}}},
		{"contract My {data{} condition{} action{}}", []out{
			{token.Contract, "contract"},
			{token.Ident, "My"},
			{token.LBrace, "{"},
			{token.Data, "data"},
			{token.LBrace, "{"},
			{token.RBrace, "}"},
			{token.Condition, "condition"},
			{token.LBrace, "{"},
			{token.RBrace, "}"},
			{token.Action, "action"},
			{token.LBrace, "{"},
			{token.RBrace, "}"},
			{token.RBrace, "}"}}},
		{"func() { return }", []out{
			{token.Func, "func"},
			{token.LParen, "("},
			{token.RParen, ")"},
			{token.LBrace, "{"},
			{token.Return, "return"},
			{token.RBrace, "}"}}},
		{"$a = false if (true) {} else {}", []out{
			{token.ExtendVar, "$a"},
			{token.Assign, "="},
			{token.False, "false"},
			{token.If, "if"},
			{token.LParen, "("},
			{token.True, "true"},
			{token.RParen, ")"},
			{token.LBrace, "{"},
			{token.RBrace, "}"},
			{token.Else, "else"},
			{token.LBrace, "{"},
			{token.RBrace, "}"}}},
		{"while (true) { continue break }", []out{
			{token.While, "while"},
			{token.LParen, "("},
			{token.True, "true"},
			{token.RParen, ")"},
			{token.LBrace, "{"},
			{token.Continue, "continue"},
			{token.Break, "break"},
			{token.RBrace, "}"}}},
		{"info warning error nil", []out{
			{token.Info, "info"},
			{token.Warning, "warning"},
			{token.Error, "error"},
			{token.Nil, "nil"}}},
		{"bool int float money string bytes array map file", []out{
			{token.TypeBool, "bool"},
			{token.TypeInt, "int"},
			{token.TypeFloat, "float"},
			{token.TypeMoney, "money"},
			{token.TypeString, "string"},
			{token.TypeBytes, "bytes"},
			{token.TypeArray, "array"},
			{token.TypeMap, "map"},
			{token.TypeFile, "file"}}},
	}

	for _, v := range tests {
		l, err := NewLexer("", v.in)
		assert.NoError(t, err)

		list := []out{}
		for {
			c := l.Scan()
			tok := token.TokenType(c.Rune)
			if tok == token.EOF {
				break
			}
			list = append(list, out{
				tok,
				string(l.TokenBytes(nil)),
			})
		}
		if !assert.Equal(t, v.out, list) {
			for _, v := range list {
				fmt.Printf("{token.%s, %q},\n", v.typ, v.literal)
			}
		}
	}
}

func TestTokenPosition(t *testing.T) {
	in := `var a = 1
var b = 2
var c = a + b`

	type out struct {
		tokenType token.TokenType
		file      string
	}

	tokens := []out{
		{token.Var, "file:1:1"},
		{token.Ident, "file:1:5"},
		{token.Assign, "file:1:7"},
		{token.Int, "file:1:9"},
		{token.NewLine, "file:1:10"},
		{token.Var, "file:2:1"},
		{token.Ident, "file:2:5"},
		{token.Assign, "file:2:7"},
		{token.Int, "file:2:9"},
		{token.NewLine, "file:2:10"},
		{token.Var, "file:3:1"},
		{token.Ident, "file:3:5"},
		{token.Assign, "file:3:7"},
		{token.Ident, "file:3:9"},
		{token.Plus, "file:3:11"},
		{token.Ident, "file:3:13"},
	}

	l, err := NewLexer("file", in)
	assert.NoError(t, err)

	for i := 0; ; i++ {
		c := l.Scan()
		tok := token.TokenType(c.Rune)
		if tok == token.EOF {
			break
		}

		assert.Equal(t, tokens[i], out{
			tok,
			fmt.Sprintf("%s", l.File.Position(c.Pos())),
		})
	}
}
