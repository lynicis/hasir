package sdkgenerator

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	DirName() string
	Validate(input GeneratorInput) error
}

type CommandRunner interface {
	Run(ctx context.Context, name string, args []string, workDir string) ([]byte, error)
}

type ProtocArgsBuilder func(input GeneratorInput) []string

type baseGenerator struct {
	sdk     SDK
	dirName string
	runner  CommandRunner
}

type protocGenerator struct {
	baseGenerator
	argsBuilder ProtocArgsBuilder
}

func (g *baseGenerator) SDK() SDK {
	return g.sdk
}

func (g *baseGenerator) DirName() string {
	return g.dirName
}

func (g *baseGenerator) Validate(input GeneratorInput) error {
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

func newProtocGenerator(sdk SDK, dirName string, runner CommandRunner, argsBuilder ProtocArgsBuilder) *protocGenerator {
	return &protocGenerator{
		baseGenerator: baseGenerator{
			sdk:     sdk,
			dirName: dirName,
			runner:  runner,
		},
		argsBuilder: argsBuilder,
	}
}

func (g *protocGenerator) Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	if err := g.Validate(input); err != nil {
		return nil, err
	}

	args := g.argsBuilder(input)

	if _, err := g.runner.Run(ctx, "protoc", args, input.RepoPath); err != nil {
		return nil, err
	}

	return &GeneratorOutput{
		OutputPath: input.OutputPath,
		FilesCount: len(input.ProtoFiles),
	}, nil
}

func GenerateFromRepo(ctx context.Context, generator Generator, repoPath, outputPath string) (*GeneratorOutput, error) {
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute output path: %w", err)
	}

	if err := os.MkdirAll(absOutputPath, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	protoFiles, err := FindProtoFiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return nil, errors.New("no proto files found in repository")
	}

	input := GeneratorInput{
		RepoPath:   repoPath,
		OutputPath: absOutputPath,
		ProtoFiles: protoFiles,
	}

	return generator.Generate(ctx, input)
}

func FindProtoFiles(repoPath string) ([]string, error) {
	var protoFiles []string
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".proto" {
			relPath, err := filepath.Rel(repoPath, path)
			if err != nil {
				return err
			}
			protoFiles = append(protoFiles, relPath)
		}

		return nil
	})

	return protoFiles, err
}
