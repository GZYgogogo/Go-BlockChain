package core

import (
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

type Blockchain struct {
	logger         log.Logger
	lock           sync.RWMutex
	headers        []*Header
	store          Storage
	validator      Validator
	contractStatus *Status
}

// 第一个创世块
func NewBlockchain(l log.Logger, gensis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers:        []*Header{},
		store:          NewMemorStore(),
		logger:         l,
		contractStatus: NewStatus(),
	}

	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockWithoutValidation(gensis)
	return bc, err
}

func (bc *Blockchain) SetVaildator(v Validator) {
	bc.validator = v
}

func (bc *Blockchain) AddBlock(block *Block) error {
	//validate block
	if err := bc.validator.Validate(block); err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}))

		vm := NewVM(tx.Data, bc.contractStatus)
		if err := vm.Run(); err != nil {
			return err
		}

		// bc.logger.Log("result", vm.stack.data[vm.stack.sp])
		fmt.Printf("%+v\n", vm.contractStatus)
	}
	return bc.addBlockWithoutValidation(block)
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given heighth (%d) too height", height)
	}
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return bc.headers[height], nil
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

	bc.logger.Log(
		"msg", "new block",
		"hash", b.Hash(BlockHasher{}),
		"height", bc.Height(),
		"transactions", len(b.Transactions),
	)
	return bc.store.Put(b)
}
