package pool

import "sync"

type Task func()

type WorkerPool struct {
	tasks   chan Task
	workers sync.WaitGroup
}

func New(n int) *WorkerPool {
	p := &WorkerPool{tasks: make(chan Task, n*2)}
	p.workers.Add(n)
	for i := 0; i < n; i++ {
		go p.worker()
	}
	return p
}

func (p *WorkerPool) worker() {
	defer p.workers.Done()
	for t := range p.tasks {
		t()
	}
}

func (p *WorkerPool) Submit(t Task) {
	p.tasks <- t
}

// Stop закрывает очередь и ждёт, пока воркеры доедят оставшиеся задачи.
// Требует: все Submit уже выполнены (гарантирует sched.Wait()).
func (p *WorkerPool) Stop() {
	close(p.tasks)
	p.workers.Wait()
}
