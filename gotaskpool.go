package gotaskpool

import (
	"sync"
	"time"
)

type GoTask interface {
	Run()
}

type LimitedTaskQueue struct {
	queue      []GoTask
	sem        chan int
	size       int
	producePos int
	consumePos int
	sync.Mutex
}

func NewLimitedTaskQueue(limit int) *LimitedTaskQueue {
	ret := new(LimitedTaskQueue)
	ret.size = 0
	ret.queue = make([]GoTask, limit)
	ret.sem = make(chan int, limit)
	ret.producePos = -1
	ret.consumePos = -1
	return ret
}

func (m *LimitedTaskQueue) GetSize() int {
	return m.size
}

func (m *LimitedTaskQueue) Push(task GoTask, waitDuration time.Duration) bool {
	if task == nil {
		return false
	}
	if waitDuration > 0 {
		select {
		case <-time.After(waitDuration):
			return false
		case m.sem <- 1:
		}
	} else if waitDuration == 0 {
		select {
		case m.sem <- 1:
		default:
			return false
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
	if waitDuration > 0 {
		select {
		case <-time.After(waitDuration):
			return nil
		case <-m.sem:
		}
	} else if waitDuration == 0 {
		select {
		case <- m.sem:
		default:
			return nil
		}
	} else {
		<-m.sem
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

type TaskPool struct {
	taskQueue      *LimitedTaskQueue
	goroutineCount int
	stopped        bool
	started        bool
}

func NewTaskPool(poolSize int, workerCount int) *TaskPool {
	if poolSize <= 0 || workerCount <= 0 {
		return nil
	}
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
		m.stopped = false
	} else {
		m.started = true
		m.stopped = false
		for i := 0; i < m.goroutineCount; i++ {
			go func() {
				for {
					task := m.PullTask(-1)
					task.Run()
				}
			}()
		}
	}
}

func (m *TaskPool) Stop() {
	m.stopped = true
}

type PriorTaskPool struct {
	sync.Mutex
	taskQueues     []*LimitedTaskQueue
	taskSem        chan int
	goroutineCount int
	stopped        bool
	started        bool
}

func NewPriorTaskPool(priorityNum int, queueLen int, workerCount int) *PriorTaskPool {
	if priorityNum <= 0 || queueLen <= 0 || workerCount <= 0 {
		return nil
	}
	ret := new(PriorTaskPool)
	ret.taskSem = make(chan int, priorityNum*queueLen)
	ret.goroutineCount = workerCount
	ret.stopped = false
	ret.started = false
	ret.taskQueues = make([]*LimitedTaskQueue, 0, priorityNum)
	for i := 0; i < priorityNum; i++ {
		queue := NewLimitedTaskQueue(queueLen)
		ret.taskQueues = append(ret.taskQueues, queue)
	}
	return ret
}

func (m *PriorTaskPool) PushTask(priority int, task GoTask, waitDuration time.Duration) bool {
	if m.stopped || priority >= len(m.taskQueues) || priority < 0 {
		return false
	}
	ret := m.taskQueues[priority].Push(task, waitDuration)
	if ret {
		m.taskSem <- 1
	}

	return ret
}

func (m *PriorTaskPool) PullTask(waitDuration time.Duration) GoTask {
	tm := waitDuration
	if tm <= 0 {
		tm = time.Hour * 30
	}
	select {
	case <-m.taskSem:
		m.Lock()
		defer m.Unlock()
		for i := 0; i < len(m.taskQueues); i++ {
			if m.taskQueues[i].GetSize() > 0 {
				return m.taskQueues[i].Pull(tm)
			}
		}
	case <-time.After(tm):
		return nil
	}
	return nil
}

func (m *PriorTaskPool) Start() {
	if m.started {
		m.stopped = false
	} else {
		m.started = true
		m.stopped = false
		for i := 0; i < m.goroutineCount; i++ {
			go func() {
				for {
					task := m.PullTask(-1)
					task.Run()
				}
			}()
		}
	}
}

func (m *PriorTaskPool) Stop() {
	m.stopped = true
}
