package threadpool

import (
	"context"
	"runtime"
	"sync"
)

type Pool interface {
	Add(f func())
	AddNoWait(f func())
	Wait()
	ForceFinish()
}

//New creates a thread pool with concurrentThreads and totalJobs.
//  Once totalJobs have been added the pool is considered
//  full/finished and no more job will be allowed for this instance.
// if concurrentThreads is <=0 it will assume runtime.NumCPU().
func New(ctx context.Context, concurrentThreads, totalJobs int) Pool {
	if concurrentThreads <= 0 {
		concurrentThreads = runtime.NumCPU()
	}

	c := make(chan bool, concurrentThreads)
	for i := 0; i < concurrentThreads; i++ {
		c <- true
	}

	cCtx, can := context.WithCancel(ctx)
	p := pool{
		size:      totalJobs,
		mux:       sync.Mutex{},
		ctx:       cCtx,
		ctxCancel: can,
		c:         c,
		wg:        sync.WaitGroup{},
	}
	p.wg.Add(totalJobs)

	return &p
}

type pool struct {
	size      int
	mux       sync.Mutex
	ctx       context.Context
	ctxCancel context.CancelFunc
	c         chan bool
	wg        sync.WaitGroup
}

//Add adds a new job to be ran. When called it will blocks until a free thread can work on the job.
func (p *pool) Add(f func()) {
	p.mux.Lock()
	if p.size == 0 {
		p.mux.Unlock()
		return
	}

	p.size--
	p.mux.Unlock()
	select {
	case <-p.c:
	case <-p.ctx.Done():
		// we zeroize the waitgroup
		p.zeroizeWaitgroup()
		return
	}
	go func() {
		f()
		p.c <- true
		p.wg.Done()
	}()
}

func (p *pool) zeroizeWaitgroup() {
	p.mux.Lock()
	if p.size > 0 {
		p.wg.Add(-(p.size + 1))
		p.size = 0
	}
	p.mux.Unlock()
}

//AddNoWait adds a new job to be ran. When called it will not block until a free thread is created.
//  Instead it will spawn a goroutine that will wait until a free thread is available.
func (p *pool) AddNoWait(f func()) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.size == 0 {
		return
	}

	p.size--

	go func() {
		defer p.wg.Done()
		select {
		case <-p.c:
		case <-p.ctx.Done():
			// we zeroize the waitgroup
			p.zeroizeWaitgroup()
			return
		}
		f()
		p.c <- true
	}()
}

//ForceFinish provides an easy method prevent any future Add() from executing and prevent
// any waiting goroutines from AddNoWait() from starting
func (p *pool) ForceFinish() {
	p.ctxCancel()
}

//Wait when called will block until all threads are completed. Note the pool will not be
// finished until all jobs have been queued and finished.
func (p *pool) Wait() {
	p.wg.Wait()
}
