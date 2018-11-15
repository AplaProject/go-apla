package lexer

import (
	"bytes"
	gotoken "go/token"
	"unicode"

	"github.com/GenesisKernel/go-genesis/packages/simvolio/token"

	"github.com/cznic/golex/lex"
)

//go:generate golex -o lexer_scan.go lexer.l

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
}

func (l *lexer) char(r token.TokenType) lex.Char {
	return lex.NewChar(l.First.Pos(), rune(r))
}

func (l *lexer) Scan() lex.Char {
	return l.scan()
}

func NewLexer(filename string, src string) (*lexer, error) {
	fs := gotoken.NewFileSet()
	file := fs.AddFile(filename, -1, len(src))
	buf := bytes.NewBufferString(src)

	lx, err := lex.New(file, buf, lex.RuneClass(runeClass))
	if err != nil {
		return nil, err
	}

	return &lexer{lx}, nil
}
