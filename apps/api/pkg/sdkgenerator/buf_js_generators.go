package sdkgenerator

func NewBufJsBufbuildEsGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkJsBufbuildEs, "js-bufbuild-es", runner, []BufGenPlugin{
		{Remote: "buf.build/bufbuild/es", Opt: "target=ts"},
	}, nil)
}

func NewBufJsProtobufGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkJsProtobuf, "js-protobuf", runner, []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/js", Opts: []string{"import_style=commonjs", "binary"}},
	}, nil)
}

func NewBufJsConnectRpcGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkJsConnectrpc, "js-connectrpc", runner, []BufGenPlugin{
		{Remote: "buf.build/bufbuild/es", Opt: "target=ts"},
		{Remote: "buf.build/connectrpc/es", Opt: "target=ts"},
	}, nil)
}
