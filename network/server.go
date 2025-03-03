package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"projectx/core"
	"projectx/crypto"
	"projectx/types"
	"sync"

	"time"

	"github.com/go-kit/log"
)

var defaultBlockTime = time.Second * 5

//搭建自己本地的传输层协议，而非使用tcp或者udp协议
//本网络中区块链的数据传输都是通过该协议进行的
//server作为节点同时也作为processer

type ServerOpts struct {
	ID            string
	ListenAddr    string
	SeedNodes     []string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	PrivateKey    *crypto.PrivateKey
	blockTime     time.Duration
}

type Server struct {
	ServerOpts
	transport *TCPTransport //TCPTransport实现接口，接口不应该作为指针吧？
	peerCh    chan *TCPPeer

	mu sync.RWMutex

	peerMap     map[net.Addr]*TCPPeer
	blockTime   time.Duration
	memPool     *TxPool
	blockchain  *core.Blockchain
	isVaildator bool          //是验证器节点还是普通节点
	rpcCh       chan RPC      // node exchange message with connected node from rpcCh
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

	peerCh := make(chan *TCPPeer)
	transport := NewTCPTransport(string(opts.ListenAddr), peerCh)
	s := &Server{
		transport:   transport,
		peerCh:      peerCh,
		peerMap:     make(map[net.Addr]*TCPPeer),
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

	return s, nil
}

func (s *Server) Start() {

	s.transport.Start()
	time.Sleep(1 * time.Second)

	// the channel of all transports connected to server are listened
	s.boostrapNodes()

free:
	//一直检查Server中来往的rpc，如果没有判断是否要退出
	for {
		select {
		// connect to new peer and build rpcCh as communication channel
		case peer := <-s.transport.peerCh:

			//TODO: add mutex PLZ!!!
			// s.mu.RLock()
			s.peerMap[peer.conn.RemoteAddr()] = peer
			// s.mu.RUnlock()

			s.Logger.Log("msg", "peer added to the server", "outgoing", peer.Outgoing, "addr", peer.conn.RemoteAddr())
			// fmt.Printf("s.ListenAddr: %+v, peer.conn.LocalAddr(): %+v, peer.conn.RemoteAddr()%+v\n", s.ListenAddr, peer.conn.LocalAddr(), peer.conn.RemoteAddr())
			go peer.readLoop(s.rpcCh)

		case rpc := <-s.rpcCh:
			// handle rpc message
			// type of msg is *DecodedMessage
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
				continue //若解码失败，则跳过该rpc，继续执行ProcessMessage会出错
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
//
//	func (s *Server) boostrapNodes() error {
//		for _, tr := range s.Transports {
//			if s.Transport.Addr() != tr.Addr() {
//				if err := s.Transport.Connect(tr); err != nil {
//					s.Logger.Log("error", "couldn't connect to remote", err)
//					return err
//				}
//				// Send the getStatusMessage so we can snyc(if needed)
//				if err := s.sendGetStatusMessage(tr); err != nil { // && tr.CheckConnection(s.Transport.Addr())
//					s.Logger.Log("error", "sendGetStatusMessage", err)
//					return err
//				}
//			}
//		}
//		return nil
//	}
func (s *Server) boostrapNodes() error {
	for _, addr := range s.SeedNodes {

		// 不开协程会导致连接错误:即主程序未进入for循环，而子协程已经开始连接，导致连接出错???未搞清楚
		go func(addr string) {
			// 当你使用 net.Dial("tcp", addr) 连接到服务器时，conn.LocalAddr() 返回的地址是 本地机器的 IP 地址和一个随机的本地端口号。这个随机端口号是由操作系统分配的，用于标识本地机器的出站连接。
			// 为什么是随机端口？因为 TCP 协议是可靠的，所以需要确保两端建立连接后，两端的端口号是相同的。但是，如果本地机器的 IP 地址发生变化，或者本地机器的防火墙设置不当，导致无法建立连接，那么随机端口号就派上了用场。
			//服务器监听在 :3000，客户端连接到 :4000。
			// 如果服务器和客户端在同一台机器上，连接的四元组可能是：
			// 源 IP: 127.0.0.1
			// 源端口: 56752（客户端随机端口）
			// 目标 IP: 127.0.0.1
			// 目标端口: 4000
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				s.Logger.Log("error", "couldn't connect to remote", "addr", addr, "err", err)
				return
				// continue
			}
			fmt.Println("trying to connect to ", addr)
			peer := &TCPPeer{
				conn: conn,
			}
			s.peerCh <- peer
			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("msg", "send get satatus error", err)
			}
		}(addr)
	}
	return nil
}

// 传输层！！！！
// Server come true Processor interface
// 函数的参数是再复制一份数据。参数使用pointer是防止传入的参数数据过大。
func (s *Server) ProcessMessage(msg *DecodedMessage) error {

	// fmt.Printf("%s receiving message from %s\n", s.Transport.Addr(), msg.From)
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
	case *BlocksMessage:
		return s.processBlocksMessage(msg.From, t)
	default:
		s.Logger.Log("msg", "unknown message type", "type", fmt.Sprintf("%T", t))
	}
	return nil
}

// send status message, if height lower than current height, we need to get block information
// TODO: Remove the logic from the main function to here
// Normally Transport which is our own transport should do the trick.
func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
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

func (s *Server) processBlocksMessage(from net.Addr, data *BlocksMessage) error {
	fmt.Printf("receive blocks message!!!!!!!!!!!!!!! from %+v\n", from)

	for _, block := range data.Blocks {
		fmt.Printf("BLOCK=>%+v", block)
		if err := s.blockchain.AddBlock(block); err != nil {
			fmt.Println("222222222222", err) //TODO:fix it
			return err
		}
	}
	fmt.Printf("BLOCKS=>+%v\n", s.blockchain)
	return nil
}

func (s *Server) processGetBlockMessage(from net.Addr, data *GetBlockMessage) error {
	fmt.Printf("get block message from %+v => %+v\n", from, data)
	blocks := []*core.Block{}
	if data.To == 0 {
		// i := 1，because we already have genesis block.
		for i := 1; i <= int(s.blockchain.Height()); i++ {
			block, err := s.blockchain.GetBlock(uint32(i))
			if err != nil {
				return err
			}
			blocks = append(blocks, block)

		}
	}

	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}
	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeBlocks, buf.Bytes())

	s.mu.RLock()
	defer s.mu.RUnlock()
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("unknown peer %s", from)
	}

	return peer.Send(msg.Bytes())
}

