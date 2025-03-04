package core

import (
	"fmt"
	"projectx/types"
	"sync"

	"github.com/go-kit/log"
)

type Blockchain struct {
	logger log.Logger
	// TODO:block and tx should use the different lock
	lock         sync.RWMutex
	headers      []*Header
	blocks       []*Block
	transactions []*Transaction
	store        Storage
	validator    Validator
	blockStore   map[types.Hash]*Block
	txStore      map[types.Hash]*Transaction
	// TODO: make this an interface
	constractState *State
}

// 第一个创世块
func NewBlockchain(l log.Logger, gensis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers:        []*Header{},
		store:          NewMemorStore(),
		blockStore:     make(map[types.Hash]*Block),
		txStore:        make(map[types.Hash]*Transaction),
		logger:         l,
		constractState: NewState(),
	}

	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockWithoutValidation(gensis)
	return bc, err
}

func (bc *Blockchain) SetVaildator(v Validator) {
	bc.validator = v
}

// execute contract when packea a new block
func (bc *Blockchain) AddBlock(block *Block) error {
	//validate block
	if err := bc.validator.Validate(block); err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(TxHasher{}))

		vm := NewVM(tx.Data, bc.constractState)

		if err := vm.Run(); err != nil {
			return err
		}
	}
	return bc.addBlockWithoutValidation(block)
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	if height > bc.Height() {
		return nil, fmt.Errorf("given heighth (%d) too height", height)
	}

	return bc.headers[height], nil
}

func (bc *Blockchain) GetTxByHash(hash types.Hash) (*Transaction, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	tx, ok := bc.txStore[hash]
	if !ok {
		return nil, fmt.Errorf("tx not found with hash (%s) ", hash)
	}
	return tx, nil
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	if height > bc.Height() {
		return nil, fmt.Errorf("given heighth (%d) too height", height)
	}

	return bc.blocks[height], nil
}

func (bc *Blockchain) GetBlockByHash(hash types.Hash) (*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	if block, ok := bc.blockStore[hash]; ok {
		return block, nil
	}
	return nil, fmt.Errorf("block with hash %s not found", hash)
}

// 判断是否有某个height的区块
func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

// [0,1,2,3] => len = 4
// [0,1,2,3] => height = 3
func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	bc.headers = append(bc.headers, b.Header)
	bc.blockStore[b.Hash(BlockHasher{})] = b
	for _, tx := range b.Transactions {
		bc.transactions = append(bc.transactions, tx)
		bc.txStore[tx.Hash(&TxHasher{})] = tx
	}
	bc.blocks = append(bc.blocks, b)

	bc.logger.Log(
		"msg", "new block",
		"hash", b.Hash(BlockHasher{}),
		"height", bc.Height(),
		"transactions", len(b.Transactions),
	)
	return bc.store.Put(b)
}
