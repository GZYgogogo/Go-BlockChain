package core

import (
	"bytes"
	"projectx/crypto"
	"projectx/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature(t)
	hearder := &Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
	}
	b, err := NewBlock(hearder, []*Transaction{tx})
	assert.Nil(t, err)

	dataHash, err := CalculateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.DataHash = dataHash
	assert.Nil(t, b.Sign(privKey))

	return b
}

func TestSignBlock(t *testing.T) {
	priv := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})
	assert.Nil(t, b.Sign(priv))
	assert.NotNil(t, b.Signature)
}

func TestVerifyBlock(t *testing.T) {
	priv := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})
	assert.Nil(t, b.Sign(priv))
	assert.Nil(t, b.Verify())
}

func TestBlockEndcodeDecode(t *testing.T) {
	block := randomBlock(t, 1, types.Hash{})
	buf := &bytes.Buffer{}
	// assert.NotNil(t, NewGobTxEncoder(buf))
	assert.Nil(t, block.Encode(NewGobBlockEncoder(buf)))

	blockDecoded := new(Block)
	assert.Nil(t, blockDecoded.Decode(NewGobBlockDecoder(buf)))
	assert.Equal(t, block, blockDecoded)
}
