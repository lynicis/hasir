package sdkgenerator

import (
	"context"
)

// MockCommandRunner is a mock implementation of CommandRunner for testing.
type MockCommandRunner struct {
	RunFunc func(ctx context.Context, name string, args []string, workDir string) ([]byte, error)
	Calls   []MockRunCall
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

// Reset clears all recorded calls.
func (m *MockCommandRunner) Reset() {
	m.Calls = make([]MockRunCall, 0)
}
