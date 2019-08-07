package server

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/kzinglzy/godis/server/protocol"
)

type Server struct {
	addr     string
	db       *Database
	listener net.Listener

	// client
	events  chan *IOEvent
	clients []*Client

	// aof
	dirty                  int64
	aof                    *os.File
	aofBuf                 []byte
	aofFsyncPolicy         int
	aofFlushPostponedStart int64
	aofFlushInProgress     bool

	// memory policy
	maxmemory       int64
	maxmemoryPolicy uint8
}

var godisServer *Server

type IOEvent struct {
	c *Client
	r *protocol.Request
}

func MakeServer(addr string) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("Failed listening at %s", addr)
	}

	server := &Server{
		addr:            addr,
		db:              NewDatabase(),
		listener:        listener,
		events:          make(chan *IOEvent, 1000),
		clients:         []*Client{},
		aofFsyncPolicy:  AOFFsyncEverysec,
		maxmemory:       MaxMemory,
		maxmemoryPolicy: MaxmemoryAllkeysLRU,
	}
	godisServer = server

	server.openAofFile()
	server.loadDataFromDisk()
	return server, nil
}

func (s *Server) Run() {
	log.Printf("Running godis server at %s", s.addr)
	go s.handleConnection()

	// event loop
	for {
		s.processIOEvent()
		s.processTimeEvent()
		s.afterEvent()
	}
}

func (s *Server) Close() {
	flushAppendOnlyFile(true)
	s.aof.Close()
	s.listener.Close()
}

func (s *Server) afterEvent() {
	s.db.doExpireCycle()
	flushAppendOnlyFile(false)
}

func (s *Server) processIOEvent() {
	n := 0
	scheduled := false

	select {
	case <-time.After(time.Millisecond):
		scheduled = true
	case e := <-s.events:
		n++
		freeMemoryIfNeed()

		dirty := s.dirty
		cmd := LoopupCommand(e.r.CommandName())
		cmd.Exec(e.c, e.r)
		e.c.wg.Done()

		if s.dirty-dirty > 0 {
			feedAppendOnlyFile(e.r)
		}

		if n >= MaxIOEventsPerLoop || scheduled {
			return
		}
	}
}

func (s *Server) processTimeEvent() {
	s.db.doExpireCycle()
	s.db.incrementallyRehash()

	if s.aofFlushPostponedStart != 0 {
		flushAppendOnlyFile(false)
	}
	rewriteAOFBackgroundIfNeed()
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

func (s *Server) handleClient(conn net.Conn) {
	log.Println("create new client")

	client := NewClient(conn, s.db)
	defer client.Close()

	s.clients = append(s.clients, client)

	for req := range client.Requests() {
		e := IOEvent{
			c: client,
			r: req,
		}
		client.wg.Add(1)
		s.events <- &e
	}

	client.wg.Wait()
}

func (s *Server) memPolicyLru() bool {
	return s.maxmemoryPolicy == MaxmemoryAllkeysLRU
}

func (s *Server) memPolicyRandom() bool {
	return s.maxmemoryPolicy == MaxmemoryAllkeysRandom
}

func (s *Server) loadDataFromDisk() {
	log.Printf("loading data from disk")
	fakeClient := NewFakeClient(s.aof, s.db)
	for req := range fakeClient.Requests() {
		cmd := LoopupCommand(req.CommandName())
		cmd.Exec(fakeClient, req)
	}
}

func (s *Server) openAofFile() {
	fd, err := os.OpenFile(AOFFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalf("can't open the append log file: %v", err)
	}
	s.aof = fd
}

func (s *Server) resetAofState() {
	s.dirty = 0
	s.aofFlushPostponedStart = 0
	s.aofBuf = []byte{}
}

func (s *Server) aofFsync() {
	s.aofFlushInProgress = true
	s.aof.Sync()
	s.aofFlushInProgress = false
}
