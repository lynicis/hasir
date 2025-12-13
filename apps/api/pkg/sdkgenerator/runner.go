package sdkgenerator

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type DefaultCommandRunner struct{}

func NewDefaultCommandRunner() *DefaultCommandRunner {
	return &DefaultCommandRunner{}
}

func (r *DefaultCommandRunner) Run(ctx context.Context, name string, args []string, workDir string) ([]byte, error) {
	// #nosec G204 -- name and args are validated by the caller
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
