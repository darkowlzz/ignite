package ignite

import (
	"testing"
)

// Ignite is a ignite command execution helper. It takes a binary with
// arguments to run with the binary.
type Ignite struct {
	T               *testing.T
	VMName          string
	Binary          string
	Arguments       []string
	RunArguments    []string
	RemoveArguments []string
	VMImage         string
	SandboxImage    string
	KernelImage     string
	Force           bool
}

// IgniteOption is used as functional option in the construction of Ignite
// helper.
type IgniteOption func(*Ignite)

// WithTest accepts a go test testing.T. It's used in error checks.
func WithTest(t *testing.T) IgniteOption {
	return func(i *Ignite) {
		i.T = t
	}
}

// WithBinary sets the ignite binary path.
func WithBinary(path string) IgniteOption {
	return func(i *Ignite) {
		i.Binary = path
	}
}

// WithArg accepts arguments that are used with all the ignite commands, to be
// used for global flags.
func WithArg(arg string) IgniteOption {
	return func(i *Ignite) {
		i.Arguments = append(i.Arguments, arg)
	}
}

// WithRuntime sets the ignite runtime value.
func WithRuntime(runtime string) IgniteOption {
	return WithArg("--runtime=" + runtime)
}

// WithNetworkPlugin sets the ignite network-plugin value.
func WithNetworkPlugin(networkPlugin string) IgniteOption {
	return WithArg("--network-plugin=" + networkPlugin)
}

// WithForce sets the force option for all the operations.
func WithForce(force bool) IgniteOption {
	return func(i *Ignite) {
		i.Force = force
	}
}

// NewIgnite returns a new ignite helper based on the passed ignite options.
func NewIgnite(opts ...IgniteOption) *Ignite {
	// Initialize with defaults.
	ignite := &Ignite{
		Binary:          "bin/ignite",
		RunArguments:    []string{"run"},
		RemoveArguments: []string{"rm"},
		VMImage:         "weaveworks/ignite-ubuntu",
		Force:           true,
	}

	for _, opt := range opts {
		opt(ignite)
	}

	return ignite
}
