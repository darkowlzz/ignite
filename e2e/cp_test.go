package e2e

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/weaveworks/ignite/cmd/ignite/run"
	"github.com/weaveworks/ignite/e2e/util/ignite"
)

func runCopyFilesToVM(t *testing.T, vmName, source, destination, wantFileContent string) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	i := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithVMName(vmName),
		ignite.WithRunArg("--ssh"),
	)

	i.Run()
	defer i.Remove()

	i.Copy(source, destination)

	// When copying to a VM, the file path succeeds the file path separator.
	// Split the destination to obtain VM destination file path.
	dest := strings.Split(destination, run.VMFilePathSeparator)
	execArgs := []string{"cat", dest[1]}
	catOut, catErr := i.ExecErr(execArgs...)
	assert.Check(t, catErr, fmt.Sprintf("cat: \n%q\n%s", i.ExecArgs(execArgs...), catOut))
	assert.Equal(t, string(catOut), wantFileContent, fmt.Sprintf("unexpected copied file content:\n\t(WNT): %q\n\t(GOT): %q", wantFileContent, string(catOut)))
}

func TestCopyFileFromHostToVM(t *testing.T) {
	cases := []struct {
		name    string
		content []byte
	}{
		{
			name:    "file_with_content",
			content: []byte("some example file content"),
		},
		{
			name:    "empty_file",
			content: []byte(""),
		},
	}

	for _, rt := range cases {
		t.Run(rt.name, func(t *testing.T) {
			// Create a file.
			file, err := ioutil.TempFile("", "ignite-cp-test")
			if err != nil {
				t.Fatalf("failed to create a file: %v", err)
			}
			defer os.Remove(file.Name())

			// Populate the file.
			if _, err := file.Write(rt.content); err != nil {
				t.Fatalf("failed to write to a file: %v", err)
			}
			if err := file.Close(); err != nil {
				t.Errorf("failed to close file: %v", err)
			}

			vmName := "e2e_test_copy_to_vm_" + rt.name
			runCopyFilesToVM(
				t,
				vmName,
				file.Name(),
				fmt.Sprintf("%s:%s", vmName, file.Name()),
				string(rt.content),
			)
		})
	}
}

func TestCopySymlinkedFileFromHostToVM(t *testing.T) {
	// Create a file.
	file, err := ioutil.TempFile("", "ignite-symlink-cp-test")
	if err != nil {
		t.Fatalf("failed to create a file: %v", err)
	}
	defer os.Remove(file.Name())

	fileContent := []byte("Some file content.")

	if _, err := file.Write(fileContent); err != nil {
		t.Fatalf("failed to write to a file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Errorf("failed to close file: %v", err)
	}

	// Create a new file symlinked to the first file.
	newName := fmt.Sprintf("%s-link", file.Name())
	if err := os.Symlink(file.Name(), newName); err != nil {
		t.Errorf("failed to create symlink: %v", err)
	}
	defer os.Remove(newName)

	vmName := "e2e_test_copy_symlink_to_vm"

	runCopyFilesToVM(
		t,
		vmName,
		newName,
		fmt.Sprintf("%s:%s", vmName, newName),
		string(fileContent),
	)
}

func TestCopyFileFromVMToHost(t *testing.T) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	vmName := "e2e_test_copy_file_from_vm_to_host"

	i := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithVMName(vmName),
		ignite.WithRunArg("--ssh"),
	)

	i.Run()
	defer i.Remove()

	// File to be copied from VM.
	vmFilePath := "/proc/version"
	execArgs := []string{"cat", vmFilePath}
	catOut, catErr := i.ExecErr(execArgs...)
	assert.Check(t, catErr, fmt.Sprintf("cat: \n%q\n%s", i.ExecArgs(execArgs...), catOut))

	// Host file path.
	hostFilePath := "/tmp/ignite-os-version"
	vmSource := fmt.Sprintf("%s:%s", vmName, vmFilePath)
	i.Copy(vmSource, hostFilePath)
	defer os.Remove(hostFilePath)

	hostContent, err := ioutil.ReadFile(hostFilePath)
	if err != nil {
		t.Errorf("failed to read host file content: %v", err)
	}

	// NOTE: Since the output of cat in the VM includes newline with "\r\n" but
	// the content of file on the host has "\n" for newline when read using go,
	// trim the whitespaces and compare the result.
	got := strings.TrimSpace(string(hostContent))
	want := strings.TrimSpace(string(catOut))
	assert.Equal(t, got, want, fmt.Sprintf("unexpected copied file content:\n\t(WNT): %q\n\t(GOT): %q", want, got))
}

