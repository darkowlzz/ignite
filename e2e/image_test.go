package e2e

import (
	"os/exec"
	"testing"

	"gotest.tools/assert"

	"github.com/weaveworks/ignite/e2e/util/ignite"
)

func TestImportTinyImage(t *testing.T) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	// NOTE: Along with tiny image, this also tests the image import failure
	// when there's no /etc directory in the image filesystem.

	testImage := "hello-world:latest"

	i := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithForce(false),
	)

	// Remove if the image already exists.
	// Ignore any remove error.
	_, _ = i.ImageRemoveErr(testImage)

	// Import the image.
	i.ImageImport(testImage)
}

func TestDockerImportImage(t *testing.T) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	testImage := "hello-world:latest"

	// Setup default ignite.
	ic := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithForce(false),
	)

	// Remove if the image already exists.
	_, _ = ic.ImageRemoveErr(testImage)

	// Remove image from docker image store if already exists.
	rmvDockerImgCmd := exec.Command(
		"docker",
		"rmi", testImage,
	)
	// Ignore error if the image doesn't exists.
	_, _ = rmvDockerImgCmd.CombinedOutput()

	// Setup ignite with docker runtime.
	id := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithRuntime("docker"),
		ignite.WithForce(false),
	)

	// Import the image.
	id.ImageImport(testImage)
}
