package core

import (
	"encoding/gob"
	"io"
)

type Encoding[T any] interface {
	Encode(T) error
}

type Decoding[T any] interface {
	Decode(T) error
}

type GobTxEncoder struct {
	w io.Writer
}

func NewGobTxEncoder(w io.Writer) *GobTxEncoder {
	//无法注册，elliptic.P256()返回的结构体字段是私有的
	// gob.Register(elliptic.P256())
	return &GobTxEncoder{w: w}
}

func (en *GobTxEncoder) Encode(tx *Transaction) error {
	return gob.NewEncoder(en.w).Encode(tx)
}

type GobTxDecoder struct {
	r io.Reader
}

func NewGobTxDecoder(r io.Reader) *GobTxDecoder {
	// gob.Register(elliptic.P256())
	return &GobTxDecoder{r: r}
}

func (de *GobTxDecoder) Decode(tx *Transaction) error {
	return gob.NewDecoder(de.r).Decode(tx)
}

type GobBlockEncoder struct {
	w io.Writer
}

func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	return &GobBlockEncoder{w: w}
}

func (en *GobBlockEncoder) Encode(block *Block) error {
	return gob.NewEncoder(en.w).Encode(block)
}

type GobBlockDecoder struct {
	r io.Reader
}

func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	return &GobBlockDecoder{r: r}
}

func (de *GobBlockDecoder) Decode(block *Block) error {
	return gob.NewDecoder(de.r).Decode(block)
}
