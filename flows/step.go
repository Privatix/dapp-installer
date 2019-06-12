package flows

import "github.com/privatix/dapp-installer/dapp"

// step is an abstract wrapper around step of core software
// lifecycle install/update etc.
type step struct {
	name string
	do   func(*dapp.Dapp) error
	undo func(*dapp.Dapp) error
}

// Name returns the steps name.
func (o step) Name() string {
	return o.name
}

// Do executes the step.
func (o step) Do(in interface{}) error {
	return o.do(in.(*dapp.Dapp))
}

// Undo executes the steps cancel function.
func (o step) Undo(in interface{}) error {
	return o.undo(in.(*dapp.Dapp))
}

func newStep(name string, do func(*dapp.Dapp) error,
	undo func(*dapp.Dapp) error) step {
	blank := func(in *dapp.Dapp) error { return nil }
	if do == nil {
		do = blank
	}
	if undo == nil {
		undo = blank
	}
	return step{name: name, do: do, undo: undo}
}
