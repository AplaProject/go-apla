package script

//	"fmt"

const (
	LEX_UNKNOWN = iota
	LEX_SYS
	LEX_OPER
	LEX_NUMBER
	LEX_IDENT

	LEX_ERROR = 0xFF
	LEXF_NEXT = 1
	LEXF_PUSH = 2
	LEXF_POP  = 4
)

type Lexem struct {
	Type   uint8  // Type of the lexem
	Offset uint32 // Absolute offset
	Right  uint32 // Right Offset of the lexem
	Line   uint32 // Line of the lexem
	Column uint32 // Position inside the line
}

type Lexems []*Lexem

func LexParser(input []rune) Lexems {
	var (
		curState, lexId                          uint8
		length, line, off, offline, flags, start uint32
	)

	lexems := make(Lexems, 0, len(input)/4)
	irune := len(ALPHABET) - 1

	todo := func(r rune) {
		var letter uint8
		if r > 127 {
			letter = ALPHABET[irune]
		} else {
			letter = ALPHABET[r]
		}
		val := LEXTABLE[curState][letter]
		curState = uint8(val >> 16)
		lexId = uint8((val >> 8) & 0xff)
		flags = val & 0xff
	}
	length = uint32(len(input)) + 1
	line = 1

	for off < length {
		if off == length-1 {
			todo(rune(' '))
		} else {
			todo(input[off])
		}
		if curState == LEX_ERROR {
			lexems = append(lexems, &Lexem{LEX_UNKNOWN, off, off + 1, line, off - offline + 1})
			break
		}
		if lexId > 0 {
			lexOff := off
			if (flags & LEXF_POP) != 0 {
				lexOff = start
			}
			right := off
			if (flags & LEXF_NEXT) != 0 {
				right++
			}
			lexems = append(lexems, &Lexem{lexId, lexOff, right, line, lexOff - offline + 1})
			if lexId == LEX_SYS && input[lexOff] == rune(0x0a) {
				line++
				offline = off
			}
		}
		if (flags & LEXF_PUSH) != 0 {
			start = off
		}
		if (flags & LEXF_NEXT) != 0 {
			off++
		}
	}
	return lexems
}
