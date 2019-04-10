package flow

import (
	"fmt"
	"strings"

	"github.com/privatix/dappctrl/util/log"
)

// Step can be done and undone.
type Step interface {
	Name() string
	Do(interface{}) error
	Undo(interface{}) error
}

// Flow is a slice of Steps interface elements
// to be run in sequence.
type Flow struct {
	Name  string
	Steps []Step
}

// Run executes the flow elements runner function.
func (flow Flow) Run(logger log.Logger, flowContext interface{}) error {
	rollback := func(steps []Step) {
		for i := range steps {
			// selecting backward
			v := steps[len(steps)-i-1]
			err := v.Undo(flowContext)
			if err != nil {
				logger.Warn(fmt.Sprintf("failed to undo '%s': %v", v.Name(), err))
			}
		}
	}

	if flow.Name != "" {
		logger.Info(fmt.Sprintf("'%s' started", flow.Name))
	}

	var err error
	for i, v := range flow.Steps {
		err = v.Do(flowContext)

		if err != nil {
			logger.Warn(fmt.Sprintf("failed to execute '%s': %v", v.Name(), err))
			rollback(flow.Steps[:i+1])
			break
		}
		logger.Info(fmt.Sprintf("'%v' is done", v.Name()))
	}
	return err
}

// Execute chooses and runs a flow based on os.Args[1].
func Execute(logger log.Logger, arg string, flows map[string]Flow, flowContext interface{}) (bool, error) {
	flow, ok := flows[strings.ToLower(arg)]
	if !ok {
		return false, nil
	}

	return true, flow.Run(logger, flowContext)
}
