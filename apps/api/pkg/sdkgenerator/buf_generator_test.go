package sdkgenerator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBufGenerator_Generate(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()

	plugins := []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/go", Out: ".", Opt: "paths=source_relative"},
	}

	g := NewBufGenerator(SdkGoProtobuf, "go-protobuf", mockRunner, plugins, nil)

	tmpDir := t.TempDir()
	protoContent := `syntax = "proto3";
package user.v1;
import "google/protobuf/timestamp.proto";
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "user.proto"), []byte(protoContent), 0o644))

	outputDir := t.TempDir()

	input := GeneratorInput{
		RepoPath:   tmpDir,
		OutputPath: outputDir,
		ProtoFiles: []string{"user.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Equal(t, outputDir, output.OutputPath)
	assert.Equal(t, 1, output.FilesCount)

	require.Len(t, mockRunner.Calls, 2)

	modUpdateCall := mockRunner.Calls[0]
	assert.Equal(t, "buf", modUpdateCall.Name)
	assert.Contains(t, modUpdateCall.Args, "mod")
	assert.Contains(t, modUpdateCall.Args, "update")

	generateCall := mockRunner.Calls[1]
	assert.Equal(t, "buf", generateCall.Name)
	assert.Contains(t, generateCall.Args, "generate")

	bufYaml, err := os.ReadFile(filepath.Join(tmpDir, "buf.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(bufYaml), "buf.build/protocolbuffers/wellknowntypes")

	bufGenYaml, err := os.ReadFile(filepath.Join(tmpDir, "buf.gen.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(bufGenYaml), "buf.build/protocolbuffers/go")
}

func TestBufGenerator_Generate_BufGenerateError(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	callCount := 0
	mockRunner.RunFunc = func(ctx context.Context, name string, args []string, workDir string) ([]byte, error) {
		callCount++
		if callCount == 2 {
			return nil, errors.New("buf generate failed: exit status 1")
		}
		return nil, nil
	}

	plugins := []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/go", Out: ".", Opt: "paths=source_relative"},
	}

	g := NewBufGenerator(SdkGoProtobuf, "go-protobuf", mockRunner, plugins, nil)

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(`syntax = "proto3";`), 0o644))

	input := GeneratorInput{
		RepoPath:   tmpDir,
		OutputPath: t.TempDir(),
		ProtoFiles: []string{"test.proto"},
	}

	output, err := g.Generate(ctx, input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "buf generate failed")
}

func TestBufGenerator_SDK(t *testing.T) {
	g := NewBufGenerator(SdkGoProtobuf, "go-protobuf", NewMockCommandRunner(), nil, nil)
	assert.Equal(t, SdkGoProtobuf, g.SDK())
	assert.Equal(t, "go-protobuf", g.DirName())
}

func TestBufGenerator_Validate(t *testing.T) {
	g := NewBufGenerator(SdkGoProtobuf, "go-protobuf", NewMockCommandRunner(), nil, nil)

	err := g.Validate(GeneratorInput{ProtoFiles: []string{}})
	assert.Error(t, err)

	err = g.Validate(GeneratorInput{ProtoFiles: []string{"test.proto"}})
	assert.NoError(t, err)
}

func TestBufGenerator_CustomBsrModules(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()

	customModules := map[string]string{
		"myorg/": "buf.build/myorg/mymodule",
	}

	plugins := []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/go", Out: ".", Opt: "paths=source_relative"},
	}

	g := NewBufGenerator(SdkGoProtobuf, "go-protobuf", mockRunner, plugins, customModules)

	tmpDir := t.TempDir()
	protoContent := `syntax = "proto3";
import "myorg/shared/types.proto";
import "google/protobuf/timestamp.proto";
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0o644))

	input := GeneratorInput{
		RepoPath:   tmpDir,
		OutputPath: t.TempDir(),
		ProtoFiles: []string{"test.proto"},
	}

	_, err := g.Generate(ctx, input)
	require.NoError(t, err)

	bufYaml, err := os.ReadFile(filepath.Join(tmpDir, "buf.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(bufYaml), "buf.build/myorg/mymodule")
	assert.Contains(t, string(bufYaml), "buf.build/protocolbuffers/wellknowntypes")
}

func TestBufGoProtobufGenerator(t *testing.T) {
	g := NewBufGoProtobufGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkGoProtobuf, g.SDK())
	assert.Equal(t, "go-protobuf", g.DirName())
}

func TestBufGoConnectRpcGenerator(t *testing.T) {
	g := NewBufGoConnectRpcGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkGoConnectRpc, g.SDK())
	assert.Equal(t, "go-connectrpc", g.DirName())
}

func TestBufGoGrpcGenerator(t *testing.T) {
	g := NewBufGoGrpcGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkGoGrpc, g.SDK())
	assert.Equal(t, "go-grpc", g.DirName())
}

func TestBufJsBufbuildEsGenerator(t *testing.T) {
	g := NewBufJsBufbuildEsGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkJsBufbuildEs, g.SDK())
	assert.Equal(t, "js-bufbuild-es", g.DirName())
}

func TestBufJsProtobufGenerator(t *testing.T) {
	g := NewBufJsProtobufGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkJsProtobuf, g.SDK())
	assert.Equal(t, "js-protobuf", g.DirName())
}

func TestBufJsConnectRpcGenerator(t *testing.T) {
	g := NewBufJsConnectRpcGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkJsConnectrpc, g.SDK())
	assert.Equal(t, "js-connectrpc", g.DirName())
}

func TestBufRustProtobufGenerator(t *testing.T) {
	g := NewBufRustProtobufGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkRustProtobuf, g.SDK())
	assert.Equal(t, "rust-protobuf", g.DirName())
}

func TestBufRustGrpcGenerator(t *testing.T) {
	g := NewBufRustGrpcGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkRustGrpc, g.SDK())
	assert.Equal(t, "rust-grpc", g.DirName())
}

func TestBufJavaProtobufGenerator(t *testing.T) {
	g := NewBufJavaProtobufGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkJavaProtobuf, g.SDK())
	assert.Equal(t, "java-protobuf", g.DirName())
}

func TestBufJavaGrpcGenerator(t *testing.T) {
	g := NewBufJavaGrpcGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkJavaGrpc, g.SDK())
	assert.Equal(t, "java-grpc", g.DirName())
}

func TestBufCsharpProtobufGenerator(t *testing.T) {
	g := NewBufCsharpProtobufGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkCsharpProtobuf, g.SDK())
	assert.Equal(t, "csharp-protobuf", g.DirName())
}

func TestBufCsharpGrpcGenerator(t *testing.T) {
	g := NewBufCsharpGrpcGenerator(NewMockCommandRunner())
	assert.Equal(t, SdkCsharpGrpc, g.SDK())
	assert.Equal(t, "csharp-grpc", g.DirName())
}
