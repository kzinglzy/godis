package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/kzinglzy/godis/dt"
	"github.com/kzinglzy/godis/protocol"
)

// CommandTable .
var CommandTable = map[string]Command{
	"get": new(cmdGet),
	"set": new(cmdSet),
}

// Command .
type Command interface {
	Exec(*Client, *protocol.Request) error
}

// LoopupCommand .
func LoopupCommand(name string) Command {
	name = strings.ToLower(name)
	cmd, ok := CommandTable[name]
	if !ok {
		return new(unknownCommand)
	}
	return cmd
}

type unknownCommand struct{}
type cmdSet struct{}
type cmdGet struct{}

func (*unknownCommand) Exec(c *Client, r *protocol.Request) error {
	return c.ReplyError("unknown command")
}

func (*cmdGet) Exec(c *Client, r *protocol.Request) error {
	v := c.db.LookupKey(r.GetString(1))
	if v == nil {
		return c.ReplyEmpty()
	}

	if v.ObjType != dt.ObjString {
		return c.Reply(fmt.Sprintf("get wrong type, excepted %d, actually %d", dt.ObjString, v.ObjType))
	}

	return c.Reply(v.Ptr)
}

// SET key value [EX <seconds>]
func (*cmdSet) Exec(c *Client, r *protocol.Request) error {
	key, value, ex := r.GetString(1), r.GetString(2), r.GetString(3)

	var when int64
	if ex != "" {
		when = time.Now().UnixNano() / int64(time.Millisecond)
	} else {
		when = -1
	}

	return nil
}
