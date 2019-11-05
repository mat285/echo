package taskrunner

import "time"

// Action is the body of a build step or listener.
type Action func(tr *TaskRunner) error

// ErrorAction is the body of a build step or listener.
type ErrorAction func(tr *TaskRunner, err error) error

// Guard determines if the step should run.
type Guard func(tr *TaskRunner) (bool, error)

// Step is a pair of Action and Guard.
type Step struct {
	Action
	Guard
	Name string
}

// Module is a section of the build process that is fully encapsulated.
type Module interface {
	Register(tr *TaskRunner)
	Name() string
}

// TimingMark is a mark for stats collection.
type TimingMark struct {
	Name      string
	Timestamp time.Time
}
