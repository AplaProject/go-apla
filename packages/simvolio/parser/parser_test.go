package parser

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrammar(t *testing.T) {
	syntaxErr := func(v string) error {
		return errors.New(v)
	}

	tests := []struct {
		in  string
		err error
	}{
		{
			"return",
			syntaxErr("file:1:1: syntax error: unexpected RETURN, expecting CONTRACT or FUNC"),
		},
		{
			"func() {}",
			syntaxErr("file:1:5: syntax error: unexpected LPAREN, expecting IDENT"),
		},
		{
			"func f {}",
			syntaxErr("file:1:8: syntax error: unexpected LBRACE, expecting LPAREN"),
		},
		{
			"func f(x) {}",
			syntaxErr("file:1:9: syntax error: unexpected RPAREN"),
		},
		{
			"func f(x, ) {}",
			syntaxErr("file:1:11: syntax error: unexpected RPAREN, expecting IDENT"),
		},
		{
			"func f(x int) {}",
			nil,
		},
		{
			"func f(x int, ) {}",
			syntaxErr("file:1:15: syntax error: unexpected RPAREN, expecting IDENT"),
		},
		{
			"func f(x, y int) {}",
			nil,
		},
		{
			"func f(x bool, y, z int) {}",
			nil,
		},
		{
			"func f(x int, y ...) {}",
			nil,
		},
		{
			"func f() int {}",
			nil,
		},
		{
			`func f(){}
			func f1(){}`,
			nil,
		},
		{
			"func f(){} func f1(){} ",
			nil,
		},
		{
			"func f(){ return }",
			nil,
		},
		{
			`func f(){
				func a(){}
				a()
			}`,
			nil,
		},
		{
			`func f(){
				var a, b int
				return a + b
			}`,
			nil,
		},
		{
			`func f() {
				contract c {}
			}`,
			syntaxErr("file:2:5: syntax error: unexpected CONTRACT"),
		},
		{
			"contract c",
			syntaxErr("file:1:11: syntax error: unexpected $end, expecting LBRACE"),
		},
		{
			"contract c {}",
			nil,
		},
		{
			"contract c { data }",
			syntaxErr("file:1:19: syntax error: unexpected RBRACE, expecting LBRACE"),
		},
		{
			`contract c {
				data {
					p1
				}
			}`,
			syntaxErr("file:4:5: syntax error: unexpected RBRACE"),
		},
		{
			`contract c {
				data {
					p1 int
					p2 string "optional"
				}
			}`,
			nil,
		},
		{
			`contract c {
				data {
					func a() {}
				}
			}`,
			syntaxErr("file:3:6: syntax error: unexpected FUNC, expecting IDENT or RBRACE"),
		},
		{
			`contract c {
				data {}
				condition {}
				action {}
			}`,
			nil,
		},
		{
			`contract c {
				condition {
					var a, b int
					return a + b
				}
			}`,
			nil,
		},
		{
			`contract c {
				action {
					var a, b int
					return a + b
				}
			}`,
			nil,
		},
		{
			`contract c {
				func f(x, y int) int {
					return x+y
				}
			}`,
			nil,
		},
	}

	yyErrorVerbose = true

	for _, v := range tests {
		l, err := NewLexer("file", v.in)
		assert.NoError(t, err)
		yyParse(l)

		if v.err != nil {
			assert.EqualError(t, l.err, v.err.Error())
			continue
		}

		if l.err != nil {
			t.Error(l.err)
			break
		}
	}
}