func (s *Server) processGetStatusMessage(from net.Addr, data *GetStatusMessage) error {
	// !!!!!!!!
	// BUG: we weil get mesage just  like "REMOTE_1 receiving GetStatus message from REMOTE_1", but the function named SendMessage in localtransport.go cannot send message to ownself.
	// !!!!!!!!
	// fmt.Printf("=> %s receiving GetStatus message from %s => %+v\n", s.Transport.Addr(), from, data)

	s.mu.RLock()
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("unknown peer %s", from)
	}
	s.mu.RUnlock()

	s.Logger.Log("msg", "receive GetStatus message from", "from", from)
	StatusMessage := &StatusMessage{
		ID:            s.ID,
		CurrentHeight: s.blockchain.Height(),
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(StatusMessage); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) processStatusMessage(from net.Addr, data *StatusMessage) error {
	fmt.Printf("=> %s receiving GetStatus Response message from %s => %+v\n", s.ListenAddr, from, data)
	if s.blockchain.Height() >= data.CurrentHeight {
		s.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.blockchain.Height(), "remoteHeight", data.CurrentHeight)
		return nil
	}

	//TODO: in this case, we 100% sure that the node has blocks heighter than us

	getBlockMessage := &GetBlockMessage{
		From: data.CurrentHeight,
		To:   s.blockchain.Height(),
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(getBlockMessage); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
	s.mu.RLock()
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("unknown peer %s", from)
	}
	s.mu.RUnlock()
	// fmt.Println("TCPPeerMap", s.peerMap)
	peer.Send(msg.Bytes())
	return nil
}

func (s *Server) initTransports(transports []Transport) {
	// rpc from all transports connect to server will consume
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	for addr, peer := range s.peerMap {
		if err := peer.Send(payload); err != nil {
			s.Logger.Log("error", "send message to peer", "addr", addr, "err", err)
			continue
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
