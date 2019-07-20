package server

import "github.com/kzinglzy/godis/dt"

type Server struct {
	db Database
}

// Database implements the kv service
type Database struct {
	dict    dt.Dict
	expires dt.Dict
}
