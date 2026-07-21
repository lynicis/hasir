package sdkgenerator

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSDK_DirName(t *testing.T) {
	tests := []struct {
		sdk      SDK
		expected string
	}{
		{SdkGoProtobuf, "go-protobuf"},
		{SdkGoConnectRpc, "go-connectrpc"},
		{SdkGoGrpc, "go-grpc"},
		{SdkJsBufbuildEs, "js-bufbuild-es"},
		{SdkJsProtobuf, "js-protobuf"},
		{SdkJsConnectrpc, "js-connectrpc"},
		{SdkRustProtobuf, "rust-protobuf"},
		{SdkRustGrpc, "rust-grpc"},
		{SdkJavaProtobuf, "java-protobuf"},
		{SdkJavaGrpc, "java-grpc"},
		{SdkCsharpProtobuf, "csharp-protobuf"},
		{SdkCsharpGrpc, "csharp-grpc"},
		{SDK("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sdk), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.sdk.DirName())
		})
	}
}

func TestValidateProtoFile(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{"valid file", "user/v1/user.proto", false},
		{"valid nested", "proto/service/v1/service.proto", false},
		{"valid root", "test.proto", false},
		{"path traversal", "../etc/passwd.proto", true},
		{"path traversal nested", "foo/../../bar.proto", true},
		{"absolute path", "/etc/passwd.proto", true},
		{"wrong extension", "user/v1/user.txt", true},
		{"no extension", "user/v1/user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProtoFile(tt.file)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBaseGenerator_Validate(t *testing.T) {
	g := &baseGenerator{sdk: SdkGoProtobuf}

	tests := []struct {
		name    string
		input   GeneratorInput
		wantErr string
	}{
		{
			name:    "no proto files",
			input:   GeneratorInput{RepoPath: "/repo", OutputPath: "/out", ProtoFiles: []string{}},
			wantErr: "at least one proto file is required",
		},
		{
			name:    "invalid proto file",
			input:   GeneratorInput{RepoPath: "/repo", OutputPath: "/out", ProtoFiles: []string{"../bad.proto"}},
			wantErr: "invalid proto file",
		},
		{
			name:    "valid input",
			input:   GeneratorInput{RepoPath: "/repo", OutputPath: "/out", ProtoFiles: []string{"user/v1/user.proto"}},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.Validate(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateGoPackageMapping(t *testing.T) {
	tests := []struct {
		protoFile string
		optPrefix string
		expected  string
	}{
		{"user/v1/user.proto", "go_opt", "--go_opt=Muser/v1/user.proto=./user/v1"},
		{"test.proto", "go_opt", "--go_opt=Mtest.proto=./"},
		{"proto/service.proto", "connect-go_opt", "--connect-go_opt=Mproto/service.proto=./proto"},
	}

	for _, tt := range tests {
		t.Run(tt.protoFile, func(t *testing.T) {
			result := generateGoPackageMapping(tt.protoFile, tt.optPrefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoProtobufGenerator_Generate(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewGoProtobufGenerator(mockRunner)

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"user/v1/user.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Equal(t, "/tmp/output", output.OutputPath)
	assert.Equal(t, 1, output.FilesCount)

	assert.Len(t, mockRunner.Calls, 1)
	call := mockRunner.Calls[0]
	assert.Equal(t, "protoc", call.Name)
	assert.Equal(t, "/tmp/repo", call.WorkDir)
	assert.Contains(t, call.Args, "--proto_path=/tmp/repo")
	assert.Contains(t, call.Args, "--go_out=/tmp/output")
	assert.Contains(t, call.Args, "--go_opt=paths=source_relative")
	assert.Contains(t, call.Args, "--go_opt=Muser/v1/user.proto=./user/v1")
	assert.Contains(t, call.Args, "user/v1/user.proto")
}

func TestGoConnectRpcGenerator_Generate(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewGoConnectRpcGenerator(mockRunner)

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"service/v1/api.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Len(t, mockRunner.Calls, 1)
	call := mockRunner.Calls[0]
	assert.Equal(t, "protoc", call.Name)
	assert.Contains(t, call.Args, "--connect-go_out=/tmp/output")
	assert.Contains(t, call.Args, "--connect-go_opt=paths=source_relative")
	assert.Contains(t, call.Args, "--connect-go_opt=Mservice/v1/api.proto=./service/v1")
}

func TestGoGrpcGenerator_Generate(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewGoGrpcGenerator(mockRunner)

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"api.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Len(t, mockRunner.Calls, 1)
	call := mockRunner.Calls[0]
	assert.Equal(t, "protoc", call.Name)
	assert.Contains(t, call.Args, "--go-grpc_out=/tmp/output")
	assert.Contains(t, call.Args, "--go-grpc_opt=paths=source_relative")
	assert.Contains(t, call.Args, "--go-grpc_opt=Mapi.proto=./")
}

func TestRustProtobufGenerator_Generate(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewRustProtobufGenerator(mockRunner)

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"user/v1/user.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Equal(t, "/tmp/output", output.OutputPath)
	assert.Equal(t, 1, output.FilesCount)

	assert.Len(t, mockRunner.Calls, 1)
	call := mockRunner.Calls[0]
	assert.Equal(t, "protoc", call.Name)
	assert.Equal(t, "/tmp/repo", call.WorkDir)
	assert.Contains(t, call.Args, "--proto_path=/tmp/repo")
	assert.Contains(t, call.Args, "--prost_out=/tmp/output")
	assert.Contains(t, call.Args, "user/v1/user.proto")
}

func TestRustGrpcGenerator_Generate(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewRustGrpcGenerator(mockRunner)

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"user/v1/user.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Equal(t, "/tmp/output", output.OutputPath)
	assert.Equal(t, 1, output.FilesCount)

	assert.Len(t, mockRunner.Calls, 1)
	call := mockRunner.Calls[0]
	assert.Equal(t, "protoc", call.Name)
	assert.Equal(t, "/tmp/repo", call.WorkDir)
	assert.Contains(t, call.Args, "--proto_path=/tmp/repo")
	assert.Contains(t, call.Args, "--prost_out=/tmp/output")
	assert.Contains(t, call.Args, "--tonic_out=/tmp/output")
	assert.Contains(t, call.Args, "user/v1/user.proto")
}

func TestGenerator_GenerateError(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	mockRunner.RunFunc = func(ctx context.Context, name string, args []string, workDir string) ([]byte, error) {
		return nil, errors.New("protoc failed: exit status 1")
	}

	g := NewGoProtobufGenerator(mockRunner)

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"test.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "protoc failed")
}

func TestGenerator_ValidationError(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	g := NewGoProtobufGenerator(mockRunner)

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"test.txt", "test.cpp"},
	}

	output, err := g.Generate(t.Context(), input)

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Empty(t, mockRunner.Calls)
}

func TestJsGenerators_Generate(t *testing.T) {
	ctx := context.Background()
	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"api.proto"},
	}

	tests := []struct {
		name        string
		generator   Generator
		expectedArg string
	}{
		{
			name:        "JsBufbuildEs",
			generator:   NewJsBufbuildEsGenerator(NewMockCommandRunner()),
			expectedArg: "--es_out=/tmp/output",
		},
		{
			name:        "JsProtobuf",
			generator:   NewJsProtobufGenerator(NewMockCommandRunner()),
			expectedArg: "--js_out=import_style=commonjs,binary:/tmp/output",
		},
		{
			name:        "JsConnectRpc",
			generator:   NewJsConnectRpcGenerator(NewMockCommandRunner()),
			expectedArg: "--connect-es_out=/tmp/output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.generator.Generate(ctx, input)
			require.NoError(t, err)
			assert.NotNil(t, output)
			assert.Equal(t, "/tmp/output", output.OutputPath)
		})
	}
}

func TestRemoveScalarValueTypesSection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "removes scalar value types section with ## header",
			input: `# API Documentation

## Messages

### User
User message description

## Scalar Value Types

| .proto Type | Notes | C++ Type | Java Type |
|-------------|-------|----------|-----------|
| double |  | double | double |

## Services

### UserService
Service description
`,
			expected: `# API Documentation

## Messages

### User
User message description

## Services

### UserService
Service description
`,
		},
		{
			name: "removes scalar value types section with # header",
			input: `# API Documentation

# Scalar Value Types

| .proto Type | Notes |
|-------------|-------|
| double |  |

## Messages
`,
			expected: `# API Documentation

## Messages
`,
		},
		{
			name: "removes scalar value types section with ### header",
			input: `# API Documentation

### Scalar Value Types

Content here

## Next Section
`,
			expected: `# API Documentation

## Next Section
`,
		},
		{
			name: "case insensitive matching",
			input: `# API Documentation

## Scalar value types

| Type | Notes |
|------|-------|
| double |  |

## Messages
`,
			expected: `# API Documentation

## Messages
`,
		},
		{
			name: "removes scalar value type (singular)",
			input: `# API Documentation

## Scalar Value Type

| Type | Notes |
|------|-------|
| double |  |

## Messages
`,
			expected: `# API Documentation

## Messages
`,
		},
		{
			name: "no scalar section - unchanged",
			input: `# API Documentation

## Messages

### User
User message description

## Services
`,
			expected: `# API Documentation

## Messages

### User
User message description

## Services
`,
		},
		{
			name: "scalar section at the end",
			input: `# API Documentation

## Messages

## Scalar Value Types

| Type | Notes |
|------|-------|
| double |  |
`,
			expected: `# API Documentation

## Messages
`,
		},
		{
			name: "preserves newline at end",
			input: `# API Documentation

## Scalar Value Types

| Type | Notes |
|------|-------|
| double |  |

## Messages
`,
			expected: `# API Documentation

## Messages
`,
		},
		{
			name: "multiple sections with scalar in middle",
			input: `# API Documentation

## Section 1
Content 1

## Scalar Value Types

| Type | Notes |
|------|-------|
| double |  |

## Section 2
Content 2

## Section 3
Content 3
`,
			expected: `# API Documentation

## Section 1
Content 1

## Section 2
Content 2

## Section 3
Content 3
`,
		},
		{
			name: "scalar section with empty lines",
			input: `# API Documentation

## Scalar Value Types


| Type | Notes |
|------|-------|
| double |  |


## Messages
`,
			expected: `# API Documentation

## Messages
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "index.md")

			err := os.WriteFile(filePath, []byte(tt.input), 0o644)
			require.NoError(t, err)

			err = removeScalarValueTypesSection(filePath)
			require.NoError(t, err)

			content, err := os.ReadFile(filePath)
			require.NoError(t, err)

			actual := string(content)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestDocumentationGenerator_Generate(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewDocumentationGenerator(mockRunner)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "docs")

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: outputPath,
		ProtoFiles: []string{"user/v1/user.proto"},
	}

	err := os.MkdirAll(outputPath, 0o750)
	require.NoError(t, err)

	indexMdPath := filepath.Join(outputPath, "index.md")
	markdownContent := `# API Documentation

## Messages

### User
User message description

## Scalar Value Types

| .proto Type | Notes | C++ Type |
|-------------|-------|----------|
| double |  | double |

## Services

### UserService
Service description
`
	err = os.WriteFile(indexMdPath, []byte(markdownContent), 0o644)
	require.NoError(t, err)

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Len(t, mockRunner.Calls, 1)
	call := mockRunner.Calls[0]
	assert.Equal(t, "protoc", call.Name)
	assert.Equal(t, "/tmp/repo", call.WorkDir)
	assert.Contains(t, call.Args, "--doc_out="+outputPath)
	assert.Contains(t, call.Args, "--doc_opt="+filepath.Join(outputPath, "proto-doc-template.mustache")+",index.md")

	content, err := os.ReadFile(indexMdPath)
	require.NoError(t, err)
	result := string(content)
	assert.NotContains(t, result, "Scalar Value Types")
	assert.Contains(t, result, "API Documentation")
	assert.Contains(t, result, "Messages")
	assert.Contains(t, result, "Services")
}

func TestDocumentationGenerator_Generate_NoScalarSection(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewDocumentationGenerator(mockRunner)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "docs")

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: outputPath,
		ProtoFiles: []string{"user/v1/user.proto"},
	}

	err := os.MkdirAll(outputPath, 0o750)
	require.NoError(t, err)

	indexMdPath := filepath.Join(outputPath, "index.md")
	markdownContent := `# API Documentation

## Messages

### User
User message description
`
	err = os.WriteFile(indexMdPath, []byte(markdownContent), 0o644)
	require.NoError(t, err)

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	content, err := os.ReadFile(indexMdPath)
	require.NoError(t, err)
	result := string(content)
	assert.Equal(t, markdownContent, result)
}

func TestDocumentationGenerator_Generate_ProtocError(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	mockRunner.RunFunc = func(ctx context.Context, name string, args []string, workDir string) ([]byte, error) {
		return nil, errors.New("protoc failed")
	}

	g := NewDocumentationGenerator(mockRunner)

	input := GeneratorInput{
		RepoPath:   "/tmp/repo",
		OutputPath: "/tmp/output",
		ProtoFiles: []string{"user/v1/user.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "protoc failed")
}

func TestDocumentationGenerator_Generate_WithBufYaml(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewDocumentationGenerator(mockRunner)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(outputPath, 0o750))

	// Create buf.yaml
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "buf.yaml"), []byte("version: v2\n"), 0o644))

	input := GeneratorInput{
		RepoPath:   tmpDir,
		OutputPath: outputPath,
		ProtoFiles: []string{"user.proto"},
	}

	indexMdPath := filepath.Join(outputPath, "index.md")
	markdownContent := `# API Documentation
## Scalar Value Types
`
	err := os.WriteFile(indexMdPath, []byte(markdownContent), 0o644)
	require.NoError(t, err)

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Verify that mockRunner was called for buf mod update and buf generate
	require.Len(t, mockRunner.Calls, 2)
	assert.Equal(t, "buf", mockRunner.Calls[0].Name)
	assert.Contains(t, mockRunner.Calls[0].Args, "mod")
	assert.Contains(t, mockRunner.Calls[0].Args, "update")

	assert.Equal(t, "buf", mockRunner.Calls[1].Name)
	assert.Contains(t, mockRunner.Calls[1].Args, "generate")

	// Ensure scalar section was removed
	content, err := os.ReadFile(indexMdPath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "Scalar Value Types")
}

func TestDocumentationGenerator_Generate_WithExistingBufGenYaml(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewDocumentationGenerator(mockRunner)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(outputPath, 0o750))

	// Create buf.yaml and buf.gen.yaml
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "buf.yaml"), []byte("version: v2\n"), 0o644))
	bufGenYamlPath := filepath.Join(tmpDir, "buf.gen.yaml")
	require.NoError(t, os.WriteFile(bufGenYamlPath, []byte("version: v2\n"), 0o644))

	input := GeneratorInput{
		RepoPath:   tmpDir,
		OutputPath: outputPath,
		ProtoFiles: []string{"user.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Verify that mockRunner was called
	require.Len(t, mockRunner.Calls, 2)
	assert.Equal(t, "buf", mockRunner.Calls[0].Name)
	assert.Equal(t, "buf", mockRunner.Calls[1].Name)

	// Ensure buf.gen.yaml was NOT deleted
	_, err = os.Stat(bufGenYamlPath)
	assert.NoError(t, err, "existing buf.gen.yaml should not be deleted")
}

func initBareRepoWithFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	bareDir := filepath.Join(tmpDir, "bare")

	require.NoError(t, os.MkdirAll(workDir, 0o750))

	cmds := [][]string{
		{"git", "init", workDir},
		{"git", "-C", workDir, "config", "user.email", "test@test.com"},
		{"git", "-C", workDir, "config", "user.name", "Test"},
		{"git", "-C", workDir, "config", "commit.gpgsign", "false"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "command %v failed: %s", args, string(out))
	}

	for path, content := range files {
		fullPath := filepath.Join(workDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o750))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
	}

	commitCmds := [][]string{
		{"git", "-C", workDir, "add", "."},
		{"git", "-C", workDir, "commit", "-m", "initial"},
		{"git", "clone", "--bare", workDir, bareDir},
	}

	for _, args := range commitCmds {
		cmd := exec.Command(args[0], args[1:]...)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "command %v failed: %s", args, string(out))
	}

	return bareDir
}

func TestFindProtoFilesInBareRepo(t *testing.T) {
	t.Run("finds proto files in bare repo", func(t *testing.T) {
		bareDir := initBareRepoWithFiles(t, map[string]string{
			"proto/organization/v1/organization.proto": "syntax = \"proto3\";",
			"proto/user/v1/user.proto":                 "syntax = \"proto3\";",
			"README.md":                                "# Test",
		})

		protoFiles, err := FindProtoFilesInBareRepo(bareDir, "HEAD")
		require.NoError(t, err)
		assert.Len(t, protoFiles, 2)
		assert.Contains(t, protoFiles, "proto/organization/v1/organization.proto")
		assert.Contains(t, protoFiles, "proto/user/v1/user.proto")
	})

	t.Run("returns empty when bare repo has no proto files", func(t *testing.T) {
		bareDir := initBareRepoWithFiles(t, map[string]string{
			"README.md": "# Test",
			"main.go":   "package main",
		})

		protoFiles, err := FindProtoFilesInBareRepo(bareDir, "HEAD")
		require.NoError(t, err)
		assert.Empty(t, protoFiles)
	})

	t.Run("works with specific commit hash", func(t *testing.T) {
		bareDir := initBareRepoWithFiles(t, map[string]string{
			"proto/v1/api.proto": "syntax = \"proto3\";",
		})

		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = bareDir
		out, err := cmd.Output()
		require.NoError(t, err)
		commitHash := string(out[:len(out)-1])

		protoFiles, err := FindProtoFilesInBareRepo(bareDir, commitHash)
		require.NoError(t, err)
		assert.Len(t, protoFiles, 1)
		assert.Contains(t, protoFiles, "proto/v1/api.proto")
	})

	t.Run("defaults to HEAD when commit hash is empty", func(t *testing.T) {
		bareDir := initBareRepoWithFiles(t, map[string]string{
			"service.proto": "syntax = \"proto3\";",
		})

		protoFiles, err := FindProtoFilesInBareRepo(bareDir, "")
		require.NoError(t, err)
		assert.Len(t, protoFiles, 1)
		assert.Contains(t, protoFiles, "service.proto")
	})

	t.Run("error on invalid repo path", func(t *testing.T) {
		_, err := FindProtoFilesInBareRepo("/nonexistent/repo", "HEAD")
		require.Error(t, err)
	})
}
