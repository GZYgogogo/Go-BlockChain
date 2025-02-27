package network

type GetBlockMessage struct{}

type GetStatusMessage struct{}

type StatusMessage struct {
	//the ID of the server
	ID            string
	Version       uint32
	CurrentHeight uint32
}
