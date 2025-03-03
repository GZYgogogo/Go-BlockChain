package core

import (
	"fmt"
	"projectx/crypto"
	"projectx/types"
)

// 通用的交易结构，而非加密的
type Transaction struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature
	//cached version of the trancation data hash
	hash types.Hash
	// firstSceen is the timestamp of when the trancation is first seen locally
	firstSceen int64
}

func (tx *Transaction) SetFirstSceen(t int64) {
	tx.firstSceen = t
}

func (tx *Transaction) FirstSceen() int64 {
	return tx.firstSceen
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

// 私钥签名交易
func (tx *Transaction) Sign(privKey crypto.PrivateKey) error {
	sig, err := privKey.Sign(tx.Data)
	if err != nil {
		return err
	}
	tx.From = privKey.PublicKey()
	tx.Signature = sig
	return nil
}

// 验证交易是否合法
func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction has no signature")
	}
	if !tx.Signature.Verify(tx.From, tx.Data) {
		return fmt.Errorf("invalid transaction signature")
	}
	return nil
}

func (tx *Transaction) Decode(dec Decoding[*Transaction]) error {
	return dec.Decode(tx)
}

func (tx *Transaction) Encode(enc Encoding[*Transaction]) error {
	return enc.Encode(tx)
}
