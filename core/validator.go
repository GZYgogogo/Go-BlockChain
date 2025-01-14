package core

import "fmt"

type Validator interface {
	Validate(*Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{
		bc: bc,
	}
}

func (v *BlockValidator) Validate(b *Block) error {
	if v.bc.HasBlock(b.Height) {
		return fmt.Errorf("chain already contains block (%d) with hash (%s)", b.Height, b.Hash(BlockHasher{}))
	}
	if v.bc.Height()+1 != b.Height {
		return fmt.Errorf("block (%s) with height (%d) is too height => current height (%d)", b.Hash(BlockHasher{}), b.Height, v.bc.Height())
	}
	prevHeader, err := v.bc.GetHeader(b.Height - 1)
	if err != nil {
		return err
	}
	prevHash := BlockHasher{}.Hash(prevHeader)
	if prevHash != b.PreBlockHash {
		return fmt.Errorf("the hash of previous block (%d) is invalid", b.Height)
	}
	if err := b.Verify(); err != nil {
		return err
	}
	return nil
}
