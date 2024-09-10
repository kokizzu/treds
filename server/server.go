package server

import (
	"fmt"
	"github.com/tidwall/evio"
	"strings"
	"treds/commands"
	"treds/store"
)

type Server struct {
	Port  int
	ErrCh chan error
	Store store.Store
}

func New(port int) *Server {
	return &Server{
		ErrCh: make(chan error),
		Port:  port,
		Store: store.NewRadixStore(),
	}
}

func (s *Server) Init() {

	commandRegistry := commands.NewRegistry()
	commands.RegisterCommands(commandRegistry)

	var events evio.Events

	// numLoops should always be 1 because datastructures do not support MVCC.
	// This is single threaded operation
	events.NumLoops = 1 // Single-threaded

	// Handle new connections
	events.Serving = func(s evio.Server) (action evio.Action) {
		fmt.Printf("Server started on %s\n", s.Addrs[0])
		return
	}

	// Handle data read from clients
	events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		// Simple command handling: reply with PONG to PING command
		inp := string(in)
		if strings.ToUpper(inp) == "PING\n" {
			out = []byte("PONG\n")
		} else {
			commandString := strings.TrimSpace(inp)
			commandStringParts := strings.Split(commandString, " ")
			command := strings.ToUpper(commandStringParts[0])
			commandReg, err := commandRegistry.Retrieve(command)
			if err != nil {
				out = []byte(fmt.Sprintf("Error Executing command - %v\n", err.Error()))
				return
			}
			if err = commandReg.Validate(commandStringParts[1:]); err != nil {
				out = []byte(fmt.Sprintf("Error Validating command - %v\n", err.Error()))
				return
			}
			res, err := commandReg.Execute(commandStringParts[1:], s.Store)
			if err != nil {
				out = []byte(fmt.Sprintf("Error Executing command - %v\n", err.Error()))
				return
			}
			out = []byte(fmt.Sprintf("%s\n", res))
		}
		return
	}

	// Define the address to listen on
	address := fmt.Sprintf("tcp://0.0.0.0:%d", s.Port)

	// Start the server
	if err := evio.Serve(events, address); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}