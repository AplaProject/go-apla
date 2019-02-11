package undo

import (
	"fmt"
)

type Storage struct {
	stack []*State
}

func NewStorage() *Storage {
	return &Storage{
		stack: make([]*State, 0),
	}
}

func (s *Storage) NewStack(db string, block string) *Stack {
	s.stack = s.stack[:0]

	return &Stack{
		db:    db,
		block: block,
		stack: make([]*State, 0),
		s:     s,
	}
}

func (s *Storage) Save() error {
	fmt.Println("SAVE STACK", len(s.stack))
	return nil
}

type Stack struct {
	s     *Storage
	stack []*State

	db, block, tx string
}

type State struct {
	DB    string `json:"db"`
	Block string `json:"block"`
	Tx    string `json:"tx"`
	Table string `json:"table,omitempty"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

func (s *Stack) PushState(st *State) {
	st.DB = s.db
	st.Block = s.block
	st.Tx = s.tx
	s.stack = append(s.stack, st)
	fmt.Println(s.stack)
}

func (s *Stack) Current() []*State {
	return s.stack
}

func (s *Stack) Reset(tx string) {
	s.tx = tx
	s.stack = s.stack[:0]
}

func (s *Stack) Release() {
	fmt.Println(s.stack)
	s.s.stack = append(s.s.stack, s.stack...)
	s.stack = s.stack[:0]
}
