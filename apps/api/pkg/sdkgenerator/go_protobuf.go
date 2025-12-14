package sdkgenerator

import (
	"context"
	"path/filepath"
)

type GoProtobufGenerator struct {
	baseGenerator
}

func NewGoProtobufGenerator(runner CommandRunner) *GoProtobufGenerator {
	return &GoProtobufGenerator{
		baseGenerator: baseGenerator{
			sdk:    SdkGoProtobuf,
			runner: runner,
		},
	}
}

func (g *GoProtobufGenerator) Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	if err := g.Validate(input); err != nil {
		return nil, err
	}

	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--go_out=" + filepath.Clean(input.OutputPath),
		"--go_opt=paths=source_relative",
	}

	for _, protoFile := range input.ProtoFiles {
		args = append(args, generateGoPackageMapping(protoFile, "go_opt"))
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
