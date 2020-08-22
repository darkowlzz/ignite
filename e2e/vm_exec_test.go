package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"gotest.tools/assert"

	"github.com/weaveworks/ignite/e2e/util/ignite"
)

func TestVMExecInteractive(t *testing.T) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	vmName := "e2e_test_ignite_exec_interactive"

	i := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithVMName(vmName),
		ignite.WithRunArg("--ssh"),
	)

	i.Run()
	defer i.Remove()

	// Pass input data from host and write to a file inside the VM.
	remoteFileName := "afile.txt"
	inputContent := "foooo..."
	input := strings.NewReader(inputContent)

	execCmd := exec.Command(
		igniteBin,
		"exec", vmName,
		"tee", remoteFileName,
	)
	execCmd.Stdin = input

	execOut, execErr := execCmd.CombinedOutput()
	assert.Check(t, execErr, fmt.Sprintf("exec: \n%q\n%s", execCmd.Args, execOut))

	// Check the file content inside the VM.
	execArgs := []string{"cat", remoteFileName}
	catOut, catErr := i.ExecErr(execArgs...)
	assert.Check(t, catErr, fmt.Sprintf("cat: \n%q\n%s", i.ExecArgs(execArgs...), catOut))
	assert.Equal(t, string(catOut), inputContent, fmt.Sprintf("unexpected file content on host:\n\t(WNT): %q\n\t(GOT): %q", inputContent, string(catOut)))
}
