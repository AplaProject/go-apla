package resp

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"
)

// Server represents a RESP server which handles reading RESP Values.
type Server struct {
	mu       sync.RWMutex
	handlers map[string]func(conn *Conn, args []Value) bool
	accept   func(conn *Conn) bool
}

// Conn represents a RESP network connection.
type Conn struct {
	*Reader
	*Writer
	base       net.Conn
	RemoteAddr string
}

// NewConn returns a Conn.
func NewConn(conn net.Conn) *Conn {
	return &Conn{
		Reader:     NewReader(conn),
		Writer:     NewWriter(conn),
		base:       conn,
		RemoteAddr: conn.RemoteAddr().String(),
	}
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{
		handlers: make(map[string]func(conn *Conn, args []Value) bool),
	}
}

// HandleFunc registers the handler function for the given command.
// The conn parameter is a Conn type and it can be used to read and write further RESP messages from and to the connection.
// Returning false will close the connection.
func (s *Server) HandleFunc(command string, handler func(conn *Conn, args []Value) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[strings.ToUpper(command)] = handler
}

// AcceptFunc registers a function for accepting connections.
// Calling this function is optional and it allows for total control over reading and writing RESP Values from and to the connections.
// Returning false will close the connection.
func (s *Server) AcceptFunc(accept func(conn *Conn) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accept = accept
}

// ListenAndServe listens on the TCP network address addr for incoming connections.
func (s *Server) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func() {
			err = s.handleConn(conn)
			if err != nil {
				if _, ok := err.(*errProtocol); ok {
					io.WriteString(conn, "-ERR "+formSingleLine(err.Error())+"\r\n")
				} else {
					io.WriteString(conn, "-ERR unknown error\r\n")
				}
			}
			conn.Close()
		}()
	}
}

func (s *Server) handleConn(nconn net.Conn) error {
	conn := NewConn(nconn)
	s.mu.RLock()
	accept := s.accept
	s.mu.RUnlock()
	if accept != nil {
		if !accept(conn) {
			return nil
		}
	}
	for {
		v, _, _, err := conn.ReadMultiBulk()
		if err != nil {
			return err
		}
		values := v.Array()
		if len(values) == 0 {
			continue
		}
		lccommandName := values[0].String()
		commandName := strings.ToUpper(lccommandName)
		s.mu.RLock()
		h := s.handlers[commandName]
		s.mu.RUnlock()
		switch commandName {
		case "QUIT":
			if h == nil {
				conn.WriteSimpleString("OK")
				return nil
			}
		case "PING":
			if h == nil {
				if err := conn.WriteSimpleString("PONG"); err != nil {
					return err
				}
				continue
			}
		}
		if h == nil {
			if err := conn.WriteError(errors.New("ERR unknown command '" + lccommandName + "'")); err != nil {
				return err
			}
		} else {
			if !h(conn, values) {
				return nil
			}
		}
	}
}
