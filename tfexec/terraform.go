package tfexec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

// tfVersionRe is a pattern to parse outputs from terraform version.
var tfVersionRe = regexp.MustCompile(`^Terraform v(.+)\s*$`)

// State is a named type for tfstate.
// We doesn't need to parse contents of tfstate,
// but we define it as a name type to clarify TerraformCLI interface.
type State string

// TerraformCLI is an interface for executing the terraform command.
// The main features of the terraform command are many of side effects, and the
// most of stdout may not be useful. In addition, the interfaces of state
// subcommands are inconsistent, and if a state file is required for the
// argument, we need a temporary file. However, It's hard to clean up the
// temporary file when an error occurs in the middle of a series  of commands.
// This means implementing the exactly same interface for the terraform command
// doesn't make sense for us. So we wrap the terraform command and provider a
// high-level and easy-to-use interface which can be used in memory as much as
// possible.
type TerraformCLI interface {
	// Verison returns a version number of Terraform.
	Version(ctx context.Context) (string, error)

	// StatePull returns the current tfstate from remote.
	StatePull(ctx context.Context) (State, error)

	// StatePush pushs a given State to remote.
	StatePush(ctx context.Context, state State) error
}

// terraformCLI implements the TerraformCLI interface.
type terraformCLI struct {
	// Executor is an interface which executes an arbitrary command.
	Executor
}

var _ TerraformCLI = (*terraformCLI)(nil)

// NewTerraformCLI returns an implementation of the TerraformCLI interface.
func NewTerraformCLI(e Executor) TerraformCLI {
	return &terraformCLI{
		Executor: e,
	}
}

// run is a helper method for running terraform comamnd.
func (c *terraformCLI) run(ctx context.Context, args ...string) (string, error) {
	cmd, err := c.Executor.NewCommandContext(ctx, "terraform", args...)
	if err != nil {
		return "", err
	}

	err = c.Executor.Run(cmd)
	if err != nil {
		return "", err
	}

	return cmd.Stdout(), err
}

// Verison returns a version number of Terraform.
func (c *terraformCLI) Version(ctx context.Context) (string, error) {
	stdout, err := c.run(ctx, "version")
	if err != nil {
		return "", err
	}

	matched := tfVersionRe.FindStringSubmatch(stdout)
	if len(matched) != 2 {
		return "", fmt.Errorf("failed to parse terraform version: %s", stdout)
	}
	version := matched[1]
	return version, nil
}

// StatePull returns the current tfstate from remote.
func (c *terraformCLI) StatePull(ctx context.Context) (State, error) {
	stdout, err := c.run(ctx, "state", "pull")
	if err != nil {
		return "", err
	}

	return State(stdout), nil
}

// StatePush pushs a given State to remote.
func (c *terraformCLI) StatePush(ctx context.Context, state State) error {
	tmpfile, err := writeTempFile([]byte(state))
	defer os.Remove(tmpfile.Name())
	if err != nil {
		return err
	}

	_, err = c.run(ctx, "state", "push", tmpfile.Name())
	return err
}

// writeTempFile writes content to a temporary file and return its file.
func writeTempFile(content []byte) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		return tmpfile, fmt.Errorf("failed to create temporary file: %s", err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		return tmpfile, fmt.Errorf("failed to write temporary file: %s", err)
	}

	if err := tmpfile.Close(); err != nil {
		return tmpfile, fmt.Errorf("failed to close temporary file: %s", err)
	}

	return tmpfile, nil
}