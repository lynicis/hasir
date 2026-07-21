package sdkgenerator

import "path/filepath"

type RustGrpcGenerator struct {
	*protocGenerator
}

func NewRustGrpcGenerator(runner CommandRunner) *RustGrpcGenerator {
	return &RustGrpcGenerator{
		protocGenerator: newProtocGenerator(SdkRustGrpc, "rust-grpc", runner, buildRustGrpcArgs),
	}
}

func buildRustGrpcArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--prost_out=" + filepath.Clean(input.OutputPath),
		"--tonic_out=" + filepath.Clean(input.OutputPath),
	}
	args = append(args, input.ProtoFiles...)
	return args
}
