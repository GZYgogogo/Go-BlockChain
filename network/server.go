package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"projectx/core"
	"projectx/crypto"
	"projectx/types"

	"time"

	"github.com/go-kit/log"
)

var defaultBlockTime = time.Second * 5

//搭建自己本地的传输层协议，而非使用tcp或者udp协议
//本网络中区块链的数据传输都是通过该协议进行的
//server作为节点同时也作为processer

type ServerOpts struct {
	ID            string
	Transport     Transport
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	PrivateKey    *crypto.PrivateKey
	blockTime     time.Duration
}

type Server struct {
	ServerOpts
	blockTime   time.Duration
	memPool     *TxPool
	blockchain  *core.Blockchain
	isVaildator bool //是验证器节点还是普通节点
	rpcCh       chan RPC
	quitCh      chan struct{} //退出信号
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.blockTime == time.Duration(0) {
		opts.blockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}
	blockchain, err := core.NewBlockchain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}
	s := &Server{
		ServerOpts:  opts,
		isVaildator: opts.PrivateKey != nil,
		blockchain:  blockchain,
		blockTime:   opts.blockTime,
		memPool:     NewTxPool(100),
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}
	//If we dont got any processor from server options, we going to
	//use the server as default
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}
	if s.isVaildator {
		go s.VaildatorLoop()
	}
	// tr := s.Transports[0].(*LocalTransport)
	// fmt.Printf("%+v\n", tr.peers)
	// for _, tr := range s.Transports {
	// 	if err := s.sendGetStatusMessage(tr); err != nil {
	// 		s.Logger.Log("send get satatus error", err)
	// 	}
	// }

	s.boostrapNodes()
	return s, nil
}

func (s *Server) Start() {
	// the channel of all transports connected to server are listened
	s.initTransports(s.Transports)

free:
	//一直检查Server中来往的rpc，如果没有判断是否要退出
	for {
		select {
		case rpc := <-s.rpcCh:
			// handle rpc message
			// type of msg is *DecodedMessage
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
			}
			// process decoded message
			if err = s.RPCProcessor.ProcessMessage(msg); err != nil {
				if err != core.ERrrBlockKnown {
					s.Logger.Log("error", err)
				}

			}

		// quit signal
		case <-s.quitCh:
			break free
		}
	}
	fmt.Println("Server showdown")
}

func (s *Server) VaildatorLoop() {
	//创建一个定时器
	ticker := time.NewTicker(s.blockTime)

	s.Logger.Log(
		"msg", "starting vaildator loop",
		"blockTime", s.blockTime,
	)

	for {
		<-ticker.C
		s.createNewBlock()
	}
}

// server connect to all nodes in transports list
func (s *Server) boostrapNodes() error {
	for _, tr := range s.Transports {
		if s.Transport.Addr() != tr.Addr() {
			if err := s.Transport.Connect(tr); err != nil {
				s.Logger.Log("error", "couldn't connect to remote", err)
				return err
			}
			// Send the getStatusMessage so we can snyc(if needed)
			if err := s.sendGetStatusMessage(tr); err != nil { // && tr.CheckConnection(s.Transport.Addr())
				s.Logger.Log("error", "sendGetStatusMessage", err)
				return err
			}
		}
	}
	return nil
}

// 传输层！！！！
// Server come true Processor interface
// 函数的参数是再复制一份数据。参数使用pointer是防止传入的参数数据过大。
func (s *Server) ProcessMessage(msg *DecodedMessage) error {
	// fmt.Printf("%s receiving message form %s\n", s.Transport.Addr(), msg.From)
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.ProcessTransaction(t)
	case *core.Block:
		return s.ProcessBlock(t)
	case *GetBlockMessage:
		return s.processGetBlockMessage(msg.From, t)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From, t)
	default:
		s.Logger.Log("msg", "unknown message type", "type", fmt.Sprintf("%T", t))
	}
	return nil
}

