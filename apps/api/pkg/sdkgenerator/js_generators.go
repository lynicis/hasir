package sdkgenerator

import (
	"context"
	"path/filepath"
)

type JsBufbuildEsGenerator struct {
	baseGenerator
}

func NewJsBufbuildEsGenerator(runner CommandRunner) *JsBufbuildEsGenerator {
	return &JsBufbuildEsGenerator{
		baseGenerator: baseGenerator{
			sdk:    SdkJsBufbuildEs,
			runner: runner,
		},
	}
}

func (g *JsBufbuildEsGenerator) Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	if err := g.Validate(input); err != nil {
		return nil, err
	}

	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--es_out=" + filepath.Clean(input.OutputPath),
		"--es_opt=target=ts",
	}
	args = append(args, input.ProtoFiles...)

	if _, err := g.runner.Run(ctx, "protoc", args, input.RepoPath); err != nil {
		return nil, err
	}

	return &GeneratorOutput{
		OutputPath: input.OutputPath,
		FilesCount: len(input.ProtoFiles),
	}, nil
}

type JsProtobufGenerator struct {
	baseGenerator
}

func NewJsProtobufGenerator(runner CommandRunner) *JsProtobufGenerator {
	return &JsProtobufGenerator{
		baseGenerator: baseGenerator{
			sdk:    SdkJsProtobuf,
			runner: runner,
		},
	}
}

func (g *JsProtobufGenerator) Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	if err := g.Validate(input); err != nil {
		return nil, err
	}

	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--js_out=import_style=commonjs,binary:" + filepath.Clean(input.OutputPath),
	}
	args = append(args, input.ProtoFiles...)

	if _, err := g.runner.Run(ctx, "protoc", args, input.RepoPath); err != nil {
		return nil, err
	}

	return &GeneratorOutput{
		OutputPath: input.OutputPath,
		FilesCount: len(input.ProtoFiles),
	}, nil
}

type JsConnectRpcGenerator struct {
	baseGenerator
}

func NewJsConnectRpcGenerator(runner CommandRunner) *JsConnectRpcGenerator {
	return &JsConnectRpcGenerator{
		baseGenerator: baseGenerator{
			sdk:    SdkJsConnectrpc,
			runner: runner,
		},
	}
}

func (g *JsConnectRpcGenerator) Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	if err := g.Validate(input); err != nil {
		return nil, err
	}

	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--es_out=" + filepath.Clean(input.OutputPath),
		"--es_opt=target=ts",
		"--connect-es_out=" + filepath.Clean(input.OutputPath),
		"--connect-es_opt=target=ts",
	}
	args = append(args, input.ProtoFiles...)

	if _, err := g.runner.Run(ctx, "protoc", args, input.RepoPath); err != nil {
		return nil, err
	}

	return &GeneratorOutput{
		OutputPath: input.OutputPath,
		FilesCount: len(input.ProtoFiles),
	}, nil
}
