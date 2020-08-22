package ignite

import (
	"fmt"
	"os/exec"

	"gotest.tools/assert"
)

// ImageRmArgs constructs image rm args, given the image.
func (i *Ignite) ImageRmArgs(image string) []string {
	// Construct image rm args, append the image name and options.
	args := append(i.Arguments, []string{"image", "rm", image}...)
	if i.Force {
		args = append(args, "-f")
	}
	return args
}

// ImageRemoveErr executes the image remove command and returns the output and
// error.
func (i *Ignite) ImageRemoveErr(name string) ([]byte, error) {
	imgRmCmd := exec.Command(
		i.Binary,
		i.ImageRmArgs(name)...,
	)
	return imgRmCmd.CombinedOutput()
}

// ImageRemove executes the image remove command and performs an error check.
func (i *Ignite) ImageRemove(name string) {
	imgRmOut, imgRmErr := i.ImageRemoveErr(name)
	assert.Check(i.T, imgRmErr, fmt.Sprintf("image rm: \n%q\n%s", i.ImageRmArgs(name), imgRmOut))
}
