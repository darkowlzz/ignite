package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"gotest.tools/assert"

	"github.com/weaveworks/ignite/e2e/util/ignite"
	"github.com/weaveworks/ignite/pkg/constants"
)

func TestConfigFile(t *testing.T) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	cases := []struct {
		name             string
		config           []byte
		vmConfig         []byte
		args             []string
		wantVMProperties string
		err              bool
	}{
		{
			name:   "invalid config",
			config: []byte(``),
			err:    true,
		},
		{
			name: "minimal valid config",
			config: []byte(`---
apiVersion: ignite.weave.works/v1alpha3
kind: Configuration
`),
			wantVMProperties: fmt.Sprintf("'512.0 MB 1 4.0 GB weaveworks/ignite-ubuntu:latest weaveworks/ignite:dev weaveworks/ignite-kernel:%s <nil>'", constants.DEFAULT_KERNEL_IMAGE_TAG),
		},
		{
			name: "custom vm properties",
			config: []byte(`---
apiVersion: ignite.weave.works/v1alpha3
kind: Configuration
metadata:
  name: test-config
spec:
  vmDefaults:
    memory: "2GB"
    diskSize: "3GB"
    cpus: 2
    sandbox:
      oci: weaveworks/ignite:dev
    kernel:
      oci: weaveworks/ignite-kernel:5.4.43
    ssh: true
`),
			wantVMProperties: "'2.0 GB 2 3.0 GB weaveworks/ignite-ubuntu:latest weaveworks/ignite:dev weaveworks/ignite-kernel:5.4.43 {true }'",
		},
		{
			name: "runtime and network config",
			config: []byte(`---
apiVersion: ignite.weave.works/v1alpha3
kind: Configuration
metadata:
  name: test-config
spec:
  runtime: docker
  networkPlugin: docker-bridge
`),
			wantVMProperties: fmt.Sprintf("'512.0 MB 1 4.0 GB weaveworks/ignite-ubuntu:latest weaveworks/ignite:dev weaveworks/ignite-kernel:%s <nil>'", constants.DEFAULT_KERNEL_IMAGE_TAG),
		},
		{
			name: "override properties",
			config: []byte(`---
apiVersion: ignite.weave.works/v1alpha3
kind: Configuration
metadata:
  name: test-config
spec:
  vmDefaults:
    memory: "2GB"
    diskSize: "3GB"
    cpus: 2
`),
			args:             []string{"--memory=1GB", "--size=1GB", "--cpus=1", "--ssh"},
			wantVMProperties: fmt.Sprintf("'1024.0 MB 1 1024.0 MB weaveworks/ignite-ubuntu:latest weaveworks/ignite:dev weaveworks/ignite-kernel:%s {true }'", constants.DEFAULT_KERNEL_IMAGE_TAG),
		},
		{
			name: "vm config",
			config: []byte(`---
apiVersion: ignite.weave.works/v1alpha3
kind: Configuration
metadata:
  name: test-config
spec:
  vmDefaults:
    memory: "2GB"
    diskSize: "3GB"
    cpus: 2
`),
			vmConfig: []byte(`
apiVersion: ignite.weave.works/v1alpha3
kind: VM
spec:
  memory: "1GB"
  diskSize: "2GB"
  cpus: 1
`),
			wantVMProperties: fmt.Sprintf("'1024.0 MB 1 2.0 GB weaveworks/ignite-ubuntu:latest weaveworks/ignite:dev weaveworks/ignite-kernel:%s <nil>'", constants.DEFAULT_KERNEL_IMAGE_TAG),
		},
		{
			name: "vm config and flags",
			config: []byte(`---
apiVersion: ignite.weave.works/v1alpha3
kind: Configuration
metadata:
  name: test-config
spec:
  vmDefaults:
    memory: "2GB"
    diskSize: "3GB"
    cpus: 2
`),
			vmConfig: []byte(`
apiVersion: ignite.weave.works/v1alpha3
kind: VM
spec:
  memory: "1GB"
  diskSize: "2GB"
`),
			args:             []string{"--size=1GB", "--cpus=1"},
			wantVMProperties: fmt.Sprintf("'1024.0 MB 1 1024.0 MB weaveworks/ignite-ubuntu:latest weaveworks/ignite:dev weaveworks/ignite-kernel:%s <nil>'", constants.DEFAULT_KERNEL_IMAGE_TAG),
		},
	}

	for _, rt := range cases {
		rt := rt
		t.Run(rt.name, func(t *testing.T) {
			// Create config file.
			file, err := ioutil.TempFile("", "ignite-config-file-test")
			if err != nil {
				t.Fatalf("failed to create a file: %v", err)
			}
			defer os.Remove(file.Name())

			// Populate the file.
			_, err = file.Write(rt.config)
			assert.NilError(t, err)
			assert.NilError(t, file.Close())

			vmConfigFileName := ""

			if len(rt.vmConfig) > 0 {
				// Create a VM config file.
				vmConfigFile, err := ioutil.TempFile("", "ignite-vm-config")
				if err != nil {
					t.Fatalf("failed to create a file: %v", err)
				}
				defer os.Remove(vmConfigFile.Name())

				vmConfigFileName = vmConfigFile.Name()

				// Populate the file.
				_, err = vmConfigFile.Write(rt.vmConfig)
				assert.NilError(t, err)
				assert.NilError(t, vmConfigFile.Close())
			}

			vmName := "e2e_test_ignite_config_file"

			// Create a VM with the ignite config file.
			// NOTE: Set a sandbox-image to have deterministic results.

			i := ignite.NewIgnite(
				ignite.WithTest(t),
				ignite.WithBinary(igniteBin),
				ignite.WithVMName(vmName),
				ignite.WithSandboxImage("weaveworks/ignite:dev"),
				ignite.WithArg("--ignite-config="+file.Name()),
			)

			// Append VM config if provided.
			if vmConfigFileName != "" {
				i.RunArguments = append(i.RunArguments, "--config="+vmConfigFileName)
			}

			// Append the args to the run args for override flags.
			i.RunArguments = append(i.RunArguments, rt.args...)

			_, err = i.RunErr()

			if err == nil {
				// Delete the VM only when the creation succeeds, with the
				// config file.
				defer i.Remove()

				// Check if run failure was expected.
				if (err != nil) != rt.err {
					t.Error("expected VM creation failure")
				}
			}

			if !rt.err {
				// Query VM properties.
				psArgs := []string{
					"--filter={{.ObjectMeta.Name}}=" + vmName,
					"--template='{{.Spec.Memory}} {{.Spec.CPUs}} {{.Spec.DiskSize}} {{.Spec.Image.OCI}} {{.Spec.Sandbox.OCI}} {{.Spec.Kernel.OCI}} {{.Spec.SSH}}'",
				}
				psOut, psErr := i.PsErr(psArgs...)
				assert.Check(t, psErr, fmt.Sprintf("ps: \n%q\n%s", i.PsArgs(psArgs...), psOut))
				got := strings.TrimSpace(string(psOut))
				assert.Equal(t, got, rt.wantVMProperties, fmt.Sprintf("unexpected VM properties:\n\t(WNT): %q\n\t(GOT): %q", rt.wantVMProperties, got))
			}
		})
	}
}
