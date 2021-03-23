package threadpool

import "sync"

type Pool interface {
	Add(f func())
	Wait()
}

func New(concurrentThreads, threadCount int) Pool {
	c := make(chan bool, concurrentThreads)
	for i := 0; i < concurrentThreads; i++ {
		c <- true
	}
	wg := sync.WaitGroup{}
	wg.Add(threadCount)

	return &pool{
		size: threadCount,
		c:    c,
		wg:   wg,
	}
}

type pool struct {
	size int
	c    chan bool
	wg   sync.WaitGroup
}

func (p *pool) Add(f func()) {
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

func (p *pool) Wait() {
	p.wg.Wait()
}
