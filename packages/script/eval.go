package script

import (
	"fmt"
)

type ValStack struct {
	Value interface{}
}

type Stack []*ValStack

func Eval(input string) interface{} {
	bytecode := Compile([]rune(input))
	stack := make(Stack, 0, 1024)
	last := bytecode[len(bytecode)-1]
	if last.Cmd == CMD_ERROR {
		return fmt.Errorf(`%v [%d:%d]`, last.Value, last.Lex.Line, last.Lex.Column)
	}
	for _, cmd := range bytecode {
		size := len(stack)
		if cmd.Cmd > CMD_PUSH && size < 2 {
			return fmt.Errorf(`Stack empty [%d:%d]`, last.Lex.Line, last.Lex.Column)
		}
		switch cmd.Cmd {
		case CMD_PUSH:
			stack = append(stack, &ValStack{Value: cmd.Value})
		case CMD_ADD:
			stack[size-2] = &ValStack{Value: stack[size-2].Value.(int64) + stack[size-1].Value.(int64)}
		case CMD_SUB:
			stack[size-2] = &ValStack{Value: stack[size-2].Value.(int64) - stack[size-1].Value.(int64)}
		case CMD_MUL:
			stack[size-2] = &ValStack{Value: stack[size-2].Value.(int64) * stack[size-1].Value.(int64)}
		case CMD_DIV:
			if stack[size-1].Value.(int64) == 0 {
				return fmt.Errorf(`divided by zero [%d:%d]`, last.Lex.Line, last.Lex.Column)
			}
			stack[size-2] = &ValStack{Value: stack[size-2].Value.(int64) / stack[size-1].Value.(int64)}
		default:
			return fmt.Errorf(`Unknown command [%d:%d]`, last.Lex.Line, last.Lex.Column)
		}
		if cmd.Cmd > CMD_PUSH {
			stack = stack[:size-1]
		}
	}
	if len(stack) == 0 {
		return fmt.Errorf(`Stack empty`)
	}
	return stack[len(stack)-1].Value
}
