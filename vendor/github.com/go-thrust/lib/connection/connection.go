package connection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-thrust/lib/commands"
	. "github.com/go-thrust/lib/common"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const (
	SOCKET_BOUNDARY = "--(Foo)++__THRUST_SHELL_BOUNDARY__++(Bar)--"
)

// Single Connection
//var conn net.Conn
var Stdin io.WriteCloser
var Stdout io.ReadCloser
var ExecCommand *exec.Cmd

type In struct {
	Commands         chan *commands.Command
	CommandResponses chan *commands.CommandResponse
	Quit             chan int
}
type Out struct {
	CommandResponses chan commands.CommandResponse
	Errors           chan error
}

var in In
var out Out

/*
Initializes threads with Channel Structs
Opens Connection
*/
func InitializeThreads() {
	//c, err := net.Dial(proto, address)
	//conn = c

	in = In{
		Commands:         make(chan *commands.Command),
		CommandResponses: make(chan *commands.CommandResponse),
		Quit:             make(chan int),
	}

	out = Out{
		CommandResponses: make(chan commands.CommandResponse),
		Errors:           make(chan error),
	}

	go Reader(&out, &in)
	go Writer(&out, &in)

	go func() {
		fmt.Println("Registering signals")
		p := []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM, os.Kill}
		c := make(chan os.Signal, len(p))
		signal.Notify(c, p...)

		for s := range c {
			fmt.Println("Getting signal ", s)
			if s == os.Interrupt || s == os.Kill || s == syscall.SIGTERM {
				fmt.Println("Finishing clean up quiting")
				CleanExit()
				return
			}
		}
	}()
	return
}

func GetOutputChannels() *Out {
	return &out
}

func GetInputChannels() *In {
	return &in
}

func GetCommunicationChannels() (*Out, *In) {
	return GetOutputChannels(), GetInputChannels()
}

func Clean() {
	if ExecCommand == nil {
		return
	}
	Log.Print("Killing Thrust Core")
	if err := ExecCommand.Process.Kill(); err != nil {
		Log.Print("failed to kill: ", err)
	}
	Log.Print("Killing Thrust Core ok")
}

func CleanExit() {
	Clean()
	if utils.DB != nil && utils.DB.DB != nil {
		utils.DB.ExecSql(`INSERT INTO stop_daemons(stop_time) VALUES (?)`, utils.Time())
	} else {
		os.Exit(0)
	}
}

func Reader(out *Out, in *In) {

	reader := bufio.NewReader(Stdout)
	defer Stdin.Close()
	for {
		line, err := reader.ReadString(byte('\n'))
		if err != nil {
			fmt.Println(err)
			// For now lets just force cleanup and exit
			CleanExit()
			return
		}

		//Log.Print("SOCKET::Line", line)
		if !strings.Contains(line, SOCKET_BOUNDARY) {
			response := commands.CommandResponse{}
			json.Unmarshal([]byte(line), &response)
			out.CommandResponses <- response
		}
	}
}

func Writer(out *Out, in *In) {
	for {
		select {
		case response := <-in.CommandResponses:
			cmd, _ := json.Marshal(response)
			//Log.Print("Writing RESPONSE", string(cmd), "\n", SOCKET_BOUNDARY)

			Stdin.Write(cmd)
			Stdin.Write([]byte("\n"))
			Stdin.Write([]byte(SOCKET_BOUNDARY))
			Stdin.Write([]byte("\n"))
		case command := <-in.Commands:
			ActionId += 1
			command.ID = ActionId

			//fmt.Println(command)
			cmd, _ := json.Marshal(command)
			//Log.Print("Writing", string(cmd), "\n", SOCKET_BOUNDARY)

			Stdin.Write(cmd)
			Stdin.Write([]byte("\n"))
			Stdin.Write([]byte(SOCKET_BOUNDARY))
			Stdin.Write([]byte("\n"))
		}
	}
}
