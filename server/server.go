package server

import (
	"log"
	"net"

	"github.com/kzinglzy/godis/protocol"
)

// Server .
type Server struct {
	addr     string
	hz       int
	db       *Database
	listener net.Listener

	events  chan *IOEvent
	clients []*Client

	dirty       int
	rdbChildPid int
	aofChildPid int
}

var godisServer *Server

// IOEvent .
type IOEvent struct {
	c *Client
	r *protocol.Request
}

// MakeServer .
func MakeServer(addr string) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("Failed listening at %s", addr)
	}

	server := &Server{
		addr:        addr,
		db:          NewDatabase(),
		listener:    listener,
		events:      make(chan *IOEvent, 1000),
		rdbChildPid: -1,
		aofChildPid: -1,
	}
	godisServer = server
	return server, nil
}

// Run .
func (s *Server) Run() {
	log.Printf("Running godis server at %s", s.addr)

	go s.handleConnection()

	// event loop
	for {
		s.processIOEvent()
		s.processTimeEvent()
	}
}

func (s *Server) processTimeEvent() {
	// clientsCron
	// databasesCron
	// rewrite aof
	// flushAppendOnlyFile
}

func (s *Server) processIOEvent() {
	n := 0
	for e := range s.events {
		cmd := LoopupCommand(e.r.CommandName())
		err := cmd.Exec(e.c, e.r)
		if err != nil {
			log.Printf("failed to exec command %s, err: %v", cmd, err)
		}
		n++
		if n >= MaxNumEventsPerLoop {
			return
		}
	}
}

func (s *Server) handleConnection() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Error on accept connection: ", err)
			continue
		}
		go s.handleClient(conn)
	}
}

func (s *Server) isSaving() bool {
	return s.rdbChildPid == -1 && s.aofChildPid == -1
}

func (s *Server) handleClient(conn net.Conn) {
	log.Println("create new client")

	client := NewClient(conn)
	client.SetDatabase(s.db)
	defer client.Close()
	s.clients = append(s.clients, client)

	for req := range client.Requests() {
		e := IOEvent{
			c: client,
			r: req,
		}
		s.events <- &e
	}
}
