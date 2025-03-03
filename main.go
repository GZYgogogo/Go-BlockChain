package main

import (
	"bytes"
	"log"
	"net"
	"projectx/core"
	"projectx/crypto"
	"projectx/network"
	"time"
)

// var (
// 	transports = []network.Transport{
// 		// network.NewLocalTransport("LOCAL"),
// 		// network.NewLocalTransport("REMOTE_0"),
// 		// network.NewLocalTransport("REMOTE_1"),
// 		// network.NewLocalTransport("REMOTE_2"),
// 		// network.NewLocalTransport("Late_Remote"),
// 	}
// 	// remoteTransport = transports[1:4]
// )

// BUG!
// we will get error about read error, when we run code more than 20 seconds.
// BUG!
func main() {
	priv := crypto.GeneratePrivateKey()

	localNode := makeServer("LOCAL_NDOE", &priv, ":3000", []string{":4000"})

	go localNode.Start()

	remoteNodeA := makeServer("REMOTE_NODE_A", nil, ":4000", []string{":3000"})
	go remoteNodeA.Start()

	// remoteNodeB := makeServer("REMOTE_NODE_B", nil, ":5000", []string{})
	// go remoteNodeB.Start()
	time.Sleep(1 * time.Second)
	go tcpTester()

	time.Sleep(7 * time.Second)
	lateNode := makeServer("LATE_NODE", nil, ":5000", []string{":3000"})
	go lateNode.Start()

	select {}
}
func makeServer(id string, privKey *crypto.PrivateKey, address string, seedNodess []string) *network.Server {
	opts := network.ServerOpts{
		ListenAddr: address,
		ID:         id,
		PrivateKey: privKey,
		SeedNodes:  seedNodess,
	}
	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func tcpTester() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		panic(err)
	}

	privKey := crypto.GeneratePrivateKey()
	// data := []byte(strconv.FormatInt(int64(rand.Intn(1000)), 10))
	tx := core.NewTransaction(contract())
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	_, err = conn.Write(msg.Bytes())
	if err != nil {
		panic(err)
	}
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

// func initRemoteServer(trs []network.Transport) {
// 	for i, tr := range trs {
// 		id := fmt.Sprintf("REMOTE_%d", i)
// 		s := makeServer(id, tr, nil)
// 		go s.Start()
// 	}
// }

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

func sendTransaction(tr network.Transport, to net.Addr) error {
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
