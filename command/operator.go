package command

import "github.com/privatix/dapp-installer/dapp"

// operator implement the Runner interface in pipeline package.
type operator struct {
	name   string
	run    func(*dapp.Dapp) error
	cancel func(*dapp.Dapp) error
}

// Name returns the operators name.
func (o operator) Name() string {
	return o.name
}

// Run executes the operators run function.
func (o operator) Run(in interface{}) error {
	return o.run(in.(*dapp.Dapp))
}

// Cancel executes the operators cancel function.
func (o operator) Cancel(in interface{}) error {
	return o.cancel(in.(*dapp.Dapp))
}

func newOperator(name string, run func(*dapp.Dapp) error,
	cancel func(*dapp.Dapp) error) operator {
	if cancel == nil {
		cancel = func(in *dapp.Dapp) error { return nil }
	}
	return operator{name: name, run: run, cancel: cancel}
}