// send status message, if height lower than current height, we need to get block information
// TODO: Remove the logic from the main function to here
// Normally Transport which is our own transport should do the trick.
func (s *Server) sendGetStatusMessage(tr Transport) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())

	fmt.Printf("=> %s sending GetStatus message to %s\n", s.Transport.Addr(), tr.Addr())
	if err := s.Transport.SendMessage(tr.Addr(), msg.Bytes()); err != nil {
		return err
	}

	//statusMessage
	return nil
}

func (s *Server) ProcessTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	//验证交易
	if err := tx.Verify(); err != nil {
		return err
	}
	//将交易是否已经加入内存池中
	if s.memPool.Contains(hash) {
		return nil
	}
	tx.SetFirstSceen(time.Now().Unix())

	// s.Logger.Log(
	// 	"msg", "adding new tx to mempool",
	// 	"hash", hash,
	// 	"mempoolPending", s.memPool.PendingCount(),
	// )

	// broacast tx to all nodes connected to server
	go s.BroadcastTx(tx)

	s.memPool.Add(tx)

	return nil
}

func (s *Server) ProcessBlock(b *core.Block) error {
	if err := s.blockchain.AddBlock(b); err != nil {
		return err
	}
	go s.BroadcastBlock(b)

	return nil
}

func (s *Server) processGetBlockMessage(form NetAddr, data *GetBlockMessage) error {
	panic("here")
	fmt.Printf("processGetBlockMessage %+v\n", data)
	return nil
}

func (s *Server) processGetStatusMessage(form NetAddr, data *GetStatusMessage) error {

	// !!!!!!!!
	// BUG: we weil get mesage just  like "REMOTE_1 receiving GetStatus message from REMOTE_1", but the function named SendMessage in localtransport.go cannot send message to ownself.
	// !!!!!!!!
	// fmt.Printf("=> %s receiving GetStatus message from %s => %+v\n", s.Transport.Addr(), form, data)

	StatusMessage := &StatusMessage{
		ID:            s.ID,
		CurrentHeight: s.blockchain.Height(),
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(StatusMessage); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeStatus, buf.Bytes())
	return s.Transport.SendMessage(form, msg.Bytes())
}

func (s *Server) processStatusMessage(form NetAddr, data *StatusMessage) error {
	fmt.Printf("=> %s receiving GetStatus Response message from %s => %+v\n", s.Transport.Addr(), form, data)
	if s.blockchain.Height() >= data.CurrentHeight {
		s.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.blockchain.Height(), "remoteHeight", data.CurrentHeight)
		return nil
	}

	//TODO: in this case, we 100% sure that the node has blocks heighter than us
	fmt.Println(11111111111111)
	getBlockMessage := &GetBlockMessage{}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(getBlockMessage); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
	s.Transport.SendMessage(form, msg.Bytes())
	return nil
}

func (s *Server) initTransports(transports []Transport) {
	// rpc form all transports connect to server will consume
	for _, tr := range transports {
		go func(tr Transport) {
			//listen channel to receive rpc message
			for rpc := range tr.Consume() {
				//将rpc消息全部存放至Server的channel中
				s.rpcCh <- rpc
			}
		}(tr)
	}
}

func (s *Server) Broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) BroadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(b); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeBlock, buf.Bytes())
	return s.Broadcast(msg.Bytes())
}

func (s *Server) BroadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(tx); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.Broadcast(msg.Bytes())
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.blockchain.GetHeader(s.blockchain.Height())
	if err != nil {
		return err
	}

	txx := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)
	if err != nil {
		return err
	}
	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}
	if err := s.blockchain.AddBlock(block); err != nil {
		return err
	}

	//TODO:pending pool of tx should only reflect on vaildator nodes,
	//Right now "normal nodes" does not have their pending pool.
	s.memPool.ClearPending()
	go s.BroadcastBlock(block)

	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Timestamp: 0000000,
		Height:    0,
	}
	block, _ := core.NewBlock(header, nil)
	return block
}