func TestCopyDirectoryFromHostToVM(t *testing.T) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	// Create a temporary directory on host.
	dir, err := ioutil.TempDir("", "ignite-cp-dir-test")
	if err != nil {
		t.Fatalf("failed to create a directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create a file in the directory.
	file, err := ioutil.TempFile(dir, "ignite-cp-file")
	if err != nil {
		t.Fatalf("failed to create a file: %v", err)
	}
	content := []byte("some file content")
	if _, err := file.Write(content); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Errorf("failed to close file: %v", err)
	}

	vmName := "e2e_test_copy_dir_to_vm"
	source := dir
	dest := fmt.Sprintf("%s:%s", vmName, source)

	i := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithVMName(vmName),
		ignite.WithRunArg("--ssh"),
	)

	// Run a VM.
	i.Run()
	defer i.Remove()

	// Copy dir to VM.
	i.Copy(source, dest)

	// Check if the directory exists in the VM.
	dirFind := fmt.Sprintf("find %s -type d -name %s", filepath.Dir(source), filepath.Base(source))
	dirFindOut, dirFindErr := i.ExecErr(dirFind)
	assert.Check(t, dirFindErr, fmt.Sprintf("find: \n%q\n%s", i.ExecArgs(dirFind), dirFindOut))
	gotDir := strings.TrimSpace(string(dirFindOut))
	assert.Equal(t, gotDir, dir, fmt.Sprintf("unexpected find directory result: \n\t(WNT): %q\n\t(GOT): %q", dir, gotDir))

	// Check if the file inside the directory in the VM has the same content as
	// on the host.
	execArgs := []string{"cat", file.Name()}
	catOut, catErr := i.ExecErr(execArgs...)
	assert.Check(t, catErr, fmt.Sprintf("cat: \n%q\n%s", i.ExecArgs(execArgs...), catOut))
	gotContent := strings.TrimSpace(string(catOut))
	assert.Equal(t, gotContent, string(content), fmt.Sprintf("unexpected copied file content:\n\t(WNT): %q\n\t(GOT): %q", content, gotContent))
}

func TestCopyDirectoryFromVMToHost(t *testing.T) {
	assert.Assert(t, e2eHome != "", "IGNITE_E2E_HOME should be set")

	vmName := "e2e_test_copy_dir_from_vm_to_host"

	i := ignite.NewIgnite(
		ignite.WithTest(t),
		ignite.WithBinary(igniteBin),
		ignite.WithVMName(vmName),
		ignite.WithRunArg("--ssh"),
	)

	// Run a VM.
	i.Run()
	defer i.Remove()

	// Create directory inside the VM.
	rand.Seed(time.Now().UnixNano())
	dirPath := fmt.Sprintf("/tmp/ignite-cp-dir-test%d", rand.Intn(10000))
	mkdir := fmt.Sprintf("mkdir -p %s", dirPath)
	i.Exec(mkdir)

	// Create file inside the directory.
	content := "some content on VM"
	filePath := filepath.Join(dirPath, "ignite-cp-file")
	writeFile := fmt.Sprintf("echo %s > %s", content, filePath)
	i.Exec(writeFile)

	// Copy the file to host.
	src := fmt.Sprintf("%s:%s", vmName, dirPath)
	i.Copy(src, dirPath)
	defer os.RemoveAll(dirPath)

	// Find copied directory on host.
	if _, err := os.Stat(dirPath); err != nil {
		assert.Check(t, err, fmt.Sprintf("error while checking if dir %q exists: %v", dirPath, err))
	}

	// Check the content of the file inside the copied directory.
	hostContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Errorf("failed to read host file %q content: %v", filePath, err)
	}
	gotContent := strings.TrimSpace(string(hostContent))
	assert.Equal(t, gotContent, content, fmt.Sprintf("unexpected copied file content:\n\t(WNT): %q\n\t(GOT): %q", content, gotContent))
}
