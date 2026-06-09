package sdkgenerator

func NewBufGoProtobufGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkGoProtobuf, "go-protobuf", runner, []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/go", Opt: "paths=source_relative"},
	}, nil)
}

func NewBufGoConnectRpcGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkGoConnectRpc, "go-connectrpc", runner, []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/go", Opt: "paths=source_relative"},
		{Remote: "buf.build/connectrpc/go", Opt: "paths=source_relative"},
	}, nil)
}

func NewBufGoGrpcGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkGoGrpc, "go-grpc", runner, []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/go", Opt: "paths=source_relative"},
		{Remote: "buf.build/grpc/go", Opt: "paths=source_relative"},
	}, nil)
}
