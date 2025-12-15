package sdkgenerator

import (
	"path/filepath"
)

type GoConnectRpcGenerator struct {
	*protocGenerator
}

func NewGoConnectRpcGenerator(runner CommandRunner) *GoConnectRpcGenerator {
	return &GoConnectRpcGenerator{
		protocGenerator: newProtocGenerator(SdkGoConnectRpc, "go-connectrpc", runner, buildGoConnectRpcArgs),
	}
}

func buildGoConnectRpcArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--go_out=" + filepath.Clean(input.OutputPath),
		"--go_opt=paths=source_relative",
		"--connect-go_out=" + filepath.Clean(input.OutputPath),
		"--connect-go_opt=paths=source_relative",
	}

	for _, protoFile := range input.ProtoFiles {
		args = append(args, generateGoPackageMapping(protoFile, "go_opt"))
		args = append(args, generateGoPackageMapping(protoFile, "connect-go_opt"))
	}

	args = append(args, input.ProtoFiles...)
	return args
}
