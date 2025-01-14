package core

import (
	"crypto/sha256"
	"projectx/types"
)

// 任何实现其中hash方法的类型，则该类型自动实现了该接口
type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct{}

func (BlockHasher) Hash(b *Header) types.Hash {
	//计算hash值
	return types.Hash(sha256.Sum256(b.Bytes()))
}

type TxHasher struct{}

func (TxHasher) Hash(tx *Transaction) types.Hash {
	return types.Hash(sha256.Sum256(tx.Data))
}
