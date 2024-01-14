package nonce

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Manager interface {
	Start(ctx context.Context)
	Wait()
	Returns(address common.Address, nonce uint64)
	Assigns(address common.Address) (nonce uint64, err error)
}

type ManagerOption func(Manager)

// ChainStateReader used to load on chain state, it should be a sub of ethereum.ChainStateReader
type ChainStateReader interface {
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
}

// OffChainStateReader used to load off chain state
type OffChainStateReader interface {
	// NonceAfter list all used nonce which are bigger than 'after' for the specified 'address'
	NonceAfter(ctx context.Context, address common.Address, after uint64) (list []uint64, err error)
}
