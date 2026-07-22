package sdkgenerator

func NewBufRustProtobufGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkRustProtobuf, "rust-protobuf", runner, []BufGenPlugin{
		{Remote: "buf.build/community/neoeinstein-prost"},
	}, nil)
}

func NewBufRustGrpcGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkRustGrpc, "rust-grpc", runner, []BufGenPlugin{
		{Remote: "buf.build/community/neoeinstein-prost"},
		{Remote: "buf.build/community/neoeinstein-tonic"},
	}, nil)
}
