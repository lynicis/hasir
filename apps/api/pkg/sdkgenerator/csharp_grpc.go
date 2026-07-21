package sdkgenerator

import "path/filepath"

type CsharpGrpcGenerator struct {
	*protocGenerator
}

func NewCsharpGrpcGenerator(runner CommandRunner) *CsharpGrpcGenerator {
	return &CsharpGrpcGenerator{
		protocGenerator: newProtocGenerator(SdkCsharpGrpc, "csharp-grpc", runner, buildCsharpGrpcArgs),
	}
}

// Deployment note: protoc auto-discovers a binary named protoc-gen-grpc on PATH.
// The upstream grpc-csharp plugin binary is named grpc_csharp_plugin (from Grpc.Tools NuGet,
// Alpine apk add grpc-plugins, or Homebrew brew install grpc). Deployment MUST symlink or
// alias BEFORE running generation, e.g. ln -sf "$(which grpc_csharp_plugin)" /usr/local/bin/protoc-gen-grpc.
func buildCsharpGrpcArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--csharp_out=" + filepath.Clean(input.OutputPath),
		"--grpc_out=" + filepath.Clean(input.OutputPath),
	}
	args = append(args, input.ProtoFiles...)
	return args
}
