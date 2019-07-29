package server

import "time"

// mstime returns the UNIX time in milliseconds
func mstime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
