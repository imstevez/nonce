package nonce

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

var (
	manage     Manager
	manageOnce sync.Once
)

func InitWithLocalManage(cr ChainStateReader, or OffChainStateReader, options ...ManagerOption) {
	manageOnce.Do(func() {
		manage = NewLocalManage(cr, or, options...)
	})
	return
}

func StartManage(ctx context.Context) { manage.Start(ctx) }

func WaitManage() { manage.Wait() }

func Returns(address common.Address, nonce uint64) { manage.Returns(address, nonce) }

func Assigns(address common.Address) (nonce uint64, err error) { return manage.Assigns(address) }
