package ignite

import (
	"fmt"
	"os/exec"

	"gotest.tools/assert"
)

// PsArgs constructs ps args, given the arguments.
func (i *Ignite) PsArgs(args ...string) []string {
	psArgs := append(i.Arguments, "ps")
	return append(psArgs, args...)
}

// PsErr executes the ps command and returns the output and error.
func (i *Ignite) PsErr(args ...string) ([]byte, error) {
	psCmd := exec.Command(
		i.Binary,
		i.PsArgs(args...)...,
	)
	return psCmd.CombinedOutput()
}

// Ps executes the ps command and performs an error check.
func (i *Ignite) Ps(args ...string) {
	psOut, psErr := i.PsErr(args...)
	assert.Check(i.T, psErr, fmt.Sprintf("ps: \n%q\n%s", i.PsArgs(args...), psOut))
}
