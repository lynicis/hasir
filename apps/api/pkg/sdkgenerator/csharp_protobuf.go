package sdkgenerator

import "path/filepath"

type CsharpProtobufGenerator struct {
	*protocGenerator
}

func NewCsharpProtobufGenerator(runner CommandRunner) *CsharpProtobufGenerator {
	return &CsharpProtobufGenerator{
		protocGenerator: newProtocGenerator(SdkCsharpProtobuf, "csharp-protobuf", runner, buildCsharpProtobufArgs),
	}
}

func buildCsharpProtobufArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--csharp_out=" + filepath.Clean(input.OutputPath),
	}
	args = append(args, input.ProtoFiles...)
	return args
}
