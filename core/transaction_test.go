package core

import (
	"bytes"
	"fmt"
	"projectx/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{Data: []byte("foo")}

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)

	fmt.Println("transaction signature:", tx.Signature)
}

func TestVerifyTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{Data: []byte("foo")}
	assert.Nil(t, tx.Sign(privKey))
	assert.Nil(t, tx.Verify())
	otherPrivKey := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()
	assert.NotNil(t, tx.Verify())
}

func randomTxWithSignature(t *testing.T) *Transaction {
	privKey := crypto.GeneratePrivateKey()
	tx := Transaction{
		Data: []byte("foo"),
	}

	assert.Nil(t, tx.Sign(privKey))
	return &tx
}

func TestTxEndcodeDecode(t *testing.T) {
	tx := randomTxWithSignature(t)
	buf := &bytes.Buffer{}
	// assert.NotNil(t, NewGobTxEncoder(buf))
	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	txDecoded := new(Transaction)
	assert.Nil(t, txDecoded.Decode(NewGobTxDecoder(buf)))
	assert.Equal(t, tx, txDecoded)
}
