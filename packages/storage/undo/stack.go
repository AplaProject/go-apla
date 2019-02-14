package undo

import (
	"github.com/AplaProject/go-apla/packages/types"
)

type Stack struct {
	stack []*types.UndoState
}

func (s *Stack) PushState(st *types.UndoState) {
	s.stack = append(s.stack, st)
}

func (s *Stack) Reset() {
	s.stack = s.stack[:0]
}

func (s *Stack) Stack() []*types.UndoState {
	return s.stack
}
