package sdkgenerator

import "path/filepath"

type JavaGrpcGenerator struct {
	*protocGenerator
}

func NewJavaGrpcGenerator(runner CommandRunner) *JavaGrpcGenerator {
	return &JavaGrpcGenerator{
		protocGenerator: newProtocGenerator(SdkJavaGrpc, "java-grpc", runner, buildJavaGrpcArgs),
	}
}

func buildJavaGrpcArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--java_out=" + filepath.Clean(input.OutputPath),
		"--grpc-java_out=" + filepath.Clean(input.OutputPath),
	}
	args = append(args, input.ProtoFiles...)
	return args
}
