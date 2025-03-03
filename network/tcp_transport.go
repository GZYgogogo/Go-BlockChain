package network

import (
	"bytes"
	"fmt"
	"net"
)

type TCPPeer struct {
	conn     net.Conn
	Outgoing bool
}

func (p *TCPPeer) readLoop(rpcCh chan RPC) {
	buf := make([]byte, 2024)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			fmt.Printf("read error from: %v\n", p.conn.RemoteAddr())
			continue
		}
		msg := buf[:n]
		// fmt.Println(string(msg))
		rpcCh <- RPC{
			From:    p.conn.RemoteAddr(),
			Proload: bytes.NewReader(msg),
		}
	}
}

func (p *TCPPeer) Send(msg []byte) error {
	_, err := p.conn.Write(msg)
	return err
}

type TCPTransport struct {
	peerCh     chan *TCPPeer
	listenAddr string
	listener   net.Listener
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln

	go t.acceptLoop()
	fmt.Println("TCP transport listening to port ", t.listenAddr)
	return nil
}

func NewTCPTransport(listenAddr string, peerCh chan *TCPPeer) *TCPTransport {
	return &TCPTransport{
		listenAddr: listenAddr,
		peerCh:     peerCh,
	}
}

// listen if the peer is trying to be connected
func (t *TCPTransport) acceptLoop() (net.Conn, error) {
	for {
		//  阻塞等待，直到有一个新的客户端连接到该listener
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("accept error from: %v\n", t.listenAddr)
			continue
		}
		peer := &TCPPeer{
			conn: conn,
		}
		t.peerCh <- peer
		// add TCPPeer to peer channel
		fmt.Printf("new incomming connection =>%+v\n", conn)
		// go t.readLoop(peer)
	}
}
