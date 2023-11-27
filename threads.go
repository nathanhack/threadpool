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

// NewFixedSize creates a thread pool with concurrentThreads and totalJobs.
//
//	Once totalJobs have been added the pool is considered
//	full/finished and no more job will be allowed for this instance.
//	Additionally, Wait() will wait until all totalJobs Add() or AddNoWait()
//	have completed. So make sure to Add()/AddNoWait() totalJobs or it will wait
//	forever.
//
// If concurrentThreads is <=0 it will assume runtime.NumCPU().
func NewFixedSize(ctx context.Context, concurrentThreads, totalJobs int) Pool {
	if concurrentThreads <= 0 {
		concurrentThreads = runtime.NumCPU()
	}

	c := make(chan bool, concurrentThreads)
	for i := 0; i < concurrentThreads; i++ {
		c <- true
	}

	cCtx, can := context.WithCancel(ctx)
	p := fixedPool{
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

type fixedPool struct {
	size      int
	mux       sync.Mutex
	ctx       context.Context
	ctxCancel context.CancelFunc
	c         chan bool
	wg        sync.WaitGroup
}

// Add adds a new job to be ran. When called it will blocks until a free thread can work on the job.
func (p *fixedPool) Add(f func()) {
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

func (p *fixedPool) zeroizeWaitgroup() {
	p.mux.Lock()
	if p.size > 0 {
		p.wg.Add(-(p.size + 1))
		p.size = 0
	}
	p.mux.Unlock()
}

// AddNoWait adds a new job to be ran. When called it will not block until a free thread is created.
//
//	Instead it will spawn a goroutine that will wait until a free thread is available.
func (p *fixedPool) AddNoWait(f func()) {
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

// ForceFinish provides an easy method prevent any future Add() from executing and prevent
// any waiting goroutines from AddNoWait() from starting
func (p *fixedPool) ForceFinish() {
	p.ctxCancel()
}

// Wait when called will block until all threads are completed. Note the pool will not be
// finished until all jobs have been queued and finished.
func (p *fixedPool) Wait() {
	p.wg.Wait()
}

// IsDone will return the status of the context if it is Done. If false it means
// additional Add*s() are still needed.
func (p *fixedPool) IsDone() bool {
	select {
	case <-p.ctx.Done():
		return true
	default:
		return false
	}
}

type dynamicPool struct {
	mux       sync.Mutex
	ctx       context.Context
	ctxCancel context.CancelFunc
	c         chan bool
	wg        sync.WaitGroup
}

// New creates a thread pool with concurrentThreads limiter.
//
//		This Pool most aligns with a WaitGroup's Add() and Wait(),
//	 with the additional layer of throttling running threads to a max
//	 of concurrentThreads concurrently running.
//
// If concurrentThreads is <=0 it will assume runtime.NumCPU().
func New(ctx context.Context, concurrentThreads int) Pool {
	if concurrentThreads <= 0 {
		concurrentThreads = runtime.NumCPU()
	}

	c := make(chan bool, concurrentThreads)
	for i := 0; i < concurrentThreads; i++ {
		c <- true
	}

	cCtx, can := context.WithCancel(ctx)
	p := dynamicPool{
		mux:       sync.Mutex{},
		ctx:       cCtx,
		ctxCancel: can,
		c:         c,
		wg:        sync.WaitGroup{},
	}

	return &p
}

// Add adds a new job to be ran. When called it will blocks until a free thread can work on the job.
func (p *dynamicPool) Add(f func()) {
	p.wg.Add(1)
	select {
	case <-p.ctx.Done():
		p.wg.Done()
		return
	case <-p.c:
	}
	go func() {
		f()
		p.c <- true
		p.wg.Done()
	}()
}

// AddNoWait adds a new job to be ran. When called it will not block until a free thread is created.
//
//	Instead it will spawn a goroutine that will wait until a free thread is available.
func (p *dynamicPool) AddNoWait(f func()) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		select {
		case <-p.ctx.Done():
			return
		case <-p.c:
		}
		f()
		p.c <- true
	}()
}

// ForceFinish provides an easy method prevent any future Add() from executing and prevent
// any waiting goroutines from AddNoWait() from starting
func (p *dynamicPool) ForceFinish() {
	p.ctxCancel()
}

// Wait when called will block until all threads are completed. Note the pool will not be
// finished until all jobs have been queued and finished.
func (p *dynamicPool) Wait() {
	p.wg.Wait()
}

// IsDone will return the status of the context if it is Done. A value true indicates the pool
func (p *dynamicPool) IsDone() bool {
	select {
	case <-p.ctx.Done():
		return true
	default:
		return false
	}
}
