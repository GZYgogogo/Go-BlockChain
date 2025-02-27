package network

import (
	"fmt"
	"net"
)

type TCPPeer struct {
	conn net.Conn
}

type TCPTransport struct {
	listenAddr string
	listener   net.Listener
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: listenAddr,
	}
}

func (t *TCPTransport) readLoop(peer *TCPPeer) {
	buf := make([]byte, 2024)
	for {
		n, err := peer.conn.Read(buf)
		if err != nil {
			fmt.Printf("read error from: %v\n", peer.conn.RemoteAddr())
			continue
		}
		msg := buf[:n]
		fmt.Printf(string(msg))
	}
}

func (t *TCPTransport) acceptLoop() (net.Conn, error) {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("accept error from: %v\n", t.listenAddr)
			continue
		}
		peer := &TCPPeer{
			conn: conn,
		}
		fmt.Printf("new incomming connection => %+v\n", conn)
		go t.readLoop(peer)
	}
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln

	go t.acceptLoop()
	fmt.Println("TCP listening to port:", t.listenAddr)
	return nil
}
