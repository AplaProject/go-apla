package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// The program creates packages/script/lex_table.go files.

type Action map[string][]string
type States map[string]Action

var (
	table    [][14]uint32
	lexem    = map[string]uint32{``: 0, `sys`: 1, `oper`: 2, `number`: 3, `ident`: 4}
	flags    = map[string]uint32{`next`: 1, `push`: 2, `pop`: 4}
	alphabet = []byte{0x01, 0x0a, ' ', '(', ')', '*', '+', '-', '/', '0', '1', 'a', '_', 128}
	//              default  n     s                                                      r
	states = `{
	"main": {
			"n()": ["main", "sys", "next"],
			"s": ["main", "", "next"],
			"*+-/": ["main", "oper", "next"],
			"01": ["number", "", "push next"],
			"a_r": ["ident", "", "push next"],
			"d": ["error", "", ""]
		},
	"number": {
			"01": ["number", "", "next"],
			"a_r": ["error", "", ""],
			"d": ["main", "number", "pop"]
		},
	"ident": {
			"01a_r": ["ident", "", "next"],
			"d": ["main", "ident", "pop"]
		}
}`
)

func main() {
	var alpha [129]byte
	for ind, ch := range alphabet {
		i := byte(ind)
		switch ch {
		case ' ':
			alpha[0x09] = i
			alpha[0x0d] = i
			alpha[' '] = i
		case '1':
			for k := '1'; k <= '9'; k++ {
				alpha[k] = i
			}
		case 'a':
			for k := 'A'; k <= 'Z'; k++ {
				alpha[k] = i
			}
			for k := 'a'; k <= 'z'; k++ {
				alpha[k] = i
			}
		case 128:
			alpha[128] = i
		default:
			alpha[ch] = i
		}
	}
	out := `package script
	// This file was generated with /tools/lextable.go
	
var (
		ALPHABET = []byte{`
	for i, ch := range alpha {
		out += fmt.Sprintf(`%d,`, ch)
		if i > 0 && i%24 == 0 {
			out += "\r\n\t\t\t"
		}
	}
	out += "\r\n\t\t}\r\n"

	var (
		data States
	)
	state2int := map[string]uint{`main`: 0}
	if err := json.Unmarshal([]byte(states), &data); err == nil {
		for key := range data {
			if key != `main` {
				state2int[key] = uint(len(state2int))
			}
		}
		table = make([][14]uint32, len(state2int))
		for key, istate := range data {
			curstate := state2int[key]
			for i := range table[curstate] {
				table[curstate][i] = 0xFE0000
			}

			for skey, sval := range istate {
				var val uint32
				if sval[0] == `error` {
					val = 0xff0000
				} else {
					val = uint32(state2int[sval[0]] << 16) // new state
				}
				val |= uint32(lexem[sval[1]] << 8) // lexem
				cmds := strings.Split(sval[2], ` `)
				var flag uint32
				for _, icmd := range cmds {
					flag |= flags[icmd]
				}
				val |= flag
				for _, ch := range []byte(skey) {
					var ind int
					switch ch {
					case 'd':
						ind = 0
					case 'n':
						ind = 1
					case 's':
						ind = 2
					case 'r':
						ind = 13
					default:
						for k, ach := range alphabet {
							if ach == ch {
								ind = k
								break
							}
						}
					}
					table[curstate][ind] = val
					if ind == 0 { // default value
						for i := range table[curstate] {
							if table[curstate][i] == 0xFE0000 {
								table[curstate][i] = val
							}
						}
					}
				}
			}
		}
		out += "\t\tLEXTABLE = [][14]uint32{\r\n"
		for _, line := range table {
			out += "\t\t\t{"
			for _, ival := range line {
				out += fmt.Sprintf(" 0x%x,", ival)
			}
			out += "\r\n\t\t\t},\r\n"
		}
		out += "\t\t\t}\r\n)\r\n"
		err = ioutil.WriteFile("../../packages/script/lex_table.go", []byte(out), 0644)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(err.Error())
	}
}
