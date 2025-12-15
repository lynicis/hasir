package sdkgenerator

import (
	"path/filepath"
)

type GoProtobufGenerator struct {
	*protocGenerator
}

func NewGoProtobufGenerator(runner CommandRunner) *GoProtobufGenerator {
	return &GoProtobufGenerator{
		protocGenerator: newProtocGenerator(SdkGoProtobuf, "go-protobuf", runner, buildGoProtobufArgs),
	}
}

func buildGoProtobufArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--go_out=" + filepath.Clean(input.OutputPath),
		"--go_opt=paths=source_relative",
	}

	for _, protoFile := range input.ProtoFiles {
		args = append(args, generateGoPackageMapping(protoFile, "go_opt"))
	}

	args = append(args, input.ProtoFiles...)
	return args
}
