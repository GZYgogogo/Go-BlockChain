package main

import (
	"bytes"
	"fmt"
	"log"
	"projectx/core"
	"projectx/crypto"
	"projectx/network"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	trLocal := network.NewLocalTransport("local")
	trRemoteA := network.NewLocalTransport("remoteA")
	trRemoteB := network.NewLocalTransport("remoteB")
	trRemoteC := network.NewLocalTransport("remoteC")

	trLocal.Connect(trRemoteA)
	trRemoteA.Connect(trRemoteB)
	trRemoteB.Connect(trRemoteC)
	trRemoteA.Connect(trLocal)

	go func() {
		for {
			if err := sendTransaction(trRemoteA, trLocal.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(3 * time.Second)
		}
	}()

	// go func() {
	// 	time.Sleep(8 * time.Second)
	// 	trLate := network.NewLocalTransport("Late_Remote")
	// 	trRemoteA.Connect(trLate)
	// 	lateServer := makeServer("LATE_REMOTE", trLate, nil)
	// 	go lateServer.Start()
	// }()

	initRemoteServer([]network.Transport{trRemoteA, trRemoteB, trRemoteC})
	privKey := crypto.GeneratePrivateKey()
	s := makeServer("LOCAL", trLocal, &privKey)
	s.Start()
}

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
		PrivateKey: privKey,
		Transports: []network.Transport{tr},
	}
	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

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
