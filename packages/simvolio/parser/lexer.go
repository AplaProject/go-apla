package parser

import (
	"bytes"
	"fmt"
	"go/token"
	"unicode"

	"github.com/cznic/golex/lex"
)

//go:generate goyacc -o parser.go parser.y
//go:generate golex -o lexer_scan.go lex.l

const (
	classUnicodeLeter = iota + 0x80
	classUnicodeDigit
	classOther
)

func runeClass(r rune) int {
	if r >= 0 && r < 0x80 { // keep ASCII as it is.
		return int(r)
	}

	if unicode.IsLetter(r) {
		return classUnicodeLeter
	}

	if unicode.IsDigit(r) {
		return classUnicodeDigit
	}

	return classOther
}

type lexer struct {
	*lex.Lexer

	result interface{}
	err    error
}

func (l *lexer) char(r int) lex.Char {
	return lex.NewChar(l.First.Pos(), rune(r))
}

func (l *lexer) Lex(lval *yySymType) int {
	c := l.scan(lval)

	if c.Rune == lex.RuneEOF {
		return 0
	}

	return int(c.Rune)
}

func (l *lexer) Error(err string) {
	l.err = fmt.Errorf("%s: %s", l.FilePosition(), err)
}

func (l *lexer) FilePosition() token.Position {
	return l.File.Position(l.First.Pos())
}

func NewLexer(filename string, src string) (*lexer, error) {
	fs := token.NewFileSet()
	file := fs.AddFile(filename, -1, len(src))
	buf := bytes.NewBufferString(src)

	l, err := lex.New(file, buf, lex.RuneClass(runeClass))
	if err != nil {
		return nil, err
	}

	return &lexer{l, nil, nil}, nil
}
