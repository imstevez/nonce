package nonce

import (
	"container/heap"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"sort"
	"sync/atomic"
)

var _ Manager = (*LocalManage)(nil)

type LocalManage struct {
	returnsch chan returnsArg
	assignsch chan assignsArg
	unsaved   map[common.Address]*List[uint64]
	cr        ChainStateReader
	or        OffChainStateReader
	started   *atomic.Bool
	done      chan struct{}
}

type returnsArg struct {
	address common.Address
	nonce   uint64
}

type assignsArg struct {
	address common.Address
	nonce   uint64
	rstch   chan assignRst
}

type assignRst struct {
	nonce uint64
	err   error
}

func NewLocalManage(cr ChainStateReader, or OffChainStateReader, options ...ManagerOption) (manage *LocalManage) {
	manage = &LocalManage{
		returnsch: make(chan returnsArg, 1),
		assignsch: make(chan assignsArg, 1),
		unsaved:   make(map[common.Address]*List[uint64]),
		cr:        cr,
		or:        or,
		started:   &atomic.Bool{},
		done:      make(chan struct{}, 1),
	}
	for _, opt := range options {
		opt(manage)
	}
	return
}

func (l *LocalManage) Start(ctx context.Context) {
	if !l.started.CompareAndSwap(false, true) {
		panic("nonce manager start multi times")
	}
	go l.run(ctx)
	return
}

func (l *LocalManage) Wait() {
	l.mustStarted()
	<-l.done
}

func (l *LocalManage) Returns(address common.Address, nonce uint64) {
	l.mustStarted()
	l.returnsch <- returnsArg{address: address, nonce: nonce}
}

func (l *LocalManage) Assigns(address common.Address) (nonce uint64, err error) {
	l.mustStarted()
	rstch := make(chan assignRst)
	l.assignsch <- assignsArg{address: address, rstch: rstch}
	rst := <-rstch
	nonce, err = rst.nonce, rst.err
	return
}

func (l *LocalManage) run(ctx context.Context) {
	for {
		select {
		case arg := <-l.returnsch:
			l.returns(ctx, arg)
		case arg := <-l.assignsch:
			l.assigns(ctx, arg)
		case <-ctx.Done():
			close(l.done)
			break
		}
	}
}

func (l *LocalManage) mustStarted() {
	if !l.started.Load() {
		panic("call nonce manager before it started")
	}
}

func (l *LocalManage) returns(_ context.Context, arg returnsArg) {
	heap.Push(l.unsaved[arg.address], arg.nonce)
}

func (l *LocalManage) assigns(ctx context.Context, arg assignsArg) {
	if _, ok := l.unsaved[arg.address]; !ok {
		err := l.loadUnsaved(ctx, arg.address)
		if err != nil {
			arg.rstch <- assignRst{0, err}
			return
		}
	}
	unsaved := l.unsaved[arg.address]
	nonce := heap.Pop(unsaved).(uint64)
	if unsaved.Len() < 1 {
		unsaved.Push(nonce + 1)
	}
	arg.rstch <- assignRst{nonce: nonce, err: nil}
	return
}

func (l *LocalManage) loadUnsaved(ctx context.Context, addr common.Address) (err error) {
	next, err := l.cr.NonceAt(ctx, addr, nil)
	if err != nil {
		return
	}
	after, err := l.or.NonceAfter(ctx, addr, next-1)
	if err != nil {
		return
	}
	l.unsaved[addr] = new(List[uint64])
	if len(after) < 1 {
		heap.Push(l.unsaved[addr], next)
		return
	}
	sort.Slice(after, func(i, j int) bool {
		return after[i] < after[j]
	})
	for prev, i := next-1, 0; i < len(after); prev, i = after[i], i+1 {
		for n := prev + 1; n < after[i]; n++ {
			heap.Push(l.unsaved[addr], n)
		}
	}
	heap.Push(l.unsaved[addr], after[len(after)-1]+1)
	return
}
