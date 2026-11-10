package pool

import (
	"sync"
	"sync/atomic"
)

type Job struct {
	run        Task
	pending    int32
	successors []*Job
}

type Scheduler struct {
	sub   submitter //Куда отправлять готовые задачи (пул или выполнить сразу syncSubmitter)
	done  sync.WaitGroup
	roots []*Job
}

type submitter interface {
	Submit(Task)
}

type syncSubmitter struct{}

func (syncSubmitter) Submit(t Task) { t() }

func NewScheduler(p *WorkerPool) *Scheduler {
	return &Scheduler{sub: p}
}

func NewSyncScheduler() *Scheduler {
	return &Scheduler{sub: syncSubmitter{}}
}

func (s *Scheduler) NewJob(run Task, deps int32) *Job {
	j := &Job{run: run, pending: deps}
	s.done.Add(1)
	if deps == 0 {
		s.roots = append(s.roots, j)
	}
	return j
}

func (s *Scheduler) Start() {
	for _, j := range s.roots {
		s.sub.Submit(s.makeTask(j))
	}
	s.roots = nil
}

func AddDependency(prev, succ *Job) {
	prev.successors = append(prev.successors, succ)
}

func (s *Scheduler) Wait() {
	s.done.Wait()
}

func (s *Scheduler) makeTask(j *Job) Task {
	return func() {
		defer s.done.Done()
		j.run()
		for _, succ := range j.successors {
			if atomic.AddInt32(&succ.pending, -1) == 0 {
				s.sub.Submit(s.makeTask(succ))
			}
		}
	}
}
