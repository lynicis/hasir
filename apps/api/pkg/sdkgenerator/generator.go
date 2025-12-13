package sdkgenerator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type SDK string

const (
	SdkGoProtobuf   SDK = "GO_PROTOBUF"
	SdkGoConnectRpc SDK = "GO_CONNECTRPC"
	SdkGoGrpc       SDK = "GO_GRPC"
	SdkJsBufbuildEs SDK = "JS_BUFBUILD_ES"
	SdkJsProtobuf   SDK = "JS_PROTOBUF"
	SdkJsConnectrpc SDK = "JS_CONNECTRPC"
)

var sdkDirNames = map[SDK]string{
	SdkGoProtobuf:   "go-protobuf",
	SdkGoConnectRpc: "go-connectrpc",
	SdkGoGrpc:       "go-grpc",
	SdkJsBufbuildEs: "js-bufbuild-es",
	SdkJsProtobuf:   "js-protobuf",
	SdkJsConnectrpc: "js-connectrpc",
}

func (s SDK) DirName() string {
	if dirName, ok := sdkDirNames[s]; ok {
		return dirName
	}

	return "unknown"
}

func (s SDK) IsGo() bool {
	return s == SdkGoProtobuf || s == SdkGoConnectRpc || s == SdkGoGrpc
}

func (s SDK) IsJs() bool {
	return s == SdkJsBufbuildEs || s == SdkJsProtobuf || s == SdkJsConnectrpc
}

type GeneratorInput struct {
	RepoPath   string
	OutputPath string
	ProtoFiles []string
}

type GeneratorOutput struct {
	OutputPath string
	FilesCount int
}

type Generator interface {
	Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error)
	SDK() SDK
	Validate(input GeneratorInput) error
}

type CommandRunner interface {
	Run(ctx context.Context, name string, args []string, workDir string) ([]byte, error)
}

type baseGenerator struct {
	sdk    SDK
	runner CommandRunner
}

func (g *baseGenerator) SDK() SDK {
	return g.sdk
}

func (g *baseGenerator) Validate(input GeneratorInput) error {
	if input.RepoPath == "" {
		return errors.New("repo path is required")
	}
	if input.OutputPath == "" {
		return errors.New("output path is required")
	}
	if len(input.ProtoFiles) == 0 {
		return errors.New("at least one proto file is required")
	}

	for _, file := range input.ProtoFiles {
		if err := validateProtoFile(file); err != nil {
			return fmt.Errorf("invalid proto file %q: %w", file, err)
		}
	}

	return nil
}

func validateProtoFile(file string) error {
	cleanFile := filepath.Clean(file)
	if strings.Contains(cleanFile, "..") {
		return errors.New("path traversal detected")
	}

	if filepath.IsAbs(cleanFile) {
		return errors.New("absolute paths not allowed")
	}

	if !strings.HasSuffix(cleanFile, ".proto") {
		return errors.New("file must have .proto extension")
	}

	return nil
}

func generateGoPackageMapping(protoFile, optPrefix string) string {
	dir := filepath.Dir(protoFile)
	goPackage := "./"
	if dir != "." {
		goPackage = "./" + dir
	}

	return fmt.Sprintf("--%s=M%s=%s", optPrefix, protoFile, goPackage)
}
