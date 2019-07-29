package server

import (
	"fmt"
	"strconv"
	"strings"

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
	v := c.db.Get(r.GetString(1))
	if v == nil {
		return c.ReplyEmpty()
	}

	if v.ObjType != dt.ObjString {
		return c.Reply(fmt.Sprintf("get wrong type, excepted %d, actually %d", dt.ObjString, v.ObjType))
	}
	return c.Reply(v.Ptr.(string))
}

// SET key value [EX <seconds>]
func (*cmdSet) Exec(c *Client, r *protocol.Request) error {
	if r.ArgCount() < 3 {
		return c.ReplyError("wrong number of arguments for 'set' command")
	}

	key, value, expire := r.GetString(1), r.GetString(2), r.GetString(3)

	obj := dt.NewObj(dt.ObjString, value)
	c.db.setKey(key, obj)
	godisServer.dirty++

	if ex, err := strconv.ParseInt(expire, 10, 64); err != nil {
		when := mstime() + ex
		c.db.setExpire(key, when)
	}

	return c.Reply("OK")
}
