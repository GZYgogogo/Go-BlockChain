package api

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"projectx/core"
	"projectx/types"
	"strconv"

	"github.com/go-kit/log"

	"github.com/labstack/echo/v4"
)

type APIError struct {
	Error string
}

type TxResponse struct {
	TxCount uint
	Hasher  []string
}

type Block struct {
	Version       uint32
	DataHash      string
	PrevBlockHash string
	Heihgt        uint32
	Signature     string
	Timestamp     int64
	TxResponse    TxResponse
}

type ServerConfig struct {
	Logger     log.Logger
	ListenAddr string
}

type Server struct {
	ServerConfig
	bc *core.Blockchain
}

func NewServer(cfg ServerConfig, blockchain *core.Blockchain) *Server {
	return &Server{
		ServerConfig: cfg,
		bc:           blockchain,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.GET("/block/:hashorid", s.handleGetBlock)
	e.GET("/tx/:hash", s.handleGetTx)

	return e.Start(s.ListenAddr)
}

func (s *Server) handleGetTx(c echo.Context) error {
	hash := c.Param("hash")
	txHashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
	}
	txHash := types.HashFromBytes(txHashBytes)
	tx, err := s.bc.GetTxByHash(txHash)
	if err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, tx)
}

func (s *Server) handleGetBlock(c echo.Context) error {
	hashorid := c.Param("hashorid")
	blockHeight, err := strconv.Atoi(hashorid)
	if err == nil {
		fmt.Println("blockHeight:", uint32(blockHeight))
		block, err := s.bc.GetBlock(uint32(blockHeight))
		if err != nil {
			return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
		}
		jsonBlock := intoJSONBlock(block)
		return c.JSON(http.StatusOK, jsonBlock)
	}

	// otherwise assume it's a hash
	hashBytes, err := hex.DecodeString(hashorid)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid hash"})
	}
	blockHash := types.HashFromBytes(hashBytes)
	block, err := s.bc.GetBlockByHash(blockHash)
	if err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
	}
	jsonBlock := intoJSONBlock(block)
	return c.JSON(http.StatusOK, jsonBlock)

	// return c.JSON(http.StatusOK, map[string]any{"msg": "Hello, World!"})
}

func intoJSONBlock(block *core.Block) *Block {
	TxResponse := TxResponse{
		TxCount: uint(len(block.Transactions)),
		Hasher:  make([]string, len(block.Transactions)),
	}

	for i, tx := range block.Transactions {
		TxResponse.Hasher[i] = hex.EncodeToString(tx.Hash(core.TxHasher{}).ToSlice())
	}

	return &Block{
		Version:       block.Header.Version,
		DataHash:      block.Header.DataHash.String(),
		PrevBlockHash: block.Header.PrevBlockHash.String(),
		Heihgt:        block.Header.Height,
		Signature:     block.Signature.String(),
		Timestamp:     block.Header.Timestamp,
		TxResponse:    TxResponse,
	}
}
