package server

import (
	"log"
	"net"

	"github.com/kzinglzy/godis/protocol"
)

// Client .
type Client struct {
	db     *Database
	conn   net.Conn
	parser *protocol.Parser
	writer *protocol.Writer
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:   conn,
		parser: protocol.NewParser(conn),
		writer: protocol.NewWriter(conn),
	}
}

func (c *Client) Close() error {
	c.conn.Close()
	return nil
}

// Requests .
func (c *Client) Requests() <-chan *protocol.Request {
	return c.parser.Requests()
}

func (c *Client) SetDatabase(db *Database) {
	c.db = db
}

func (c *Client) ReplyEmpty() error {
	err := c.writer.WriteBulk(nil)
	if err != nil {
		log.Printf("failed to reply empty %v", err)
	}
	return err
}

func (c *Client) ReplyBulk(s []string) error {
	err := c.writer.WriteBulkStrings(s)
	if err != nil {
		log.Printf("failed to write bulk strings %v", err)
	}
	return err
}

func (c *Client) Reply(v ...interface{}) error {
	err := c.writer.WriteObjects(v...)
	if err != nil {
		log.Printf("failed to write object %v, %v", v, err)
	}
	return err
}

func (c *Client) ReplyError(s string) error {
	err := c.writer.WriteError(s)
	if err != nil {
		log.Printf("failed to write error %v", err)
	}
	return err
}
