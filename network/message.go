package network

import "projectx/core"

type GetBlockMessage struct {
	From uint32
	To   uint32 //if To is 0, the maximum blocks will be returned
}

type BlocksMessage struct {
	Blocks []*core.Block
}

type GetStatusMessage struct{}

type StatusMessage struct {
	//the ID of the server
	ID            string
	Version       uint32
	CurrentHeight uint32
}
