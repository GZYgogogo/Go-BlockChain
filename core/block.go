package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"projectx/crypto"
	"projectx/types"
	"time"
)

type Header struct {
	Version       uint32     // 版本号
	DataHash      types.Hash // 交易的root hash of merkel tree
	PrevBlockHash types.Hash // 前一个区块的hash
	Timestamp     int64      // 时间戳
	Height        uint32
	Nonce         uint64 // 随机数
}

func (h *Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	// 创建 gob 编码器，能够将 Go 数据结构编码为二进制形式，并写入到 buf 中。
	enc := gob.NewEncoder(buf)
	// 编码器对Block Header编码并存入到buf中。
	enc.Encode(h)
	return buf.Bytes()
}

type Block struct {
	*Header
	Transactions []*Transaction

	Vaildator crypto.PublicKey
	Signature *crypto.Signature
	// Cached version of the header hash
	hash types.Hash
}

func NewBlock(h *Header, txx []*Transaction) (*Block, error) {
	return &Block{
		Header:       h,
		Transactions: txx,
	}, nil
}

func NewBlockFromPrevHeader(prevHeader *Header, txx []*Transaction) (*Block, error) {
	dataHash, err := CalculateDataHash(txx)
	if err != nil {
		return nil, err
	}
	header := &Header{
		Version:       prevHeader.Version + 1,
		DataHash:      dataHash,
		PrevBlockHash: BlockHasher{}.Hash(prevHeader),
		Timestamp:     time.Now().Unix(),
		Height:        prevHeader.Height + 1,
	}
	return NewBlock(header, txx)
}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}
	return b.hash
}

func (b *Block) Sign(privKey crypto.PrivateKey) error {
	sig, err := privKey.Sign(b.Header.Bytes())
	if err != nil {
		return err
	}
	b.Vaildator = privKey.PublicKey()
	b.Signature = sig
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("block has no signature")
	}
	if !b.Signature.Verify(b.Vaildator, b.Header.Bytes()) {
		return fmt.Errorf("invalid block signature")
	}
	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}
	dataHash, err := CalculateDataHash(b.Transactions)
	if err != nil {
		return err
	}
	if dataHash != b.Header.DataHash {
		return fmt.Errorf("block (%s) has an invalid data hash", b.Hash(BlockHasher{}))
	}
	return nil
}

func (b *Block) Decode(dec Decoding[*Block]) error {
	return dec.Decode(b)
}

func (b *Block) Encode(enc Encoding[*Block]) error {
	return enc.Encode(b)
}

func CalculateDataHash(transactions []*Transaction) (hash types.Hash, err error) {
	buf := &bytes.Buffer{}
	for _, transaction := range transactions {
		if err := transaction.Encode(NewGobTxEncoder(buf)); err != nil {
			return types.Hash{}, err
		}
	}
	hash = sha256.Sum256(buf.Bytes())
	return
}
