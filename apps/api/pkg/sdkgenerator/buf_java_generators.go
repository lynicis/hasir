package sdkgenerator

func NewBufJavaProtobufGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkJavaProtobuf, "java-protobuf", runner, []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/java"},
	}, nil)
}

func NewBufJavaGrpcGenerator(runner CommandRunner) Generator {
	return NewBufGenerator(SdkJavaGrpc, "java-grpc", runner, []BufGenPlugin{
		{Remote: "buf.build/protocolbuffers/java"},
		{Remote: "buf.build/grpc/java"},
	}, nil)
}
