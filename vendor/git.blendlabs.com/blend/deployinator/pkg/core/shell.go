package core

import (
	"os"
	"os/exec"

	exception "github.com/blend/go-sdk/exception"
)

// ShellExecute runs a command as a subprocess.
func ShellExecute(workDir, cmd string, args ...string) error {
	cmdFullPath, err := exec.LookPath(cmd)

	if err != nil {
		return exception.New(err)
	}

	execCmd := exec.Command(cmdFullPath, args...)
	execCmd.Dir = workDir
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return exception.New(execCmd.Run())
}

// ShellCapture runs a command and returns the output.
func ShellCapture(workDir, cmd string, args ...string) ([]byte, error) {
	cmdFullPath, err := exec.LookPath(cmd)

	if err != nil {
		return nil, exception.New(err)
	}

	execCmd := exec.Command(cmdFullPath, args...)
	execCmd.Dir = workDir
	output, err := execCmd.CombinedOutput()
	return output, exception.New(err)
}

// Command returns a command in the working dir ensuring that the command exists in the path
func Command(workDir, cmd string, args ...string) (*exec.Cmd, error) {
	cmdFullPath, err := exec.LookPath(cmd)

	if err != nil {
		return nil, exception.New(err)
	}

	execCmd := exec.Command(cmdFullPath, args...)
	execCmd.Dir = workDir
	return execCmd, nil
}
