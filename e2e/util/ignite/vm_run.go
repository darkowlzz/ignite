package ignite

import (
	"fmt"
	"os/exec"

	"gotest.tools/assert"
)

// WithRunArg accepts arguments that are specific to run command.
func WithRunArg(arg string) IgniteOption {
	return func(i *Ignite) {
		i.RunArguments = append(i.RunArguments, arg)
	}
}

func WithVMImage(image string) IgniteOption {
	return func(i *Ignite) {
		i.VMImage = image
	}
}

func WithSandboxImage(image string) IgniteOption {
	return func(i *Ignite) {
		i.SandboxImage = image
	}
}

func WithKernelImage(image string) IgniteOption {
	return func(i *Ignite) {
		i.KernelImage = image
	}
}

func WithVMName(name string) IgniteOption {
	return func(i *Ignite) {
		i.VMName = name
	}
}

// RunArgs returns all the arguments used with run command.
func (i *Ignite) RunArgs() []string {
	args := append(i.Arguments, i.RunArguments...)
	args = append(args, "--name="+i.VMName)

	if i.KernelImage != "" {
		args = append(args, "--kernel-image="+i.KernelImage)
	}

	if i.SandboxImage != "" {
		args = append(args, "--sandbox-image="+i.SandboxImage)
	}

	// Append the last argument, VM image.
	return append(args, i.VMImage)
}

// RunErr executes the run command and returns the run output and error.
func (i *Ignite) RunErr() ([]byte, error) {
	runCmd := exec.Command(
		i.Binary,
		i.RunArgs()...,
	)
	return runCmd.CombinedOutput()
}

// Run executes the run command and performs an error check.
func (i *Ignite) Run() {
	runOut, runErr := i.RunErr()
	assert.Check(i.T, runErr, fmt.Sprintf("vm run: \n%q\n%s", i.RunArgs(), runOut))
	// Issue fatal error if err isn't nil.
	// For tests that want to continue even after a VM run failure, use RunErr.
	if runErr != nil {
		i.T.Fatal("failed to run VM", runErr)
	}
}
