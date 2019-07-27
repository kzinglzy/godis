package server

import "github.com/kzinglzy/godis/protocol"

type Client struct {
	// parser protocol.Parser
	writer protocol.Writer
}
