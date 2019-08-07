package server

import (
	"strconv"
	"strings"

	"github.com/kzinglzy/godis/dt"
	"github.com/kzinglzy/godis/server/protocol"
)

// CommandTable .
var CommandTable = map[string]Command{
	CmdNamePing:     new(cmdPing),
	CmdNameGet:      new(cmdGet),
	CmdNameSet:      new(cmdSet),
	CmdNameTTL:      new(cmdTTL),
	CmdNameExpire:   new(cmdExpire),
	CmdNameExpireAt: new(cmdExpireAt),
	CmdNamePush:     new(cmdPush),
	CmdNamePop:      new(cmdPop),
	CmdNameRange:    new(cmdRange),
}

type unknownCommand struct{}
type cmdPing struct{}
type cmdGet struct{}
type cmdSet struct{}
type cmdTTL struct{}
type cmdExpire struct{}
type cmdExpireAt struct{}
type cmdPush struct{}
type cmdPop struct{}
type cmdRange struct{}

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

func (*unknownCommand) Exec(c *Client, r *protocol.Request) error {
	return c.ReplyError("unknown command")
}

func (*cmdGet) Exec(c *Client, r *protocol.Request) error {
	v := c.db.Get(r.ArgvAt(1))
	if v == nil {
		return c.ReplyEmpty()
	}

	if v.ObjType != dt.ObjString {
		return c.Reply("key holding a wrong kind of value")
	}
	return c.Reply(v.Ptr.(string))
}

// SET key value [EX <seconds>] [nx]
func (*cmdSet) Exec(c *Client, r *protocol.Request) error {
	if r.ArgCount() < 3 {
		return c.ReplyError("wrong number of arguments for 'set' command")
	}

	key, value, expire := r.ArgvAt(1), r.ArgvAt(2), r.ArgvAt(3)

	if (r.ArgvAt(3) == FlagSetNX || r.ArgvAt(4) == FlagSetNX) && c.db.Get(key) != nil {
		return c.ReplyEmpty()
	}

	obj := dt.NewObj(dt.ObjString, value)
	c.db.Set(key, obj)
	godisServer.dirty++

	if ex, err := strconv.ParseInt(expire, 10, 64); ex != 0 && err == nil {
		when := mstime() + ex*1000
		c.db.setExpire(key, when)
	}

	return c.Reply("OK")
}

func (*cmdTTL) Exec(c *Client, r *protocol.Request) error {
	key := r.ArgvAt(1)
	return c.ReplyInt(c.db.ttl(key))
}

func (*cmdPing) Exec(c *Client, r *protocol.Request) error {
	return c.Reply("PONG")
}

func (*cmdExpire) Exec(c *Client, r *protocol.Request) error {
	if r.ArgCount() != 3 {
		return c.ReplyError("wrong number of arguments for 'expire' command")
	}

	key, expire := r.ArgvAt(1), r.ArgvAt(2)
	if c.db.Get(key) == nil {
		return c.ReplyInt(0)
	}

	ex, err := strconv.ParseInt(expire, 10, 64)
	if err != nil {
		return c.Reply("wrong type of expire argument")
	}
	when := ex*1000 + mstime()
	if when < mstime() {
		c.db.deleteKey(key)
	} else {
		c.db.setExpire(key, when)
	}
	godisServer.dirty++
	return c.ReplyInt(1)
}

func (*cmdExpireAt) Exec(c *Client, r *protocol.Request) error {
	if r.ArgCount() != 3 {
		return c.ReplyError("wrong number of arguments for 'expireat' command")
	}

	key, ms := r.ArgvAt(1), r.ArgvAt(2)
	if c.db.Get(key) == nil {
		return c.ReplyInt(0)
	}

	when, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return c.Reply("wrong type of expireat argument")
	}

	if when < mstime() {
		c.db.deleteKey(key)
	} else {
		c.db.setExpire(key, when)
	}
	godisServer.dirty++
	return c.ReplyInt(1)
}

func (*cmdPush) Exec(c *Client, r *protocol.Request) error {
	if r.ArgCount() < 3 {
		return c.ReplyError("wrong number of arguments for 'push' command")
	}

	key := r.ArgvAt(1)
	obj := c.db.Get(key)
	if obj != nil && obj.ObjType != dt.ObjList {
		return c.ReplyError("key holding a wrong kind of value")
	}

	list := []string{}
	if obj != nil {
		list = obj.Ptr.([]string)
	}

	var pushed int64
	for i := 2; i < r.ArgCount(); i++ {
		list = append(list, r.ArgvAt(i))
		pushed++
	}

	obj = dt.NewList(dt.ObjList, list)
	c.db.Add(key, obj)
	godisServer.dirty += pushed
	return c.ReplyInt(pushed)
}

func (*cmdRange) Exec(c *Client, r *protocol.Request) error {
	start, err := strconv.ParseInt(r.ArgvAt(2), 10, 64)
	if err != nil {
		return c.ReplyError("value is not an illegal integer")
	}
	end, err := strconv.ParseInt(r.ArgvAt(3), 10, 64)
	if err != nil {
		return c.ReplyError("value is not an illegal integer")
	}

	var ret []string

	key := r.ArgvAt(1)
	old := c.db.Get(key)
	if old == nil {
		return c.ReplyList(ret)
	}
	if old.ObjType != dt.ObjList {
		return c.ReplyError("key holding a wrong kind of value")
	}

	list := old.Ptr.([]string)
	length := int64(len(list))
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}
	if end < 0 {
		end = length + end
	}

	if start > end || start >= length {
		return c.ReplyList(ret)
	}
	if end >= length {
		end = length - 1
	}

	for i := start; i <= end; i++ {
		ret = append(ret, list[i])
	}
	return c.ReplyList(ret)
}

func (s *cmdPop) Exec(c *Client, r *protocol.Request) error {
	if r.ArgCount() != 2 {
		return c.ReplyError("wrong number of arguments for 'pop' command")
	}

	key := r.ArgvAt(1)
	obj := c.db.Get(key)
	if obj == nil {
		return c.ReplyEmpty()
	}

	list, ok := obj.Ptr.([]string)
	if len(list) == 0 || !ok {
		return c.ReplyEmpty()
	}

	v, list := list[len(list)-1], list[:len(list)-1]
	c.db.Add(key, dt.NewList(dt.ObjList, list))
	godisServer.dirty++
	return c.Reply(v)
}
