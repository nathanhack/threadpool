package threadpool

import (
	"runtime"
	"sync"
)

type Pool interface {
	Add(f func())
	AddNoWait(f func())
	Wait()
}

//New creates a thread pool with concurrentThreads and totalJobs.
//  Once totalJobs have been added the pool is considered
//  full/finished and no more job will be allowed for this instance.
// if concurrentThreads is <=0 it will assume runtime.NumCPU().
func New(concurrentThreads, totalJobs int) Pool {
	if concurrentThreads <= 0 {
		concurrentThreads = runtime.NumCPU()
	}

	c := make(chan bool, concurrentThreads)
	for i := 0; i < concurrentThreads; i++ {
		c <- true
	}
	wg := sync.WaitGroup{}
	wg.Add(totalJobs)

	return &pool{
		size: totalJobs,
		c:    c,
		wg:   wg,
	}
}

type pool struct {
	size int
	c    chan bool
	wg   sync.WaitGroup
}

//Add adds a new job to be ran. When called it will blocks until a free thread can work on the job.
func (p *pool) Add(f func()) {
	if p.size == 0 {
		return
	}

	p.size--
	<-p.c
	go func() {
		f()
		p.c <- true
		p.wg.Done()
	}()
}

//AddNoWait adds a new job to be ran. When called it will not block until a free thread is created.
//  Instead it will spawn a goroutine that will wait until a free thread is available.
func (p *pool) AddNoWait(f func()) {
	if p.size == 0 {
		return
	}

	p.size--
	go func() {
		<-p.c
		f()
		p.c <- true
		p.wg.Done()
	}()
}

//Wait when called will block until all threads are completed. Note the pool will not be
// finished until all jobs have been queued and finished.
func (p *pool) Wait() {
	p.wg.Wait()
}
