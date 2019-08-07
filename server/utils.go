package server

import "time"

func mstime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func ustime() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

func usedmemory() int64 {
	/* FIXME how do we trace the memory useage painless?
	 */
	return godisServer.db.store.Used()
}
