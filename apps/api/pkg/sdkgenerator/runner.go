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

func (r *DefaultCommandRunner) RunCombined(ctx context.Context, name string, args []string, workDir string) ([]byte, error) {
	// #nosec G204 -- name and args are validated by the caller
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir

	return cmd.CombinedOutput()
}

func (r *DefaultCommandRunner) Pipe(
	ctx context.Context,
	srcName string,
	srcArgs []string,
	srcDir string,
	dstName string,
	dstArgs []string,
	dstDir string,
) error {
	// #nosec G204 -- command names and args are validated by the caller
	srcCmd := exec.CommandContext(ctx, srcName, srcArgs...)
	srcCmd.Dir = srcDir

	// #nosec G204 -- command names and args are validated by the caller
	dstCmd := exec.CommandContext(ctx, dstName, dstArgs...)
	dstCmd.Dir = dstDir

	pipeOutput, err := srcCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}
	dstCmd.Stdin = pipeOutput

	if err := srcCmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", srcName, err)
	}

	if err := dstCmd.Start(); err != nil {
		_ = srcCmd.Wait()
		return fmt.Errorf("failed to start %s: %w", dstName, err)
	}

	if err := srcCmd.Wait(); err != nil {
		_ = dstCmd.Wait()
		return fmt.Errorf("%s failed: %w", srcName, err)
	}

	if err := dstCmd.Wait(); err != nil {
		return fmt.Errorf("%s failed: %w", dstName, err)
	}

	return nil
}

