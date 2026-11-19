package sdkgenerator

import (
	"bytes"
	"context"
	"fmt"
	"os"
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

	pr, pw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	srcCmd.Stdout = pw
	dstCmd.Stdin = pr

	if err := srcCmd.Start(); err != nil {
		_ = pw.Close()
		_ = pr.Close()
		return fmt.Errorf("failed to start %s: %w", srcName, err)
	}
	_ = pw.Close()

	if err := dstCmd.Start(); err != nil {
		_ = pr.Close()
		if srcCmd.Process != nil {
			_ = srcCmd.Process.Kill()
		}
		_ = srcCmd.Wait()
		return fmt.Errorf("failed to start %s: %w", dstName, err)
	}
	_ = pr.Close()

	dstErr := dstCmd.Wait()
	srcErr := srcCmd.Wait()

	if srcErr != nil && dstErr != nil {
		return fmt.Errorf("%s failed: %v; %s failed: %w", srcName, srcErr, dstName, dstErr)
	}
	if srcErr != nil {
		return fmt.Errorf("%s failed: %w", srcName, srcErr)
	}
	if dstErr != nil {
		return fmt.Errorf("%s failed: %w", dstName, dstErr)
	}

	return nil
}

