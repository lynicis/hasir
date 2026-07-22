package sdkgenerator

func NewBufCsharpProtobufGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkCsharpProtobuf, "csharp-protobuf", runner, []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/csharp", Opt: "base_namespace="},
	}, nil)
}

func NewBufCsharpGrpcGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkCsharpGrpc, "csharp-grpc", runner, []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/csharp", Opt: "base_namespace="},
		{Remote: "buf.build/grpc/csharp", Opt: "base_namespace="},
	}, nil)
}
