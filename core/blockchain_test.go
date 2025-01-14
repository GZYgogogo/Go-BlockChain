package core

import (
	"fmt"
	"os"
	"projectx/types"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	l := log.NewLogfmtLogger(os.Stderr)
	bc, err := NewBlockchain(l, randomBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)
	return bc
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	// assert.Nil(t, err)
	if err != nil {
		return types.Hash{}
	}
	return BlockHasher{}.Hash(prevHeader)
}

func TestAddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenBlocks := 100
	for i := 0; i < lenBlocks; i++ {
		block := randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
	}
	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)
	assert.NotNil(t, bc.AddBlock(randomBlock(t, 89, types.Hash{})))
}

func TestGetHeader(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenBlocks := 100
	for i := 0; i < lenBlocks; i++ {
		block := randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
		header, err := bc.GetHeader(block.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}
}

func TestBlockchain(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
	fmt.Println("bc=", bc, "bc.Height=", bc.Height())
}

func TestHashBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	assert.True(t, bc.HasBlock(0))
	assert.False(t, bc.HasBlock(1))
	assert.False(t, bc.HasBlock(100))
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestAddBlockToHeigh(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.Nil(t, bc.AddBlock(randomBlock(t, 1, getPrevBlockHash(t, bc, 1))))
	assert.NotNil(t, bc.AddBlock(randomBlock(t, 3, types.Hash{})))
}
