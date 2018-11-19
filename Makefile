gen_example:
	@protoc --go_out=plugins=grpc:. --gohttp_out=. ./examples/helloworld.proto
