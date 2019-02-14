package undo

import (
	"fmt"

	"github.com/AplaProject/go-apla/packages/types"
)

type Storage struct {
	stack []*types.UndoState
}

func (s *Storage) NewStack() types.UndoStack {
	s.stack = s.stack[:0]

	return &Stack{
		stack: make([]*types.UndoState, 0),
	}
}

func (s *Storage) Save() error {
	fmt.Println("SAVE STACK", len(s.stack))
	return nil
}

func NewStorage() *Storage {
	return &Storage{
		stack: make([]*types.UndoState, 0),
	}
}
