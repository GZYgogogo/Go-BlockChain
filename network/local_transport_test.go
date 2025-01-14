package network

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	LtA := NewLocalTransport("A").(*LocalTransport)
	LtB := NewLocalTransport("B").(*LocalTransport)
	LtA.Connect(LtB)
	LtB.Connect(LtA)
	assert.Equal(t, LtA.peers[LtB.Addr()], LtB)
	assert.Equal(t, LtB.peers[LtA.Addr()], LtA)
}

func TestSendMessage(t *testing.T) {
	LtA := NewLocalTransport("A").(*LocalTransport)
	LtB := NewLocalTransport("B").(*LocalTransport)
	LtA.Connect(LtB)
	LtB.Connect(LtA)

	msg := []byte("Hello, world!")
	assert.Nil(t, LtA.SendMessage(LtB.addr, msg))

	rpc := <-LtB.Consume()
	b, err := io.ReadAll(rpc.Proload)
	assert.Nil(t, err)
	assert.Equal(t, b, msg)
	assert.Equal(t, rpc.From, LtA.addr)
}

func TestBroadcast(t *testing.T) {
	LtA := NewLocalTransport("A").(*LocalTransport)
	LtB := NewLocalTransport("B").(*LocalTransport)
	LtC := NewLocalTransport("C").(*LocalTransport)

	LtA.Connect(LtB)
	LtA.Connect(LtC)
	msg := []byte("Hello, world!")

	assert.Equal(t, len(LtA.peers), 2)
	assert.Nil(t, LtA.Broadcast(msg))

	rpcB := <-LtB.Consume()
	b, err := io.ReadAll(rpcB.Proload)
	assert.Nil(t, err)
	assert.Equal(t, b, msg)

	rpcC := <-LtC.Consume()
	c, err := io.ReadAll(rpcC.Proload)
	assert.Nil(t, err)
	assert.Equal(t, c, msg)
}
