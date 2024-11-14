package simplethreadpool_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	simplethreadpool "github.com/cvartan/collections/threadpool"
	"github.com/google/uuid"
)

func TestPool(t *testing.T) {
	pool := simplethreadpool.NewThreadPool(4)
	for i := 0; i < 10; i++ {
		process := func(ctx context.Context) {
			processId := uuid.NewString()
			t.Log("Start process", processId)
			dur := rand.Intn(8) + 1
			time.Sleep(time.Duration(dur) * time.Second)
			t.Log("End process", processId, " ", dur)
		}
		pool.Put(process)
	}
	pool.Wait()
}

func TestPoolWithCancel(t *testing.T) {
	pool := simplethreadpool.NewThreadPool(4)
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		process := func(ctx context.Context) {
			isActive := true
			info := make(chan time.Time)
			defer close(info)
			// Timer function
			wg.Add(1)
			go func() {
				defer wg.Done()
				for isActive {
					info <- time.Now()
					time.Sleep(time.Second)
				}
				t.Log("Ending timer ", i)
			}()
			for {
				select {
				case current := <-info:
					{
						t.Log("Process ", i, " time ", current)
					}
				case <-ctx.Done():
					{
						isActive = false
						time.Sleep(2 * time.Second)
					}
				}
			}
		}
		if err := pool.Put(process); err != nil {
			t.Fatal(err)
		}
	}
	// Terminate function
	go func() {
		dur := rand.Intn(15) + 4
		t.Log("Paused for", dur)
		time.Sleep(time.Duration(dur) * time.Second)
		pool.CancelAll()
		t.Log("Canceling")
		time.Sleep(5 * time.Second)
	}()
	pool.Wait()
	wg.Wait()
}
