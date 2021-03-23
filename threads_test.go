package threadpool

import (
	"math"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
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
