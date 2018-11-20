ensure:
	@dep ensure -vendor-only

install: ensure
	@go install

gen_example: install
	@protoc --gohttp_out=. ./examples/*.proto

test: gen_example
	@go test ./...

run_examples:
	go run examples/main.go examples/greeter.pb.go examples/greeter.http.go
