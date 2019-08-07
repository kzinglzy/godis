package server

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kzinglzy/godis/server/protocol"
)

func feedAppendOnlyFile(r *protocol.Request) {
	// translate SETEX to SET and EXPIREAT
	if r.CommandName() == CmdNameSet && r.ArgvAt(3) != "" && r.ArgvAt(3) != FlagSetNX {
		setArgv := []string{CmdNameSet, r.ArgvAt(1), r.ArgvAt(2)}
		catAppendOnlyCommand(3, setArgv)

		ex, _ := strconv.ParseInt(r.ArgvAt(3), 10, 64)
		when := ex*1000 + mstime()
		expireArgv := []string{CmdNameExpireAt, r.ArgvAt(1), fmt.Sprintf("%d", when)}
		catAppendOnlyCommand(3, expireArgv)
	} else if r.CommandName() == CmdNameExpire {
		ex, _ := strconv.ParseInt(r.ArgvAt(2), 10, 64)
		when := ex*1000 + mstime()
		expireArgv := []string{CmdNameExpireAt, r.ArgvAt(1), fmt.Sprintf("%d", when)}
		catAppendOnlyCommand(3, expireArgv)
	} else {
		catAppendOnlyCommand(r.ArgCount(), r.Argv())
	}
}

func catAppendOnlyCommand(argc int, argv []string) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("*%d", argc))
	buf.WriteString("\r\n")
	for i := 0; i < argc; i++ {
		cmd := argv[i]
		buf.WriteString(fmt.Sprintf("$%d", len(cmd)))
		buf.WriteString("\r\n")
		buf.WriteString(cmd)
		buf.WriteString("\r\n")
	}

	godisServer.aofBuf = append(godisServer.aofBuf, []byte(buf.String())...)
}

func flushAppendOnlyFile(force bool) {
	if len(godisServer.aofBuf) == 0 {
		return
	}

	if godisServer.aofFsyncPolicy == AOFFsyncEverysec && !force {
		if godisServer.aofFlushPostponedStart == 0 {
			godisServer.aofFlushPostponedStart = mstime()
			return
		} else if mstime()-godisServer.aofFlushPostponedStart < 1000 {
			return
		}
	}

	n, err := godisServer.aof.Write(godisServer.aofBuf)
	if err != nil || n != len(godisServer.aofBuf) {
		panic("Can't recover from AOF write error, Exiting...")
	}

	godisServer.resetAofState()
	if godisServer.aofFsyncPolicy == AOFFsyncAlways {
		godisServer.aof.Sync()
	} else {
		if !godisServer.aofFlushInProgress {
			log.Printf("do fsync")
			go godisServer.aofFsync()
		}
	}
}

func rewriteAOFBackgroundIfNeed() {
	// TODO
}
