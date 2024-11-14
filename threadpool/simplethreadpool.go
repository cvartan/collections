package simplethreadpool

import (
	"context"
	"errors"

	"github.com/cvartan/collections/queue"
)

type ProcessFunc func(ctx context.Context)

type processTask struct {
	process ProcessFunc
	cancel  context.CancelFunc
}

type ThreadPool struct {
	activeTasks  []*processTask
	waitingTasks *queue.Queue
	isCancelling bool
}

func NewThreadPool(maxThreads int) *ThreadPool {
	return &ThreadPool{
		activeTasks:  make([]*processTask, maxThreads),
		waitingTasks: queue.New(),
		isCancelling: false,
	}
}

func (t *ThreadPool) Put(process ProcessFunc) error {
	if t.isCancelling {
		return errors.New("pool is cancelling")
	}
	task := &processTask{
		process: process,
	}
	for i, item := range t.activeTasks {
		if item == nil {

			t.addActiveTask(i, task)
			return nil
		}
	}
	t.waitingTasks.Put(task)
	return nil
}

func (t *ThreadPool) addActiveTask(index int, task *processTask) {
	if t.isCancelling {
		return
	}
	if len(t.activeTasks) == index {
		t.activeTasks = append(t.activeTasks, task)
	} else {
		t.activeTasks[index] = task
	}
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		task.cancel = cancel
		task.process(ctx)
		func() {
			t.activeTasks[index] = nil
			if task := t.waitingTasks.Read(); task != nil {
				t.addActiveTask(index, task.(*processTask))
			}
		}()
	}()
}

func (t *ThreadPool) Wait() {
	for func() bool {
		for _, item := range t.activeTasks {
			if item != nil {
				return true
			}
		}
		return t.waitingTasks.Len() > 0
	}() {
	}
}

func (t *ThreadPool) CancelAll() {
	t.isCancelling = true
	t.waitingTasks.Clear()
	for i, task := range t.activeTasks {
		if task != nil {
			task.cancel()
			t.activeTasks[i] = nil
		}
	}
}
