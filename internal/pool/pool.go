package pool

import "sync"

type Task func()

type WorkerPool struct {
	tasks   chan Task
	pending sync.WaitGroup // задачи: Add перед Submit, Done после выполнения
	workers sync.WaitGroup // воркеры: Add(N) при старте, Done при выходе
}

// New создаёт пул из n воркеров и запускает их.
func New(n int) *WorkerPool {
	p := &WorkerPool{
		tasks: make(chan Task),
	}
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
		p.pending.Done()
	}
}

// Submit кладёт задачу в пул. Блокируется, если все воркеры заняты.
func (p *WorkerPool) Submit(t Task) {
	p.pending.Add(1)
	p.tasks <- t
}

// Wait блокируется, пока все отправленные задачи не завершатся.
func (p *WorkerPool) Wait() {
	p.pending.Wait()
}

// Stop закрывает канал задач и дожидается выхода всех воркеров.
// Вызывать строго после Wait().
func (p *WorkerPool) Stop() {
	close(p.tasks)
	p.workers.Wait()
}
