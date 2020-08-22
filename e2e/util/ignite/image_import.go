package ignite

import (
	"fmt"
	"os/exec"

	"gotest.tools/assert"
)

// ImageImportArgs constructs image import args, given the image.
func (i *Ignite) ImageImportArgs(image string) []string {
	return append(i.Arguments, []string{"image", "import", image}...)
}

// ImageImportErr executes the image import command and returns the output and
// error.
func (i *Ignite) ImageImportErr(name string) ([]byte, error) {
	imgImportCmd := exec.Command(
		i.Binary,
		i.ImageImportArgs(name)...,
	)
	return imgImportCmd.CombinedOutput()
}

// ImageImport executes the image import command and performs an error check.
func (i *Ignite) ImageImport(name string) {
	imgImportOut, imgImportErr := i.ImageImportErr(name)
	assert.Check(i.T, imgImportErr, fmt.Sprintf("image import: \n%q\n%s", i.ImageImportArgs(name), imgImportOut))
}
