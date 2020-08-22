package ignite

import (
	"fmt"
	"os/exec"

	"gotest.tools/assert"
)

// StopArgs returns all the arguments used with stop command.
func (i *Ignite) StopArgs() []string {
	args := append(i.Arguments, "stop")
	// Append force stop flag.
	if i.Force {
		args = append(args, "-f")
	}
	// Append the last argument, VM name.
	return append(args, i.VMName)
}

// StopErr executes the stop command and return the stop output and error.
func (i *Ignite) StopErr() ([]byte, error) {
	stopCmd := exec.Command(
		i.Binary,
		i.StopArgs()...,
	)
	return stopCmd.CombinedOutput()
}

// Stop executes the stop command and performs an error check.
func (i *Ignite) Stop() {
	stopOut, stopErr := i.StopErr()
	assert.Check(i.T, stopErr, fmt.Sprintf("vm stop: \n%q\n%s", i.StopArgs(), stopOut))
	// Issue fatal error if err isn't nil.
	// For tests that want to continue even after a VM stop failure, use
	// StopErr.
	if stopErr != nil {
		i.T.Fatal("failed to stop VM", stopErr)
	}
}
