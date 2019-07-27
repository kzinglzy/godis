package server

import (
	"log"
	"net"

	"github.com/kzinglzy/godis/dt"
)

// Server .
type Server struct {
	addr     string
	hz       int
	db       *Database
	listener net.Listener

	clients []*Client
}

// Database implements the kv service
type Database struct {
	store   *dt.Dict
	expires *dt.Dict
}

// MakeServer .
func MakeServer(addr string) (*Server, error) {
	db := &Database{
		store:   dt.NewDict(),
		expires: dt.NewDict(),
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("Failed listening at %s", addr)
	}
	InitEvictionPoolEntry()
	server := Server{
		addr:     addr,
		db:       db,
		listener: listener,
	}
	return &server, nil
}

func (s *Server) Run() {
	go s.serverCron()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Error on accept connection: ", err)
			continue
		}
		go s.handleRequest(&conn)
	}

}

func (s *Server) serverCron() {

}

func (s *Server) handleRequest(net *net.Conn) {

}
