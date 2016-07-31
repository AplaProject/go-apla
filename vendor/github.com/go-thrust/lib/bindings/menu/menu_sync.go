package menu

import "github.com/go-thrust/lib/commands"

type ChildCommand struct {
	Command *commands.Command
	Child   *Menu
}

type MenuSync struct {
	/* Channels for singaling queues */
	ReadyChan       chan bool
	DisplayedChan   chan bool
	ChildStableChan chan uint
	TreeStableChan  chan bool

	/*Queues for preserving command order and Priority*/
	ReadyQueue     []*commands.Command
	DisplayedQueue []*commands.Command
	// Not Exactly a Queue, more of a priority queue. Send out the first child that is stable.
	ChildStableQueue []*ChildCommand
	TreeStableQueue  []*commands.Command

	/* Channels for control */
	QuitChan chan bool
}
