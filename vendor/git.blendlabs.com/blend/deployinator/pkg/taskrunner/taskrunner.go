package taskrunner

import (
	"fmt"
	"time"

	exception "github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

// NewLogger returns a new logger.
func NewLogger() *logger.Logger {
	agent := logger.All()
	agent.WithHeading("task-runner")
	return agent
}

// New creates a TaskRunner from the os environment.
func New(name string) *TaskRunner {
	return &TaskRunner{
		name: name,
		log:  NewLogger(),
	}
}

// TaskRunner is a basic structure to execute tasks
type TaskRunner struct {
	name              string
	log               *logger.Logger
	modules           []Module
	steps             []Step
	startListeners    []Action
	completeListeners []Action
	failureListeners  []ErrorAction
	config            interface{}
	timings           []TimingMark
}

// Name returns the name
func (tr *TaskRunner) Name() string {
	return tr.name
}

// Logger returns the logger
func (tr *TaskRunner) Logger() *logger.Logger {
	return tr.log
}

// AddStep adds a step to the task queue
func (tr *TaskRunner) AddStep(action Action, guard Guard, name string) {
	tr.steps = append(tr.steps, Step{
		Name:   name,
		Guard:  guard,
		Action: action,
	})
}

// AddNonFatalStep adds a non-fatal step to the task queue. Guard failure can still terminate the task.
func (tr *TaskRunner) AddNonFatalStep(action Action, guard Guard, name string) {
	tr.AddStep(NonFatal(action), guard, name)
}

// AddStartedListener adds a started listener
func (tr *TaskRunner) AddStartedListener(action Action) {
	tr.startListeners = append(tr.startListeners, action)
}

// AddCompleteListener adds a complete listener
func (tr *TaskRunner) AddCompleteListener(action Action) {
	tr.completeListeners = append(tr.completeListeners, action)
}

// AddFailureListener adds a failure listener
func (tr *TaskRunner) AddFailureListener(failureAction ErrorAction) {
	tr.failureListeners = append(tr.failureListeners, failureAction)
}

// Timings return timing marks for the TaskRunner
func (tr *TaskRunner) Timings() []TimingMark {
	return tr.timings
}

// Register registers a given set of modules.
func (tr *TaskRunner) Register(modules ...Module) {
	for _, module := range modules {
		tr.modules = append(tr.modules, module)
		module.Register(tr)
	}
}

// Modules returns the registered modules for the taskrunner
func (tr *TaskRunner) Modules() []Module {
	return tr.modules
}

// Steps retrusn the current steps list
func (tr *TaskRunner) Steps() []Step {
	return tr.steps
}

// Mark marks a timing event
func (tr *TaskRunner) Mark(event string) {
	tr.timings = append(tr.timings, TimingMark{Name: event, Timestamp: time.Now()})
}

// MarkStarted marks the run as started
func (tr *TaskRunner) MarkStarted() {
	tr.Mark("start")
	tr.log.Infof("`%s` Run Starting", tr.name)

}

// MarkComplete marks the run as complete
func (tr *TaskRunner) MarkComplete() {
	tr.Mark("complete")
	tr.log.Infof("Run Complete")
}

// NotifyStarted fires start listeners.
func (tr *TaskRunner) NotifyStarted() {
	var err error
	for _, action := range tr.startListeners {
		err = action(tr)
		if err != nil {
			tr.log.Error(err)
		}
	}
}

// NotifyComplete notifies complete listeners on completion
func (tr *TaskRunner) NotifyComplete() {
	var err error
	for _, action := range tr.completeListeners {
		err = action(tr)
		if err != nil {
			tr.log.Error(err)
		}
	}
}

// NotifyFailed notifies complete listeners on failure
func (tr *TaskRunner) NotifyFailed(stepErr error) {
	var err error
	for _, action := range tr.failureListeners {
		err = action(tr, stepErr)
		if err != nil {
			tr.log.Error(err)
		}
	}
}

// RunSteps runs the steps in order
func (tr *TaskRunner) RunSteps() error {
	var err error
	for _, step := range tr.steps {
		shouldRunStep := false
		if step.Guard == nil {
			shouldRunStep = true
		} else {
			shouldRunStep, err = step.Guard(tr)
			if err != nil {
				return err
			}
		}
		if shouldRunStep {
			if len(step.Name) > 0 {
				tr.Infof("TaskRunner step `%s`", step.Name)
			}
			err = step.Action(tr)
			if err != nil {
				return err
			}
			if len(step.Name) > 0 {
				tr.Infof("TaskRunner step `%s` complete", step.Name)
			}
		} else {
			tr.Infof("Taskrunner step %s skipped", step.Name)
		}
	}
	return nil
}

// SetConfig sets the config
func (tr *TaskRunner) SetConfig(config interface{}) {
	tr.config = config
}

// GetConfig sets the config
func (tr *TaskRunner) GetConfig() interface{} {
	return tr.config
}

// Fail fails the job and reports the failure.
func (tr *TaskRunner) Fail(err error) {
	tr.NotifyFailed(err)
	tr.log.SyncFatalExit(err)
}

// Failf fails the job and reports the failure, creating an error
// with the given format and arguments.
func (tr *TaskRunner) Failf(format string, args ...interface{}) {
	tr.Fail(exception.New(fmt.Sprintf(format, args...)))
}

// PrintTimings prints a breakdown of how long each phase took to the logger.
func (tr *TaskRunner) PrintTimings() {
	tr.Debugf("Task Stats:")
	for markIndex := 0; markIndex < len(tr.timings)-2; markIndex++ {
		m0 := tr.timings[markIndex]
		m1 := tr.timings[(markIndex+1)%len(tr.timings)]
		tr.Debugf("%s %v", m1.Name, m1.Timestamp.Sub(m0.Timestamp))
	}
	tr.Debugf("%s %v",
		"overall",
		tr.timings[len(tr.timings)-1].Timestamp.Sub(tr.timings[0].Timestamp),
	)
}

// Run runs the configured steps from start to end, including notifications and timing analysis
func (tr *TaskRunner) Run() {
	tr.MarkStarted()
	tr.NotifyStarted()

	tr.Infof("Log Flags: %s", tr.log.Flags().String())

	err := tr.RunSteps()
	if err != nil {
		tr.Fail(err)
	}

	tr.MarkComplete()
	tr.PrintTimings()
	tr.NotifyComplete()
}

// Infof is a stub for the logger.
func (tr *TaskRunner) Infof(format string, args ...interface{}) {
	if tr.log != nil {
		tr.log.SyncInfof(format, args...)
	}
}

// Debugf is a stub for the logger.
func (tr *TaskRunner) Debugf(format string, args ...interface{}) {
	if tr.log != nil {
		tr.log.SyncDebugf(format, args...)
	}
}

// Warningf is a stub for the logger.
func (tr *TaskRunner) Warningf(format string, args ...interface{}) {
	if tr.log != nil {
		tr.log.SyncWarningf(format, args...)
	}
}

// Errorf is a stub for the logger.
func (tr *TaskRunner) Errorf(format string, args ...interface{}) {
	if tr.log != nil {
		tr.log.SyncErrorf(format, args...)
	}
}

// Warning is a stub for the logger.
func (tr *TaskRunner) Warning(err error) {
	if tr.log != nil {
		tr.log.SyncWarning(err)
	}
}

// Error is a stub for the logger.
func (tr *TaskRunner) Error(err error) {
	if tr.log != nil {
		tr.log.SyncError(err)
	}
}

// FatalExit is a stub for the logger.
func (tr *TaskRunner) FatalExit(err error) {
	if tr.log != nil {
		tr.log.SyncFatalExit(err)
	}
}

// NonFatal is an action that logs an error instead of failing the task
func NonFatal(action Action) Action {
	return func(tr *TaskRunner) error {
		if err := action(tr); err != nil {
			tr.Errorf("non-fatal error: %v", err)
		}
		return nil
	}
}
