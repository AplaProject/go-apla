package script

import (
	"fmt"
	"strconv"
)

const (
	CMD_ERROR = iota // error
	CMD_PUSH         // Push value to stack
	CMD_ADD
	CMD_SUB
	CMD_MUL
	CMD_DIV

	CMD_SYS = 0xff
)

type Oper struct {
	Cmd      uint16
	Priority uint16
}

var (
	OPERS = map[string]Oper{
		`+`: {CMD_ADD, 1}, `-`: {CMD_SUB, 1}, `*`: {CMD_MUL, 2},
		`/`: {CMD_DIV, 2}, `(`: {CMD_SYS, 0xff}, `)`: {CMD_SYS, 0},
	}
)

type Bytecode struct {
	Cmd   uint16
	Value interface{}
	Lex   *Lexem
}

type Bytecodes []*Bytecode

func Compile(input []rune) Bytecodes {
	bytecode := make(Bytecodes, 0, 100)

	lexems := LexParser(input)
	if len(lexems) == 0 {
		return append(bytecode, &Bytecode{CMD_ERROR, `empty program`, nil})
	}
	last := lexems[len(lexems)-1]
	if last.Type == LEX_UNKNOWN {
		return append(bytecode, &Bytecode{CMD_ERROR, fmt.Sprintf(`unknown lexem %s`,
			string(input[last.Offset:last.Right])), last})
	}
	buffer := make(Bytecodes, 0, 20)
	for _, lexem := range lexems {
		var cmd *Bytecode
		strlex := string(input[lexem.Offset:lexem.Right])
		switch lexem.Type {
		case LEX_SYS:
			switch strlex {
			case `(`:
				buffer = append(buffer, &Bytecode{CMD_SYS, uint16(0xff), lexem})
			case `)`:
				for {
					if len(buffer) == 0 {
						cmd = &Bytecode{CMD_ERROR, `there is not pair`, lexem}
						break
					} else {
						prev := buffer[len(buffer)-1]
						buffer = buffer[:len(buffer)-1]
						if prev.Value.(uint16) == 0xff {
							break
						} else {
							bytecode = append(bytecode, prev)
						}
					}

				}
			}
		case LEX_OPER:
			if oper, ok := OPERS[strlex]; ok {
				byteOper := &Bytecode{oper.Cmd, oper.Priority, lexem}
				for {
					if len(buffer) == 0 {
						buffer = append(buffer, byteOper)
						break
					} else {
						prev := buffer[len(buffer)-1]
						if prev.Value.(uint16) >= oper.Priority && prev.Cmd != CMD_SYS {
							bytecode = append(bytecode, prev)
							buffer = buffer[:len(buffer)-1]
						} else {
							buffer = append(buffer, byteOper)
							break
						}
					}
				}
			} else {
				cmd = &Bytecode{CMD_ERROR, `unknown operator`, lexem}
			}
		case LEX_NUMBER:
			if val, err := strconv.ParseInt(strlex, 10, 64); err == nil {
				cmd = &Bytecode{CMD_PUSH, val, lexem}
			} else {
				cmd = &Bytecode{CMD_ERROR, err.Error(), lexem}
			}
		}
		if cmd != nil {
			bytecode = append(bytecode, cmd)
			if cmd.Cmd == CMD_ERROR {
				cmd.Value = fmt.Sprintf(`%s %s`, cmd.Value.(string), strlex)
				cmd.Lex = lexem
				break
			}
		}
	}
	for i := len(buffer) - 1; i >= 0; i-- {
		if buffer[i].Cmd == CMD_SYS {
			bytecode = append(bytecode, &Bytecode{CMD_ERROR, fmt.Sprintf(`there is not pair`), buffer[i].Lex})
			break
		} else {
			bytecode = append(bytecode, buffer[i])
		}
	}
	return bytecode
}
