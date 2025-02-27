package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"projectx/core"
	"projectx/crypto"
	"projectx/network"
	"time"
)

var (
	transports = []network.Transport{
		network.NewLocalTransport("LOCAL"),
		// network.NewLocalTransport("REMOTE_0"),
		// network.NewLocalTransport("REMOTE_1"),
		// network.NewLocalTransport("REMOTE_2"),
		network.NewLocalTransport("Late_Remote"),
	}
	// remoteTransport = transports[1:4]
)

func main() {
	tr := network.NewTCPTransport(":3000")
	go tr.Start()

	time.Sleep(1 * time.Second)
	tcpTester()

	select {}
}

func tcpTester() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("hello world!"))
}

// func main() {
// 	trLocal := transports[0]
// 	// remote_1 := transports[1]
// 	// remote_2 := transports[2]
// 	// remote_3 := transports[3]
// 	trLate := transports[len(transports)-1]

// 	// initRemoteServer(remoteTransport)
// 	privKey := crypto.GeneratePrivateKey()
// 	s := makeServer(string(trLocal.Addr()), trLocal, &privKey)
// 	// go func() {
// 	// 	for {
// 	// 		if err := sendTransaction(remote_1, trLocal.Addr()); err != nil {
// 	// 			logrus.Error(err)
// 	// 		}
// 	// 		time.Sleep(3 * time.Second)
// 	// 	}
// 	// }()

// 	// sendGetStatusMessage(trRemoteA, "remoteB")

// 	go func() {
// 		time.Sleep(8 * time.Second)
// 		lateServer := makeServer(string(trLate.Addr()), trLate, nil)
// 		go lateServer.Start()
// 	}()
// 	s.Start()
// 	// trLocal := network.NewLocalTransport("local")
// 	// trRemoteA := network.NewLocalTransport("remoteA")
// 	// trRemoteB := network.NewLocalTransport("remoteB")
// 	// trRemoteC := network.NewLocalTransport("remoteC")

// 	// trLocal.Connect(trRemoteA)
// 	// trRemoteA.Connect(trRemoteB)
// 	// trRemoteB.Connect(trRemoteA)
// 	// trRemoteB.Connect(trRemoteC)
// 	// trRemoteA.Connect(trLocal)

// }

func initRemoteServer(trs []network.Transport) {
	for i, tr := range trs {
		id := fmt.Sprintf("REMOTE_%d", i)
		s := makeServer(id, tr, nil)
		go s.Start()
	}
}

func makeServer(id string, tr network.Transport, privKey *crypto.PrivateKey) *network.Server {
	opts := network.ServerOpts{
		ID:         id,
		Transport:  tr,
		PrivateKey: privKey,
		Transports: transports,
	}
	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

// send status message, if height lower than current height, we need to get block information
// func sendGetStatusMessage(tr network.Transport, to network.NetAddr) error {
// 	var (
// 		getStatusMsg = new(network.GetStatusMessage)
// 		buf          = new(bytes.Buffer)
// 	)

// 	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
// 		return err
// 	}
// 	msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())

// 	return tr.SendMessage(to, msg.Bytes())
// }

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	privKey := crypto.GeneratePrivateKey()
	// data := []byte(strconv.FormatInt(int64(rand.Intn(1000)), 10))
	tx := core.NewTransaction(contract())
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())
}

func contract() []byte {
	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	return append(data, pushFoo...)
}
