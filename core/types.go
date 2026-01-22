// File: core/types.go
// Version: 1.0.7
package core

import (
	"database/sql"
	"fmt"
	"math/big"
)

type Config struct {
	Port         int
	WorkInterval int
	ChainRPC     string
}

type RPCClient struct {
	Endpoint string
}

func NewRPCClient(endpoint string) *RPCClient {
	return &RPCClient{Endpoint: endpoint}
}

var (
	cfg       Config
	rpcClient *RPCClient
	db        *sql.DB
)

func SetConfig(c Config)        { cfg = c }
func SetRPCClient(c *RPCClient) { rpcClient = c }
func SetDB(d *sql.DB)           { db = d }

// DifficultyFromNBits calculates mining difficulty from compact nBits representation.
func DifficultyFromNBits(nBits string) float64 {
	if len(nBits) != 8 {
		return 0
	}

	exp := parseHexByte(nBits[0:2])
	mantissa := new(big.Int)
	mantissa.SetString(nBits[2:], 16)

	diff1 := new(big.Int)
	diff1.SetString("00000000FFFF0000000000000000000000000000000000000000000000000000", 16)
	target := new(big.Int).Lsh(mantissa, uint(8*(exp-3)))
	if target.Sign() == 0 {
		return 0
	}

	ratio := new(big.Rat).SetFrac(diff1, target)
	f, _ := ratio.Float64()
	return f
}

// HumanDiff returns a human-readable difficulty string.
func HumanDiff(d float64) string {
	suffixes := []string{"", "K", "M", "G", "T", "P"}
	idx := 0
	for d >= 1000 && idx < len(suffixes)-1 {
		d /= 1000
		idx++
	}
	return fmt.Sprintf("%.2f%s", d, suffixes[idx])
}

func parseHexByte(s string) int {
	var n int
	fmt.Sscanf(s, "%x", &n)
	return n
}