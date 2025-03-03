package network

import (
	"fmt"
	"net"
	"sync"
)

type LocalTransport struct {
	addr      net.Addr
	consumeCh chan RPC
	lock      sync.RWMutex
	peers     map[net.Addr]*LocalTransport
}

func NewLocalTransport(addr net.Addr) *LocalTransport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC, 1024),
		peers:     make(map[net.Addr]*LocalTransport),
	}
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

func (t *LocalTransport) Connect(tr Transport) error {
	trans := tr.(*LocalTransport)
	t.lock.Lock()
	defer t.lock.Unlock()

	t.peers[tr.Addr()] = trans
	return nil
}

func (t *LocalTransport) Addr() net.Addr {
	return t.addr
}

func (t *LocalTransport) SendMessage(to net.Addr, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.addr == to {
		return nil
	}
	// check if peer is connected
	_, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("%s colud not send message to unkonwn peer %s", t.Addr(), to)
	}
	// peer.consumeCh <- RPC{From: t.addr, Proload: bytes.NewReader(payload)}
	return nil
}

func (t *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range t.peers {
		if err := t.SendMessage(peer.Addr(), payload); err != nil {
			return err
		}
	}
	return nil
}

// func (t *LocalTransport) CheckConnection(tr NetAddr) bool {
// 	t.lock.RLock()
// 	defer t.lock.RUnlock()
// 	if _, ok := t.peers[tr]; ok {
// 		return true
// 	}
// 	return false
// }
