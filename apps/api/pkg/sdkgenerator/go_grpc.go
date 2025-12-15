package sdkgenerator

import (
	"path/filepath"
)

type GoGrpcGenerator struct {
	*protocGenerator
}

func NewGoGrpcGenerator(runner CommandRunner) *GoGrpcGenerator {
	return &GoGrpcGenerator{
		protocGenerator: newProtocGenerator(SdkGoGrpc, "go-grpc", runner, buildGoGrpcArgs),
	}
}

func buildGoGrpcArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--go_out=" + filepath.Clean(input.OutputPath),
		"--go_opt=paths=source_relative",
		"--go-grpc_out=" + filepath.Clean(input.OutputPath),
		"--go-grpc_opt=paths=source_relative",
	}

	for _, protoFile := range input.ProtoFiles {
		args = append(args, generateGoPackageMapping(protoFile, "go_opt"))
		args = append(args, generateGoPackageMapping(protoFile, "go-grpc_opt"))
	}

	args = append(args, input.ProtoFiles...)
	return args
}
