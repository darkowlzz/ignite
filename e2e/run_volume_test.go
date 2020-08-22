package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"gotest.tools/assert"

	"github.com/weaveworks/ignite/e2e/util/ignite"
)

// runVolume is a helper for testing volume persistence
// vmName should be unique for each test
func runVolume(t *testing.T, vmName, runtime, networkPlugin string) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	// Create a loop device backed by a test-specific file
	volFile := "/tmp/" + vmName + "_vol"

	createDiskCmd := exec.Command(
		"dd",
		"if=/dev/zero",
		"of="+volFile,
		"bs=1M",
		"count=1024",
	)
	createDiskOut, createDiskErr := createDiskCmd.CombinedOutput()
	assert.Check(t, createDiskErr, fmt.Sprintf("create disk: \n%q\n%s", createDiskCmd.Args, createDiskOut))
	if createDiskErr != nil {
		return
	}
	defer func() {
		os.Remove(volFile)
	}()

	mkfsCmd := exec.Command(
		"mkfs.ext4",
		volFile,
	)
	mkfsOut, mkfsErr := mkfsCmd.CombinedOutput()
	assert.Check(t, mkfsErr, fmt.Sprintf("create disk: \n%q\n%s", mkfsCmd.Args, mkfsOut))
	if mkfsErr != nil {
		return
	}

	losetupCmd := exec.Command(
		"losetup",
		"--find",
		"--show",
		volFile,
	)
	losetupOut, losetupErr := losetupCmd.CombinedOutput()
	assert.Check(t, losetupErr, fmt.Sprintf("vm losetup: \n%q\n%s", losetupCmd.Args, losetupOut))
	if losetupErr != nil {
		return
	}

	loopPath := strings.TrimSpace(string(losetupOut))
	defer func() {
		detachLoopCmd := exec.Command(
			"losetup",
			"--detach",
			loopPath,
		)
		detachLoopOut, detachLoopErr := detachLoopCmd.CombinedOutput()
		assert.Check(t, detachLoopErr, fmt.Sprintf("loop detach: \n%q\n%s", detachLoopCmd.Args, detachLoopOut))
	}()

	// Run a vm with the loop-device mounted as a volume @ /my-vol
	i := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithVMName(vmName),
		ignite.WithRuntime(runtime),
		ignite.WithNetworkPlugin(networkPlugin),
		ignite.WithRunArg("--ssh"),
		ignite.WithRunArg("--volumes="+loopPath+":/my-vol"),
	)

	i.Run()
	defer i.Remove()

	// Touch a file in /my-vol
	i.Exec("touch", "/my-vol/hello-world")

	// Stop the vm without force.
	i.Force = false
	i.Stop()
	// Restore force to be used by VM remove.
	i.Force = true

	// Start another vm so we can check my-vol
	i2 := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithVMName(vmName+"_2"),
		ignite.WithRuntime(runtime),
		ignite.WithNetworkPlugin(networkPlugin),
		ignite.WithRunArg("--ssh"),
		ignite.WithRunArg("--volumes="+loopPath+":/my-vol"),
	)

	i2.Run()
	defer i2.Remove()

	// Stat the file in /my-vol using the new vm
	i2.Exec("stat", "/my-vol/hello-world")
}

func TestVolumeWithDockerAndDockerBridge(t *testing.T) {
	// TODO: https://github.com/weaveworks/ignite/issues/658
	t.Skip("SKIPPING\nThis test fails to stop the VM within docker\nTODO: https://github.com/weaveworks/ignite/issues/658")
	runVolume(
		t,
		"e2e_test_volume_docker_and_docker_bridge",
		"docker",
		"docker-bridge",
	)
}

func TestVolumeWithDockerAndCNI(t *testing.T) {
	runVolume(
		t,
		"e2e_test_volume_docker_and_cni",
		"docker",
		"cni",
	)
}

func TestVolumeWithContainerdAndCNI(t *testing.T) {
	runVolume(
		t,
		"e2e_test_volume_containerd_and_cni",
		"containerd",
		"cni",
	)
}
