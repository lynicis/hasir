package sdkgenerator

import (
	"path/filepath"
)

type JsBufbuildEsGenerator struct {
	*protocGenerator
}

func NewJsBufbuildEsGenerator(runner CommandRunner) *JsBufbuildEsGenerator {
	return &JsBufbuildEsGenerator{
		protocGenerator: newProtocGenerator(SdkJsBufbuildEs, "js-bufbuild-es", runner, buildJsBufbuildEsArgs),
	}
}

func buildJsBufbuildEsArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--es_out=" + filepath.Clean(input.OutputPath),
		"--es_opt=target=ts",
	}
	args = append(args, input.ProtoFiles...)
	return args
}

type JsProtobufGenerator struct {
	*protocGenerator
}

func NewJsProtobufGenerator(runner CommandRunner) *JsProtobufGenerator {
	return &JsProtobufGenerator{
		protocGenerator: newProtocGenerator(SdkJsProtobuf, "js-protobuf", runner, buildJsProtobufArgs),
	}
}

func buildJsProtobufArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--js_out=import_style=commonjs,binary:" + filepath.Clean(input.OutputPath),
	}
	args = append(args, input.ProtoFiles...)
	return args
}

type JsConnectRpcGenerator struct {
	*protocGenerator
}

func NewJsConnectRpcGenerator(runner CommandRunner) *JsConnectRpcGenerator {
	return &JsConnectRpcGenerator{
		protocGenerator: newProtocGenerator(SdkJsConnectrpc, "js-connectrpc", runner, buildJsConnectRpcArgs),
	}
}

func buildJsConnectRpcArgs(input GeneratorInput) []string {
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--es_out=" + filepath.Clean(input.OutputPath),
		"--es_opt=target=ts",
		"--connect-es_out=" + filepath.Clean(input.OutputPath),
		"--connect-es_opt=target=ts",
	}
	args = append(args, input.ProtoFiles...)
	return args
}
