package threadpool

import (
	"math"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestPool_Add(t *testing.T) {
	total := 16
	concur := 3

	h := New(concur, total)

	start := time.Now()
	for i := 0; i < total; i++ {
		h.Add(func() {
			time.Sleep(time.Second)
		})
	}
	h.Wait()

	actual := int(time.Since(start).Seconds())
	expectedTime := int(math.Ceil(float64(total) / float64(concur)))
	if actual > expectedTime {
		t.Fatalf("expected %v but found %v", expectedTime, actual)
	}
}

func TestPool_AddNoWait(t *testing.T) {
	total := 16
	concur := 3

	h := New(concur, total)

	start := time.Now()
	for i := 0; i < total; i++ {
		h.AddNoWait(func() {
			time.Sleep(time.Second)
		})
	}
	h.Wait()

	actual := int(time.Since(start).Seconds())
	expectedTime := int(math.Ceil(float64(total) / float64(concur)))
	if actual > expectedTime {
		t.Fatalf("expected %v but found %v", expectedTime, actual)
	}
}

func TestPool_AddNoWait2(t *testing.T) {
	total := runtime.NumCPU() * 3
	concur := -1

	h := New(concur, total)

	actual := 0
	mut := sync.Mutex{}

	for i := 0; i < total; i++ {
		h.AddNoWait(func() {
			mut.Lock()
			defer mut.Unlock()

			actual++
		})
	}
	h.Wait()

	if total != actual {
		t.Fatalf("expected %v but found %v", total, actual)
	}
}

func TestPool_MultiThreadAdd(t *testing.T) {
	threadAmount := 10
	threadCount := 10
	h := New(0, threadCount*threadAmount)

	for i := 0; i < threadCount; i++ {
		t.Logf("creating thread: %v", i)
		go func() {
			for j := 0; j < threadAmount; j++ {
				if j%2 == 0 {
					h.Add(func() { time.Sleep(100 * time.Millisecond) })
				} else {
					h.AddNoWait(func() { time.Sleep(100 * time.Millisecond) })
				}
			}
		}()
	}
	h.Wait()
}
