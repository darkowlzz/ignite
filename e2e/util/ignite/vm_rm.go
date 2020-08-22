package ignite

import (
	"fmt"
	"os/exec"

	"gotest.tools/assert"
)

// WithRemoveArg accepts arguments that are specific to remove command.
func WithRemoveArg(arg string) IgniteOption {
	return func(i *Ignite) {
		i.RemoveArguments = append(i.RemoveArguments, arg)
	}
}

// RemoveArgs returns all the arguments used with remove command.
func (i *Ignite) RemoveArgs() []string {
	args := append(i.Arguments, i.RemoveArguments...)
	// Append force remove flag.
	if i.Force {
		args = append(args, "-f")
	}
	// Append the last argument, VM name.
	return append(args, i.VMName)
}

// RemoveErr executes the remove command and return the remove output and
// error.
func (i *Ignite) RemoveErr() ([]byte, error) {
	rmCmd := exec.Command(
		i.Binary,
		i.RemoveArgs()...,
	)
	return rmCmd.CombinedOutput()
}

// Remove executes the remove command and performs an error check.
func (i *Ignite) Remove() {
	rmvOut, rmvErr := i.RemoveErr()
	assert.Check(i.T, rmvErr, fmt.Sprintf("vm removal: \n%q\n%s", i.RemoveArgs(), rmvOut))
}
