package server

import (
	"log"
	"net"
)

// Server .
type Server struct {
	addr     string
	hz       int
	db       *Database
	listener net.Listener

	clients []*Client

	dirty       int
	rdbChildPid int
	aofChildPid int
}

var godisServer *Server

// MakeServer .
func MakeServer(addr string) (*Server, error) {
	InitEvictionPoolEntry()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("Failed listening at %s", addr)
	}

	server := &Server{
		addr:        addr,
		db:          NewDatabase(),
		listener:    listener,
		rdbChildPid: -1,
		aofChildPid: -1,
	}
	godisServer = server
	return server, nil
}

// Run .
func (s *Server) Run() {
	log.Printf("Running godis server at %s", s.addr)

	go s.serverCron()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Error on accept connection: ", err)
			continue
		}
		go s.handleClient(conn)
	}

}

func (s *Server) addClients(c *Client) {
	s.clients = append(s.clients, c)
}

func (s *Server) serverCron() {

}

func (s *Server) isSaving() bool {
	return s.rdbChildPid == -1 && s.aofChildPid == -1
}

func (s *Server) handleClient(conn net.Conn) {
	client := NewClient(conn)
	client.SetDatabase(s.db)
	defer client.Close()
	s.addClients(client)

	for req := range client.Requests() {
		cmd := LoopupCommand(req.CommandName())
		err := cmd.Exec(client, req)
		if err != nil {
			log.Printf("failed to exec command %s, err: %v", cmd, err)
		}
	}
}
