package sdkgenerator

import (
	"context"
)

// MockCommandRunner is a mock implementation of CommandRunner for testing.
type MockCommandRunner struct {
	RunFunc         func(ctx context.Context, name string, args []string, workDir string) ([]byte, error)
	RunCombinedFunc func(ctx context.Context, name string, args []string, workDir string) ([]byte, error)
	PipeFunc        func(ctx context.Context, srcName string, srcArgs []string, srcDir, dstName string, dstArgs []string, dstDir string) error
	Calls           []MockRunCall
}

type MockRunCall struct {
	Name    string
	Args    []string
	WorkDir string
}

func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		Calls: make([]MockRunCall, 0),
	}
}

func (m *MockCommandRunner) Run(ctx context.Context, name string, args []string, workDir string) ([]byte, error) {
	m.Calls = append(m.Calls, MockRunCall{
		Name:    name,
		Args:    args,
		WorkDir: workDir,
	})

	if m.RunFunc != nil {
		return m.RunFunc(ctx, name, args, workDir)
	}

	return nil, nil
}

func (m *MockCommandRunner) RunCombined(ctx context.Context, name string, args []string, workDir string) ([]byte, error) {
	m.Calls = append(m.Calls, MockRunCall{
		Name:    name,
		Args:    args,
		WorkDir: workDir,
	})

	if m.RunCombinedFunc != nil {
		return m.RunCombinedFunc(ctx, name, args, workDir)
	}

	return nil, nil
}

func (m *MockCommandRunner) Pipe(
	ctx context.Context,
	srcName string,
	srcArgs []string,
	srcDir string,
	dstName string,
	dstArgs []string,
	dstDir string,
) error {
	m.Calls = append(m.Calls,
		MockRunCall{Name: srcName, Args: srcArgs, WorkDir: srcDir},
		MockRunCall{Name: dstName, Args: dstArgs, WorkDir: dstDir},
	)

	if m.PipeFunc != nil {
		return m.PipeFunc(ctx, srcName, srcArgs, srcDir, dstName, dstArgs, dstDir)
	}

	return nil
}

// Reset clears all recorded calls.
func (m *MockCommandRunner) Reset() {
	m.Calls = make([]MockRunCall, 0)
}
