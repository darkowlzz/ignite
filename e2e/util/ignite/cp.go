package ignite

import (
	"fmt"
	"os/exec"

	"gotest.tools/assert"
)

// CopyArgs constructs copy args, given source and destination.
func (i *Ignite) CopyArgs(src, dst string) []string {
	return append(i.Arguments, []string{"cp", src, dst}...)
}

// CopyErr executes the copy command and returns the output and error.
func (i *Ignite) CopyErr(src, dst string) ([]byte, error) {
	cpCmd := exec.Command(
		i.Binary,
		i.CopyArgs(src, dst)...,
	)
	return cpCmd.CombinedOutput()
}

// Copy executes the copy command and performs an error check.
func (i *Ignite) Copy(src, dst string) {
	cpOut, cpErr := i.CopyErr(src, dst)
	assert.Check(i.T, cpErr, fmt.Sprintf("cp: \n%q\n%s", i.CopyArgs(src, dst), cpOut))
}
