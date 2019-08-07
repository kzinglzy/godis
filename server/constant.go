package server

// server
const (
	MaxIOEventsPerLoop = 10
	AOFFileName        = "godis.aof"
)

// db
const (
	ActiveExpireCycleLookupsPerLoop = 20
	MaxCycleTimeLimitUSPerLoop      = 25000 // us
)

// maxmemory strategies
const (
	MaxmemoryAllkeysLRU = iota
	MaxmemoryAllkeysRandom
	MaxmemoryNoEviction
	MaxmemorySamples = 5
	MaxMemory        = 1024 * 1024 * 16
)

// evict pool
const (
	EvPoolSize = 15
)

// command
const (
	CmdNameSet      = "set"
	CmdNameGet      = "get"
	CmdNamePing     = "ping"
	CmdNameTTL      = "ttl"
	CmdNameExpire   = "expire"
	CmdNameExpireAt = "expireat"
	CmdNamePush     = "push"
	CmdNamePop      = "pop"
	CmdNameRange    = "range"
	// CmdNameZadd     = "zset"

	FlagSetNX = "nx"
)

// aof
const (
	AOFRewriteMinSize = 64 * 1024 * 1024

	AOFFsyncEverysec = 1
	AOFFsyncAlways   = 2
)
