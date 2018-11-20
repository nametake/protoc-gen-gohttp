ensure:
	@dep ensure

install: ensure
	@go install

gen_example: install
	@protoc --go_out=plugins=grpc:. --gohttp_out=. ./examples/*.proto

test: gen_example
	@go test ./...
