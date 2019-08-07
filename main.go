package main

import (
	"github.com/kzinglzy/godis/server"
)

func main() {
	s, err := server.MakeServer(":7777")
	if err != nil {
		panic("failed to create godis server: " + err.Error())
	}
	s.Run()
	defer s.Close()
}
