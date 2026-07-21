package sdkgenerator

import "path/filepath"

type JavaProtobufGenerator struct {
	*protocGenerator
}

func NewJavaProtobufGenerator(runner CommandRunner) *JavaProtobufGenerator {
	return &JavaProtobufGenerator{
		protocGenerator: newProtocGenerator(SdkJavaProtobuf, "java-protobuf", runner, buildJavaProtobufArgs),
	}
}

func buildJavaProtobufArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--java_out=" + filepath.Clean(input.OutputPath),
	}
	args = append(args, input.ProtoFiles...)
	return args
}
