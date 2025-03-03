package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"projectx/core"

	"github.com/sirupsen/logrus"
)

// data encode=> msgData => NewMessage(msgType, msgData)Message => rpc{From, Proload:Message}

type MessageType byte

const (
	MessageTypeTx        MessageType = 0x1
	MessageTypeBlock     MessageType = 0x2
	MessageTypeGetBlocks MessageType = 0x3
	MessageTypeStatus    MessageType = 0x4
	MessageTypeGetStatus MessageType = 0x5
	MessageTypeBlocks    MessageType = 0x6
)

type RPC struct {
	From    net.Addr
	Proload io.Reader
}

type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(t MessageType, msg []byte) *Message {
	return &Message{
		Header: t,
		Data:   msg,
	}
}

func (msg *Message) Bytes() []byte {
	buf := bytes.Buffer{}
	gob.NewEncoder(&buf).Encode(msg)
	return buf.Bytes()
}

type DecodedMessage struct {
	From net.Addr
	Data any
}

type RPCDecodeFunc func(RPC) (*DecodedMessage, error)

func DefaultRPCDecodeFunc(rpc RPC) (*DecodedMessage, error) {
	msg := Message{}
	//decode preload data of rpc, preloade date construct by Header and Data
	if err := gob.NewDecoder(rpc.Proload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message from: %s: %s", rpc.From, err)
	}

	logrus.WithFields(logrus.Fields{
		"from": rpc.From,
		"type": msg.Header,
	}).Debug("received new message")

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		//decode tx data from msg.Data
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageTypeBlock:
		b := new(core.Block)
		if err := b.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: b,
		}, nil
	case MessageTypeGetBlocks:
		getBlocksMessage := new(GetBlockMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(getBlocksMessage); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: getBlocksMessage,
		}, nil
	case MessageTypeStatus:
		statusMessage := new(StatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(statusMessage); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: statusMessage,
		}, nil
	case MessageTypeGetStatus:
		return &DecodedMessage{
			From: rpc.From,
			Data: &GetStatusMessage{},
		}, nil
	case MessageTypeBlocks:
		blocksMessage := new(BlocksMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(blocksMessage); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: blocksMessage,
		}, nil
	default:
		return nil, fmt.Errorf("invaild message header: %x", msg.Header)
	}
}

type RPCProcessor interface {
	ProcessMessage(*DecodedMessage) error
}
