package server

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/kzinglzy/godis/server/protocol"
)

// Client .
type Client struct {
	db     *Database
	conn   net.Conn
	parser *protocol.Parser
	writer *protocol.Writer
	wg     *sync.WaitGroup
	fake   bool
}

func NewClient(conn net.Conn, db *Database) *Client {
	return &Client{
		db:     db,
		conn:   conn,
		parser: protocol.NewParser(conn),
		writer: protocol.NewWriter(conn),
		wg:     new(sync.WaitGroup),
	}
}

func NewFakeClient(reader io.Reader, db *Database) *Client {
	return &Client{
		db:     db,
		conn:   nil,
		parser: protocol.NewParser(reader),
		writer: protocol.NewWriter(nil),
		wg:     new(sync.WaitGroup),
		fake:   true,
	}
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Requests .
func (c *Client) Requests() <-chan *protocol.Request {
	return c.parser.Requests()
}

func (c *Client) SetDatabase(db *Database) {
	c.db = db
}

func (c *Client) ReplyEmpty() error {
	if c.fake {
		return nil
	}
	err := c.writer.WriteBulk(nil)
	if err != nil {
		log.Printf("failed to reply empty %v", err)
	}
	return err
}

func (c *Client) Reply(s string) error {
	if c.fake {
		return nil
	}
	err := c.writer.WriteSimpleString(s)
	if err != nil {
		log.Printf("failed to write string %v", err)
	}
	return err
}

func (c *Client) ReplyBulk(v ...interface{}) error {
	if c.fake {
		return nil
	}
	err := c.writer.WriteObjects(v...)
	if err != nil {
		log.Printf("failed to write objects %v, %v", v, err)
	}
	return err
}

func (c *Client) ReplyError(s string) error {
	if c.fake {
		return nil
	}
	err := c.writer.WriteError(s)
	if err != nil {
		log.Printf("failed to write error %v", err)
	}
	return err
}

func (c *Client) ReplyInt(n int64) error {
	if c.fake {
		return nil
	}
	return c.writer.WriteInt(n)
}

func (c *Client) ReplyList(list []string) error {
	if c.fake {
		return nil
	}
	var bts [][]byte
	for _, e := range list {
		bts = append(bts, []byte(e))
	}
	return c.writer.WriteBulksSlice(bts)
}
