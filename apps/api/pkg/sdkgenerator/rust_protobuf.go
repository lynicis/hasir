package sdkgenerator

import "path/filepath"

type RustProtobufGenerator struct {
	*protocGenerator
}

func NewRustProtobufGenerator(runner CommandRunner) *RustProtobufGenerator {
	return &RustProtobufGenerator{
		protocGenerator: newProtocGenerator(SdkRustProtobuf, "rust-protobuf", runner, buildRustProtobufArgs),
	}
}

func buildRustProtobufArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--prost_out=" + filepath.Clean(input.OutputPath),
	}
	args = append(args, input.ProtoFiles...)
	return args
}
