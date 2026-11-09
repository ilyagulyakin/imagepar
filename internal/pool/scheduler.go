package pool

import (
	"sync"
	"sync/atomic"
)

type Job struct {
	run        Task
	pending    int32
	successors []*Job
	owner      *Scheduler
}

type Scheduler struct {
	sub   submitter
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
	j := &Job{run: run, pending: deps, owner: s}
	s.done.Add(1)
	if deps == 0 {
		s.roots = append(s.roots, j)
	}
	return j
}

func (s *Scheduler) Start() {
	for _, j := range s.roots {
		s.sub.Submit(j.wrap())
	}
	s.roots = nil
}

func AddDependency(prev, succ *Job) {
	prev.successors = append(prev.successors, succ)
}

func (s *Scheduler) Wait() {
	s.done.Wait()
}

func (j *Job) wrap() Task {
	return func() {
		defer j.owner.done.Done()
		j.run()
		for _, succ := range j.successors {
			if atomic.AddInt32(&succ.pending, -1) == 0 {
				j.owner.sub.Submit(succ.wrap())
			}
		}
	}
}
