GO111MODULE=on

install:
	@go get

gen_example: install
	@protoc --gohttp_out=. ./examples/*.proto

test: gen_example
	@go test ./...

run_examples:
	@go run examples/main.go examples/greeter.pb.go examples/greeter.http.go

curl_google_api:
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > testdata/google/api/http.proto

gen_test_grpc_go_files:
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/auth/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/hellostreamingworld/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/helloworld/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/httprule/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/routeguide/*.proto
