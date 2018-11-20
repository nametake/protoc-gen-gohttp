ensure:
	@dep ensure -vendor-only

install: ensure
	@go install

gen_example: install
	@protoc --gohttp_out=. ./examples/*.proto

test: gen_example
	@go test ./...
