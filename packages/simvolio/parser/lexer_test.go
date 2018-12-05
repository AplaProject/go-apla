package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenScan(t *testing.T) {
	type out struct {
		typ     int
		literal string
	}
	tests := []struct {
		in  string
		out []out
	}{
		{"\n\r\t", []out{}},
		{"var one = 1", []out{
			{VAR, "var"},
			{IDENT, "one"},
			{ASSIGN, "="},
			{INT, "1"}}},
		{`3.14 314 "314"`, []out{
			{FLOAT, "3.14"},
			{INT, "314"},
			{STRING, `"314"`}}},
		{"1+2-3*4/2%2", []out{
			{INT, "1"},
			{ADD, "+"},
			{INT, "2"},
			{SUB, "-"},
			{INT, "3"},
			{MUL, "*"},
			{INT, "4"},
			{DIV, "/"},
			{INT, "2"},
			{MOD, "%"},
			{INT, "2"}}},
		{"1<2>3<=4>=5", []out{
			{INT, "1"},
			{LT, "<"},
			{INT, "2"},
			{GT, ">"},
			{INT, "3"},
			{LTE, "<="},
			{INT, "4"},
			{GTE, ">="},
			{INT, "5"}}},
		{"1==2 || 1 != 2 || 1 && 2", []out{
			{INT, "1"},
			{EQ, "=="},
			{INT, "2"},
			{OR, "||"},
			{INT, "1"},
			{NOT_EQ, "!="},
			{INT, "2"},
			{OR, "||"},
			{INT, "1"},
			{AND, "&&"},
			{INT, "2"}}},
		{"fn(a, ...arr){return}", []out{
			{IDENT, "fn"},
			{LPAREN, "("},
			{IDENT, "a"},
			{COMMA, ","},
			{ELLIPSIS, "..."},
			{IDENT, "arr"},
			{RPAREN, ")"},
			{LBRACE, "{"},
			{RETURN, "return"},
			{RBRACE, "}"}}},
		{"a[:1]", []out{
			{IDENT, "a"},
			{LBRAKET, "["},
			{COLON, ":"},
			{INT, "1"},
			{RBRAKET, "]"}}},
		{"contract My {data{} condition{} action{}}", []out{
			{CONTRACT, "contract"},
			{IDENT, "My"},
			{LBRACE, "{"},
			{DATA, "data"},
			{LBRACE, "{"},
			{RBRACE, "}"},
			{CONDITION, "condition"},
			{LBRACE, "{"},
			{RBRACE, "}"},
			{ACTION, "action"},
			{LBRACE, "{"},
			{RBRACE, "}"},
			{RBRACE, "}"}}},
		{"func() { return }", []out{
			{FUNC, "func"},
			{LPAREN, "("},
			{RPAREN, ")"},
			{LBRACE, "{"},
			{RETURN, "return"},
			{RBRACE, "}"}}},
		{"$a = false if (true) {} else {}", []out{
			{EXTEND_VAR, "$a"},
			{ASSIGN, "="},
			{FALSE, "false"},
			{IF, "if"},
			{LPAREN, "("},
			{TRUE, "true"},
			{RPAREN, ")"},
			{LBRACE, "{"},
			{RBRACE, "}"},
			{ELSE, "else"},
			{LBRACE, "{"},
			{RBRACE, "}"}}},
		{"while (true) { continue break }", []out{
			{WHILE, "while"},
			{LPAREN, "("},
			{TRUE, "true"},
			{RPAREN, ")"},
			{LBRACE, "{"},
			{CONTINUE, "continue"},
			{BREAK, "break"},
			{RBRACE, "}"}}},
		{"info warning error nil", []out{
			{INFO, "info"},
			{WARNING, "warning"},
			{ERROR, "error"},
			{NIL, "nil"}}},
		{"bool int float money string bytes array map file", []out{
			{T_BOOL, "bool"},
			{T_INT, "int"},
			{T_FLOAT, "float"},
			{T_MONEY, "money"},
			{T_STRING, "string"},
			{T_BYTES, "bytes"},
			{T_ARRAY, "array"},
			{T_MAP, "map"},
			{T_FILE, "file"}}},
	}

	for _, v := range tests {
		yyv := new(yySymType)
		l, err := NewLexer("", v.in)
		assert.NoError(t, err)

		list := []out{}
		for {
			c := l.Lex(yyv)
			if c == 0 {
				break
			}
			list = append(list, out{
				c,
				string(l.TokenBytes(nil)),
			})
		}
		if !assert.Equal(t, v.out, list) {
			fmt.Println(v.in)
		}
	}
}

func TestTokenPosition(t *testing.T) {
	in := `var a = 1
var b = 2
var c = a + b`

	type out struct {
		tokenType int
		file      string
	}

	tokens := []out{
		{VAR, "file:1:1"},
		{IDENT, "file:1:5"},
		{ASSIGN, "file:1:7"},
		{INT, "file:1:9"},
		{VAR, "file:2:1"},
		{IDENT, "file:2:5"},
		{ASSIGN, "file:2:7"},
		{INT, "file:2:9"},
		{VAR, "file:3:1"},
		{IDENT, "file:3:5"},
		{ASSIGN, "file:3:7"},
		{IDENT, "file:3:9"},
		{ADD, "file:3:11"},
		{IDENT, "file:3:13"},
	}

	l, err := NewLexer("file", in)
	assert.NoError(t, err)

	for i := 0; ; i++ {
		yyv := new(yySymType)
		c := l.Lex(yyv)
		if c == 0 {
			break
		}

		assert.Equal(t, tokens[i], out{
			c,
			l.FilePosition().String(),
		})
	}
}
