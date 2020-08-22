package ignite

import (
	"fmt"
	"os/exec"

	"gotest.tools/assert"
)

// ExecArgs constructs VM exec args, given the arguments.
func (i *Ignite) ExecArgs(args ...string) []string {
	// Construct exec args, append the VM name and provided args.
	execArgs := append(i.Arguments, "exec")
	execArgs = append(execArgs, i.VMName)
	return append(execArgs, args...)
}

// ExecErr executes the exec command and returns the exec output and error.
func (i *Ignite) ExecErr(args ...string) ([]byte, error) {
	execCmd := exec.Command(
		i.Binary,
		i.ExecArgs(args...)...,
	)
	return execCmd.CombinedOutput()
}

// Exec executes the exec command and performs an error check.
func (i *Ignite) Exec(args ...string) {
	execOut, execErr := i.ExecErr(args...)
	assert.Check(i.T, execErr, fmt.Sprintf("exec: \n%q\n%s", i.ExecArgs(args...), execOut))
}
