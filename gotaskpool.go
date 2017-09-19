package gotaskpool

import (
	"sync"
	"time"
)

type GoTask interface {
	Run()
}

type LimitedTaskQueue struct {
	queue []GoTask
	sem chan int
	size int32
	producePos int
	consumePos int
	sync.Mutex
}

func NewLimitedTaskQueue(limit int) *LimitedTaskQueue {
	ret := new(LimitedTaskQueue)
	ret.size = 0
	ret.queue = make([]GoTask,limit)
	ret.sem = make(chan int, limit)
	ret.producePos = -1
	ret.consumePos = -1
	return ret
}

func (m *LimitedTaskQueue) Push(task GoTask, waitDuration time.Duration) bool {
	if task == nil  {
		return false
	}
	if waitDuration > 0 {
		select {
		case <- time.After(waitDuration):
			return false
		case m.sem <- 1:
		}
	} else {
		m.sem <- 1
	}

	m.Lock()
	defer m.Unlock()
	m.producePos++
	if m.producePos >= len(m.queue) {
		m.producePos = 0
	}
	m.queue[m.producePos] = task
	m.size++
	return true
}

func (m *LimitedTaskQueue) Pull(waitDuration time.Duration) GoTask {
	if waitDuration >= 0 {
		select {
		case <- time.After(waitDuration):
			return nil
		case <- m.sem:
		}
	} else {
		<- m.sem
	}
	m.Lock()
	defer m.Unlock()
	m.consumePos++
	if m.consumePos >= len(m.queue) {
		m.consumePos = 0
	}
	m.size--
	return m.queue[m.consumePos]
}

type TaskPool struct{
	taskQueue *LimitedTaskQueue
	goroutineCount int
	stopped bool
	started bool
}

func NewTaskPool(poolSize int, workerCount int) *TaskPool{
	ret := new(TaskPool)
	ret.taskQueue = NewLimitedTaskQueue(poolSize)
	ret.goroutineCount = workerCount
	if ret.goroutineCount <= 0 {
		ret.goroutineCount = 3
	}
	return ret
}

func (m *TaskPool) PushTask(task GoTask, waitDuration time.Duration) bool {
	if m.stopped {
		return false
	}
	return m.taskQueue.Push(task, waitDuration)
}

func (m *TaskPool) PullTask(waitDuration time.Duration) GoTask {
	return m.taskQueue.Pull(waitDuration)
}

func (m *TaskPool) Start() {
	if m.started {
		return
	}
	m.started = true
	for i := 0; i < m.goroutineCount; i++ {
		go func() {
			for{
				task := m.PullTask(-1)
				task.Run()
			}
		}()
	}
}

func (m *TaskPool) Stop() {
	m.stopped = true
}